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
	import leatherPng from '$lib/assets/leather.png';
	import {
		audioInputDeviceID,
		videoInputDeviceID,
		backgroundImage,
		backgroundSound,
	} from '$lib/stores';
	import Entity from '$lib/entity';
	import { MAX_FPS } from '$lib/env';
	import SettingsSVG from '$lib/assets/settings.svg';
	import Settings from '$lib/components/Settings.svelte';
	import { backgroundImages, backgroundSounds } from '$lib/resources';
	import { get } from 'svelte/store';

	const faviconBackground = '#FF0000';
	const faviconColor = '#FFFFFF';

	export let name;

	const avatarsByStreamID = {};
	let avatars = [];

	const entitiesByKey = {};
	let entities = [];

	const cursor = new Entity('cursor');
	let isMouseDown = false;
	let onMouseMove, onMouseDown, onMouseUp;

	let showSettings = false;

	let onSmartBarEntity;

	const backgroundAudioPlayer = new AudioPlayer();
	backgroundAudioPlayer.setLoop(true);
	backgroundAudioPlayer.setVolume(0.3);
	backgroundSound.subscribe((resource) => backgroundAudioPlayer.play(resource.url));

	onMount(async function () {
		const peerAudioPlayer = new AudioPlayer();

		const { Client } = await import('ion-sdk-js'); // TODO: top import
		const peer = new Peer(Client, name, {
			onBackgroundImage: function (_streamID, id) {
				backgroundImage.set(backgroundImages.find((r) => r.id === id));
			},

			onBackgroundSound: function (_streamID, id) {
				backgroundSound.set(backgroundSounds.find((r) => r.id === id));
			},

			onEntity: function (_streamID, entity) {
				let e = entitiesByKey[entity.key];
				if (!e) {
					e = entity.asComponentProps();
					entitiesByKey[entity.key] = e;
					entities = [...entities, e];
				}
				e[1] = entity.props;
				entities = entities;
			},

			onSpeaker: function (streamIDs) {
				for (const avatar of avatars) {
					avatar.active = false;
				}
				for (const streamID of streamIDs) {
					avatarsByStreamID[streamID].active = true;
				}
				avatars = avatars;
			},

			onTrack: function (stream) {
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

			onTrackRemoved: function (stream) {
				const avatar = avatarsByStreamID[stream.id];
				if (avatar) {
					delete avatarsByStreamID[stream.id];
					avatars = removeFromArray(avatars, avatar);
					peerAudioPlayer.play(stealthOGG);
				}
			},
		});

		function onMouseEvent(event) {
			cursor.props.clientX = event.clientX;
			cursor.props.clientY = event.clientY;
			cursor.props.isDown = isMouseDown;
			peer.unreliable({ ENTITY: cursor });
		}

		onMouseDown = function (event) {
			isMouseDown = event.which === 1;
			onMouseEvent(event);
		};

		onMouseMove = throttle(function (event) {
			onMouseEvent(event);
		}, 1000 / MAX_FPS);

		onMouseUp = function (event) {
			isMouseDown = false;
			onMouseEvent(event);
		};

		onSmartBarEntity = function (event) {
			const entity = event.detail;
			entities = [...entities, entity.asComponentProps()];
			peer.reliable({ ENTITY: entity });
		};

		const { LocalStream } = await import('ion-sdk-js'); // TODO: top import

		audioInputDeviceID.subscribe(initStream);
		videoInputDeviceID.subscribe(initStream);

		async function initStream() {
			// Get user media
			const stream = await LocalStream.getUserMedia({
				audio: {
					autoGainControl: true,
					echoCancellation: true,
					noiseSuppression: true,
					deviceId: get(audioInputDeviceID),
				},
				video: {
					width: { ideal: 360, max: 360 },
					height: { ideal: 360, max: 360 },
					facingMode: { ideal: 'user' },
					frameRate: { ideal: MAX_FPS, max: MAX_FPS },
					deviceId: get(videoInputDeviceID),
				},
				simulcast: true,
			});

			// Immediate UI feedback
			let avatar = avatarsByStreamID['__LOCAL_STREAM__'];
			if (!avatar) {
				avatar = { muted: true };
				avatars = [...avatars, avatar];
			}
			avatarsByStreamID['__LOCAL_STREAM__'] = avatar;
			avatarsByStreamID[stream.id] = avatar;
			avatar.srcObject = stream;
			avatars = avatars;

			// WebRTC publish
			peer.publish(stream);
		}
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
		  background: repeat url({$backgroundImage.url});
		  cursor: url({isMouseDown ? gauntletGlowPNG : gauntletPNG}), auto;
		"
	>
		<header style="background: repeat url({leatherPng});">
			{#each avatars as props}
				<Avatar {...props} />
			{/each}
			<div />
		</header>

		{#each entities as [component, props]}
			<div class="entity">
				<svelte:component this={component} {...props} />
			</div>
		{/each}

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
		margin: 0 auto;
		padding: 5px;
		border: 3px solid #281b0d;
		border-top: 0;
		border-radius: 20px;
		border-top-left-radius: 0;
		border-top-right-radius: 0;
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
		width: 100%;
		height: 100%;
		position: absolute;
	}
</style>
