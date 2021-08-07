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
	import campfireMp3 from '$lib/assets/campfire.mp3';
	import { backgroundURL } from '$lib/stores';
	import Entity from '$lib/entity';
	import { MAX_FPS } from '$lib/env';

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

	let onSmartBarEntity;

	onMount(async function () {
		const backgroundAudioPlayer = new AudioPlayer();
		backgroundAudioPlayer.setLoop(true);
		backgroundAudioPlayer.setVolume(0.3);
		backgroundAudioPlayer.play(campfireMp3);

		const peerAudioPlayer = new AudioPlayer();

		const { Client } = await import('ion-sdk-js'); // TODO: top import
		const peer = new Peer(Client, name, {
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

		// Acquire stream
		const { LocalStream } = await import('ion-sdk-js'); // TODO: top import
		const stream = await LocalStream.getUserMedia({
			audio: {
				autoGainControl: true,
				echoCancellation: true,
				noiseSuppression: true,
			},
			video: {
				width: { ideal: 360, max: 360 },
				height: { ideal: 360, max: 360 },
				facingMode: { ideal: 'user' },
				frameRate: { ideal: MAX_FPS, max: MAX_FPS },
			},
			simulcast: true,
		});
		// Immediate UI feedback
		const avatar = { srcObject: stream, muted: true };
		avatarsByStreamID[stream.id] = avatar;
		avatars = [...avatars, avatar];
		// WebRTC publish
		peer.publish(stream);
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
		  background: repeat url({$backgroundURL});
		  cursor: url({isMouseDown ? gauntletGlowPNG : gauntletPNG}), auto;
		"
	>
		<header style="background: repeat url({leatherPng});">
			{#each avatars as props}
				<Avatar {...props} />
			{/each}
		</header>

		{#each entities as [component, props]}
			<div class="entity">
				<svelte:component this={component} {...props} />
			</div>
		{/each}
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
</style>
