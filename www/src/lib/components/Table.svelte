<script>
	import { onMount, onDestroy } from 'svelte';
	import Avatar from './Avatar.svelte';
	import SmartBar from './SmartBar.svelte';
	import Peer from '$lib/peer';
	import { throttle } from 'lodash';
	import { removeFromArray } from '$lib/utils';
	import AudioPlayer from '$lib/audio-player';
	import Badge from 'svelte-favicon-badge';
	import {
		audioInputDeviceID,
		videoInputDeviceID,
		ambianceID,
		ambianceSoundVolume,
		ambianceEffectsVolume,
	} from '$lib/stores';
	import Entity from '$lib/entity';
	import { MAX_FPS } from '$lib/env';
	import settingsButtonURL from '$lib/assets/settings.svg';
	import Settings from '$lib/components/Settings.svelte';
	import ambiances from '$lib/assets/ambiances';
	import { get } from 'svelte/store';
	import StreamManager from '$lib/StreamManager';
	import EntityComponent from '$lib/components/Entity.svelte';

	const faviconBackground = '#FF0000';
	const faviconColor = '#FFFFFF';

	export let tableName: string;

	const avatarsByStreamID: any = {};
	let avatars: Array<any> = [];

	const entitiesByKey: any = {};
	const entitiesByStreamID: any = {};
	let entities: Array<any> = [];

	const screenSizesByStreamID: { [key: string]: { width: number; height: number } } = {};

	const cursor = new Entity('cursor');
	let isMouseDown = false;
	let onMouseMove: svelte.JSX.MouseEventHandler<HTMLDivElement>;
	let onMouseDown: svelte.JSX.MouseEventHandler<HTMLDivElement>;
	let onMouseUp: svelte.JSX.MouseEventHandler<HTMLDivElement>;

	let showSettings = false;

	let onSmartBarEntity: any;

	const backgroundAudioPlayer = new AudioPlayer();
	backgroundAudioPlayer.setLoop(true);
	$: backgroundAudioPlayer.setVolume($ambianceSoundVolume);
	$: backgroundAudioPlayer.play(ambiances[$ambianceID]?.soundURL || '');

	const peerAudioPlayer = new AudioPlayer();
	$: peerAudioPlayer.setVolume($ambianceEffectsVolume);

	let entitiesElement: HTMLDivElement;
	let ro: ResizeObserver;
	let safeboxWidth = '100%';
	let safeboxHeight = '100%';

	function recomputeSafebox() {
		let smallestWidth: number | undefined;
		let smallestHeight: number | undefined;
		for (const size of Object.values(screenSizesByStreamID)) {
			if (size.width > 0 && size.height > 0) {
				smallestWidth = smallestWidth ? Math.min(smallestWidth, size.width) : size.width;
				smallestHeight = smallestHeight ? Math.min(smallestHeight, size.height) : size.height;
			}
		}

		if (typeof smallestWidth !== 'undefined' && typeof smallestHeight !== 'undefined') {
			safeboxWidth = `${smallestWidth}px`;
			safeboxHeight = `${smallestHeight}px`;
		} else {
			safeboxWidth = '100%';
			safeboxHeight = '100%';
		}
	}

	let lastReceivedAmbianceID = '';

	onMount(async function () {
		const { Client, LocalStream } = await import('ion-sdk-js'); // TODO: top import

		const peer = new Peer(Client, tableName, {
			onAmbiance: (_streamID: string, id: string) => {
				lastReceivedAmbianceID = id;
				ambianceID.set(id);
			},

			onEntity: (streamID: string, entity: Entity) => {
				let e = entitiesByKey[entity.key];
				if (!e) {
					e = entity;
					entitiesByKey[entity.key] = entity;
					entitiesByStreamID[streamID] = entitiesByStreamID[streamID] || [];
					entitiesByStreamID[streamID].push(entity);
					entities = [...entities, entity];
				}
				e.props = entity.props;
				entities = entities;
			},

			onScreen: (streamID: string, width: number, height: number) => {
				screenSizesByStreamID[streamID] = { width, height };
				recomputeSafebox();
			},

			onSpeaker: (streamIDs: any) => {
				for (const avatar of avatars) {
					avatar.active = false;
				}
				for (const streamID of streamIDs) {
					const avatar = avatarsByStreamID[streamID];
					if (avatar) {
						avatar.active = true;
					}
				}
				avatars = avatars;
			},

			onTrack: (stream: any) => {
				let avatar = avatarsByStreamID[stream.id];
				if (!avatar) {
					avatar = {};
					avatarsByStreamID[stream.id] = avatar;
					avatars = [...avatars, avatar];

					peerAudioPlayer.play(ambiances[get(ambianceID)]?.loginURL);
				}
				avatar.srcObject = stream;
				avatars = avatars;
			},

			onTrackRemoved: (stream: any) => {
				const avatarForStream = avatarsByStreamID[stream.id];
				if (avatarForStream) {
					delete avatarsByStreamID[stream.id];
					avatars = removeFromArray(avatars, avatarForStream);

					peerAudioPlayer.play(ambiances[get(ambianceID)]?.logoutURL);
				}

				const entitiesForStream = entitiesByStreamID[stream.id];
				if (entitiesForStream) {
					delete entitiesByStreamID[stream.id];
					for (const entity of entitiesForStream) {
						delete entitiesByKey[entity.key];
						entities = removeFromArray(entities, entity);
					}
				}

				delete screenSizesByStreamID[stream.id];
				recomputeSafebox();
			},
		});

		ambianceID.subscribe((v) => {
			// Idempotent
			if (v === lastReceivedAmbianceID) {
				return;
			}
			peer.reliable({ AMBIANCE: v });
		});

		/* Watch for mouse events */

		function onMouseEvent(event: any) {
			cursor.props.clientX = event.clientX;
			cursor.props.clientY = event.clientY;
			cursor.props.isDown = isMouseDown;
			peer.unreliable({ ENTITY: cursor });
		}

		onMouseDown = (event) => {
			isMouseDown = event.button === 0;
			onMouseEvent(event);
		};

		onMouseMove = throttle(onMouseEvent, 1000 / MAX_FPS);

		onMouseUp = function (event) {
			isMouseDown = false;
			onMouseEvent(event);
		};

		/* Watch for smartbar entities */

		onSmartBarEntity = function (event: any) {
			const entity = event.detail;
			entities = [...entities, entity];
			peer.reliable({ ENTITY: entity });
		};

		/* Watch for resize */

		let screenReliable: any;
		ro = new ResizeObserver(
			throttle((entries) => {
				clearTimeout(screenReliable);
				const { width, height } = entries[0].contentRect;
				const message = { SCREEN: { width, height } };
				peer.unreliable(message);
				screenReliable = setTimeout(() => peer.reliable(message), 250);
			}, 1000 / MAX_FPS),
		);
		ro.observe(entitiesElement);

		/* Detect and publish streams */

		const streamManager = new StreamManager(LocalStream, MAX_FPS);
		for (const store of [audioInputDeviceID, videoInputDeviceID]) {
			store.subscribe(async () => {
				const out = await streamManager.reset(get(audioInputDeviceID), get(videoInputDeviceID));
				if (!out) {
					return;
				}
				const [newStream, _hasAudioChanged, hasVideoChanged] = out;

				if (newStream) {
					// Immediate UI feedback
					if (hasVideoChanged) {
						let avatar = avatarsByStreamID['__LOCAL_STREAM__'];
						if (!avatar) {
							avatar = { muted: true };
							avatars = [...avatars, avatar];
						}
						avatarsByStreamID['__LOCAL_STREAM__'] = avatar;
						avatarsByStreamID[newStream.id] = avatar;
						avatar.srcObject = newStream;
						avatars = avatars;
					}

					// WebRTC publish
					peer.publish(newStream);
				}
			});
		}

		const [aDeviceID, vDeviceID] = await StreamManager.detect();
		audioInputDeviceID.set(aDeviceID);
		videoInputDeviceID.set(vDeviceID);
	});

	onDestroy(() => {
		ro && ro.disconnect();
	});
</script>

<Badge
	count={avatars.length}
	color={faviconColor}
	background={faviconBackground}
	href="/favicon.png"
/>

<SmartBar on:entity={onSmartBarEntity}>
	<div
		class="container"
		on:mousedown={onMouseDown}
		on:mousemove={onMouseMove}
		on:mouseup={onMouseUp}
		style="
		  background-image: url({ambiances[$ambianceID]?.backgroundURL});
      background-position: center;
		  background-size: cover;
		  cursor: url({isMouseDown
			? ambiances[$ambianceID]?.clickURL
			: ambiances[$ambianceID]?.cursorURL}), auto;
		"
	>
		<header>
			{#each avatars as props}
				<Avatar {...props} />
			{/each}
			<div />
		</header>

		<div class="entities" bind:this={entitiesElement}>
			<div class="safebox" style="width: {safeboxWidth}; height: {safeboxHeight};">
				{#each entities as entity}
					<div class="entity">
						<EntityComponent {entity} />
					</div>
				{/each}
			</div>
		</div>

		<div class="settings-button" on:click={() => (showSettings = !showSettings)}>
			<img src={settingsButtonURL} alt="Settings" />
		</div>

		{#if showSettings}
			<div class="settings-wrapper">
				<Settings on:close={() => (showSettings = false)} />
			</div>
		{/if}
	</div>
</SmartBar>

<style>
	.container {
		height: 100vh;
		display: flex;
		flex-direction: column;
	}

	header {
		display: flex;
		justify-content: center;
		flex-wrap: wrap;
	}

	.entities {
		width: 100%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}

	@media screen and (max-width: 600px), screen and (max-height: 600px) {
		.entities {
			display: none;
		}
	}

	.safebox {
		max-width: 100%;
		max-height: 100%;
		border: 2px dashed grey;
	}

	.settings-button {
		position: absolute;
		top: 30px;
		right: 30px;
		border-radius: 50%;
		padding: 5px;
		background-color: orange;
	}

	.settings-button:hover {
		cursor: pointer;
		background-color: brown;
	}

	.settings-button img {
		width: 30px;
		height: 30px;
	}

	.settings-wrapper {
		z-index: 1000;
		position: absolute;
		left: 50%;
		top: 30%;
		transform: translate(-50%, -30%);
	}
</style>
