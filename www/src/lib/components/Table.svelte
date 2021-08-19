<script>
	import logoURL from '$lib/assets/logo.png';
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
	import wandButtonURL from '$lib/assets/wand.svg';
	import Settings from '$lib/components/Settings.svelte';
	import ambiances from '$lib/assets/ambiances';
	import { get } from 'svelte/store';
	import StreamManager from '$lib/StreamManager';
	import EntityComponent from '$lib/components/Entity.svelte';
	import type { LocalStream as LocalStreamType } from 'ion-sdk-js';

	const faviconBackground = '#FF0000';
	const faviconColor = '#FFFFFF';

	export let tableName: string;

	const avatarsByStreamID: any = {};
	let avatars: Array<any> = [];

	const entitiesByKey: { [key: string]: Entity } = {};
	const entitiesByStreamID: { [key: string]: Array<Entity> } = {};
	let entities: Array<Entity> = [];

	const screenSizesByStreamID: { [key: string]: { width: number; height: number } } = {};

	const cursor = new Entity('cursor');
	let isMouseDown = false;
	let onMouseMove: svelte.JSX.MouseEventHandler<HTMLDivElement>;
	let onMouseDown: svelte.JSX.MouseEventHandler<HTMLDivElement>;
	let onMouseUp: svelte.JSX.MouseEventHandler<HTMLDivElement>;
	let onSmartBarEntity: (entity: CustomEvent<Entity>) => void;

	let modal: 'smartbar' | 'settings' | undefined;
	let smartbarInitialValue = '';

	const backgroundAudioPlayer = new AudioPlayer();
	backgroundAudioPlayer.setLoop(true);
	$: backgroundAudioPlayer.setVolume($ambianceSoundVolume);
	$: backgroundAudioPlayer.play(ambiances[$ambianceID]?.soundURL);

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
	let lastReceivedAmbianceSoundVolume = 0;
	let lastReceivedAmbianceEffectsVolume = 0;

	onMount(async function () {
		const { Client, LocalStream } = await import('ion-sdk-js'); // TODO: top import

		const peer = new Peer(Client, tableName, {
			onAmbianceID: (_streamID: string, id: string) => {
				lastReceivedAmbianceID = id;
				ambianceID.set(id);
			},

			onAmbianceSoundVolume: (_streamID: string, volume: number) => {
				lastReceivedAmbianceSoundVolume = volume;
				ambianceSoundVolume.set(volume);
			},

			onAmbianceEffectsVolume: (_streamID: string, volume: number) => {
				lastReceivedAmbianceEffectsVolume = volume;
				ambianceEffectsVolume.set(volume);
			},

			onEntity: (streamID: string, entity: Entity) => {
				let e = entitiesByKey[entity.key];
				if (!e) {
					e = entity;
					entitiesByKey[entity.key] = entity;
					entitiesByStreamID[streamID] = entitiesByStreamID[streamID] || [];
					entitiesByStreamID[streamID]!.push(entity);
					entities = [...entities, entity];
				}
				e.props = entity.props;
				entities = entities;
			},

			onScreen: (streamID: string, width: number, height: number) => {
				screenSizesByStreamID[streamID] = { width, height };
				recomputeSafebox();
			},

			onSpeaker: (streamIDs: Array<string>) => {
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

			onTrack: (stream: LocalStreamType) => {
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

			onTrackRemoved: (stream: LocalStreamType) => {
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
			if (v === lastReceivedAmbianceID) {
				return;
			}
			peer.reliable({ AMBIANCE_ID: v });
		});

		ambianceSoundVolume.subscribe((v) => {
			if (v === lastReceivedAmbianceSoundVolume) {
				return;
			}
			peer.reliable({ AMBIANCE_SOUND_VOLUME: v });
		});

		ambianceEffectsVolume.subscribe((v) => {
			if (v === lastReceivedAmbianceEffectsVolume) {
				return;
			}
			peer.reliable({ AMBIANCE_EFFECTS_VOLUME: v });
		});

		/* Watch for mouse events */

		function onMouseEvent(event: MouseEvent) {
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

		onSmartBarEntity = function (event: CustomEvent<Entity>) {
			const entity = event.detail;
			entities = [...entities, entity];
			peer.reliable({ ENTITY: entity });
		};

		/* Watch for resize */

		let screenReliableTimeout: ReturnType<typeof setTimeout> | undefined;
		ro = new ResizeObserver(
			throttle((entries) => {
				if (screenReliableTimeout) {
					clearTimeout(screenReliableTimeout);
					screenReliableTimeout = undefined;
				}

				const { width, height } = entries[0].contentRect;
				const message = { SCREEN: { width, height } };
				peer.unreliable(message);

				screenReliableTimeout = setTimeout(() => peer.reliable(message), 250);
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

	const onKeyDown = (event: KeyboardEvent) => {
		if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
			event.stopPropagation();
			event.preventDefault();
			if (modal !== 'smartbar') {
				displaySmartbar();
			} else {
				modal = undefined;
			}
			return;
		}
	};

	const onPaste = (event: ClipboardEvent) => {
		displaySmartbar(event.clipboardData?.getData('text'));
	};

	function displaySmartbar(initialValue = '') {
		if (!modal) {
			smartbarInitialValue = initialValue;
			modal = 'smartbar';
		}
	}

	function displaySettings() {
		modal = 'settings';
	}
</script>

<svelte:window on:keydown={onKeyDown} on:paste={onPaste} />

<Badge
	count={avatars.length}
	color={faviconColor}
	background={faviconBackground}
	href="/favicon.png"
/>

<div
	class="container"
	on:mousedown={onMouseDown}
	on:mousemove={onMouseMove}
	on:mouseup={onMouseUp}
	style="
    background-image: url({ambiances[$ambianceID]?.backgroundURL});
    cursor: url({isMouseDown
		? ambiances[$ambianceID]?.clickURL
		: ambiances[$ambianceID]?.cursorURL}), auto;
  "
>
	<header>
		<div>
			<a class="logo" rel="external" href="/"><img src={logoURL} alt="Table" /></a>
		</div>

		<div>
			<button class:active={modal === 'smartbar'} on:click={() => displaySmartbar()}>
				<img src={wandButtonURL} alt="Smart Bar (CMD + K)" />
			</button>

			<button class:active={modal === 'settings'} on:click={displaySettings}>
				<img src={settingsButtonURL} alt="Settings" />
			</button>
		</div>
	</header>

	<main>
		{#if modal === 'smartbar'}
			<SmartBar
				initialValue={smartbarInitialValue}
				on:entity={onSmartBarEntity}
				on:close={() => (modal = undefined)}
			/>
		{:else if modal === 'settings'}
			<Settings on:close={() => (modal = undefined)} />
		{/if}

		<div class="avatars">
			{#each avatars as props}
				<Avatar {...props} />
			{/each}
			<div />
		</div>

		<div class="entities" bind:this={entitiesElement}>
			<div class="safebox" style="width: {safeboxWidth}; height: {safeboxHeight};">
				{#each entities as entity}
					<div class="entity">
						<EntityComponent {entity} />
					</div>
				{/each}
			</div>
		</div>
	</main>
</div>

<style>
	.container {
		height: 100vh;
		display: flex;
		flex-direction: column;
		background-position: center;
		background-size: cover;
	}

	header {
		position: absolute;
		top: 0;
		width: 100%;
		display: flex;
		justify-content: space-between;
		padding: 20px;
	}

	header img {
		width: 48px;
		height: 48px;
		background: lightgrey;
		padding: 7px;
		border-radius: 50%;
	}

	header img:hover {
		background: orange;
	}

	header button {
		background: none;
		border: none;
		outline: none;
		margin-left: 10px;
	}

	header button:hover {
		cursor: pointer;
	}

	header button.active img {
		background: orange;
	}

	main {
		height: 100%;
		display: flex;
		flex-direction: column;
	}

	.avatars {
		display: flex;
		justify-content: center;
		flex-wrap: wrap;
		padding: 15px;
	}

	.entities {
		width: 100%;
		flex: 1;
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.safebox {
		max-width: 100%;
		max-height: 100%;
		border: 2px dashed grey;
	}

	@media screen and (max-width: 600px), screen and (max-height: 600px) {
		header,
		.entities {
			display: none;
		}
	}
</style>
