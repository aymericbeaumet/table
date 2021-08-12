<script>
	import { onMount } from 'svelte';
	import Avatar from './Avatar.svelte';
	import SmartBar from './SmartBar.svelte';
	import Peer from '$lib/peer';
	import { throttle } from 'lodash';
	import { removeFromArray } from '$lib/utils';
	import AudioPlayer from '$lib/audio-player';
	import murlocOGG from '$lib/assets/murloc.ogg';
	import stealthOGG from '$lib/assets/stealth.ogg';
	import gauntletPNG from '$lib/assets/gauntlet.png';
	import gauntletGlowPNG from '$lib/assets/gauntlet-glow.png';
	import Badge from 'svelte-favicon-badge';
	import {
		audioInputDeviceID,
		videoInputDeviceID,
		backgroundImageID,
		backgroundSoundID,
		backgroundSoundVolume,
	} from '$lib/stores';
	import Entity from '$lib/entity';
	import { MAX_FPS } from '$lib/env';
	import SettingsSVG from '$lib/assets/settings.svg';
	import Settings from '$lib/components/Settings.svelte';
	import { backgroundImages, backgroundSounds } from '$lib/resources';
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

	const cursor = new Entity('cursor');
	let isMouseDown = false;
	let onMouseMove: svelte.JSX.MouseEventHandler<HTMLDivElement>;
	let onMouseDown: svelte.JSX.MouseEventHandler<HTMLDivElement>;
	let onMouseUp: svelte.JSX.MouseEventHandler<HTMLDivElement>;

	let showSettings = false;

	let onSmartBarEntity: any;

	const backgroundAudioPlayer = new AudioPlayer();
	backgroundAudioPlayer.setLoop(true);
	$: backgroundAudioPlayer.setVolume($backgroundSoundVolume);
	$: backgroundAudioPlayer.play(backgroundSounds[$backgroundSoundID]?.url || '');

	$: backgroundImageURL = backgroundImages[$backgroundImageID]?.url;

	onMount(async function () {
		const { Client, LocalStream } = await import('ion-sdk-js'); // TODO: top import

		const peerAudioPlayer = new AudioPlayer();

		const peer = new Peer(Client, tableName, {
			onBackgroundImageID: (id: string) => backgroundImageID.set(id),

			onBackgroundSoundID: (id: string) => backgroundSoundID.set(id),

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
					peerAudioPlayer.play(murlocOGG);
				}
				avatar.srcObject = stream;
				avatars = avatars;
			},

			onTrackRemoved: (stream: any) => {
				const entitiesForStream = entitiesByStreamID[stream.id];
				if (entitiesForStream) {
					delete entitiesByStreamID[stream.id];
					for (const entity of entitiesForStream) {
						delete entitiesByKey[entity.key];
						entities = removeFromArray(entities, entity);
					}
				}

				const avatarForStream = avatarsByStreamID[stream.id];
				if (avatarForStream) {
					delete avatarsByStreamID[stream.id];
					avatars = removeFromArray(avatars, avatarForStream);
					peerAudioPlayer.play(stealthOGG);
				}
			},
		});

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

		onSmartBarEntity = function (event: any) {
			const entity = event.detail;
			entities = [...entities, entity];
			peer.reliable({ ENTITY: entity });
		};

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
		  background-image: url({backgroundImageURL});
      background-position: center;
		  background-size: cover;
		  cursor: url({isMouseDown ? gauntletGlowPNG : gauntletPNG}), auto;
		"
	>
		<header>
			{#each avatars as props}
				<Avatar {...props} />
			{/each}
			<div />
		</header>

		<div class="entities">
			{#each entities as entity}
				<div class="entity">
					<EntityComponent {entity} />
				</div>
			{/each}
		</div>

		<div class="settings-button" on:click={() => (showSettings = !showSettings)}>
			<img src={SettingsSVG} alt="Settings" />
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

	@media screen and (max-width: 600px), screen and (max-height: 600px) {
		.entities {
			display: none;
		}
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
		position: absolute;
		left: 50%;
		top: 30%;
		transform: translate(-50%, -30%);
	}
</style>
