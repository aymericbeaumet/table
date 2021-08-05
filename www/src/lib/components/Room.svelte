<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import SmartBar from './SmartBar.svelte';
	import Peer from '$lib/peer';
	import { throttle } from 'lodash';
	import EntityCursor from './EntityCursor.svelte';
	import { removeFromArray } from '$lib/utils';
	import AudioPlayer from '$lib/audio-player';
	import murlocOGG from '$lib/assets/murloc.ogg';
	import stealthOGG from '$lib/assets/stealth.ogg';
	import gauntletPNG from '$lib/assets/gauntlet.png';
	import gauntletGlowPNG from '$lib/assets/gauntlet-glow.png';
	import Badge from 'svelte-favicon-badge';
	import leatherPng from '$lib/assets/leather.png';
	import campfireMp3 from '$lib/assets/campfire.mp3';
	import { audioEnabled, videoEnabled, backgroundURL } from '$lib/stores';

	const MAX_FPS = 30;

	export let name;

	let topEntities = [];
	let entities = [];

	let faviconBackground = '#FF0000';
	let faviconColor = '#FFFFFF';

	function onSmartBarEntity(event) {
		const entity = event.detail;
		entities = [...entities, entity];
		// TODO: send entity through network
	}

	let isMouseDown = false;
	let onMouseMove, onMouseDown, onMouseUp;

	onMount(async function () {
		// Some hashmaps indexed by peer ids
		const latestKnownStatus = {};
		const cursorEntities = {};
		const videoEntities = {};

		const backgroundAudioPlayer = new AudioPlayer();
		backgroundAudioPlayer.setLoop(true);
		backgroundAudioPlayer.setVolume(0.3);
		backgroundAudioPlayer.play(campfireMp3);

		const peerAudioPlayer = new AudioPlayer();

		const { Client } = await import('ion-sdk-js'); // TODO: top import
		const peer = new Peer(Client, name, {
			onMouse: throttle(function (peerID, data) {
				let entity = cursorEntities[peerID];
				if (!entity) {
					entity = ['cursor', EntityCursor, { fps: MAX_FPS }];
					cursorEntities[peerID] = entity;
					entities = [...entities, entity];
				}
				entity[2].x = data.clientX;
				entity[2].y = data.clientY;
				entity[2].isDown = data.isDown;
				entities = entities;
			}, 1000 / MAX_FPS),

			onPeer: function (peerID, data) {
				// Make this handler idempotent
				if (latestKnownStatus[peerID] === data.status) {
					return;
				}
				latestKnownStatus[peerID] = data.status;

				if (data.status === 'online') {
					peerAudioPlayer.play(murlocOGG);

					if (!videoEntities[peerID]) {
						const entity = ['peer', EntityVideo, {}];
						videoEntities[peerID] = entity;
						topEntities = [...topEntities, entity];
					}

					return;
				}

				if (data.status === 'offline') {
					peerAudioPlayer.play(stealthOGG);

					const cursorEntity = cursorEntities[peerID];
					if (cursorEntity) {
						delete cursorEntities[peerID];
						entities = removeFromArray(entities, cursorEntity);
					}

					const videoEntity = videoEntities[peerID];
					if (videoEntity) {
						delete videoEntities[peerID];
						topEntities = removeFromArray(topEntities, videoEntity);
					}

					return;
				}
			},

			onTrack: function (_peerID, _track, stream) {
				let entity = videoEntities[stream.id];
				if (!entity) {
					entity = ['peer', EntityVideo, {}];
					videoEntities[stream.id] = entity;
					topEntities = [...topEntities, entity];
				}
				entity[2].srcObject = stream;
				topEntities = topEntities;
			},

			onTrackRemoved: function (_peerID, _track, stream) {
				const videoEntity = videoEntities[stream.id];
				if (videoEntity) {
					delete videoEntities[stream.id];
					topEntities = removeFromArray(topEntities, videoEntity);
				}
			},
		});

		function sendMouseEvent(event) {
			peer.unreliable({
				MOUSE: { clientX: event.clientX, clientY: event.clientY, isDown: isMouseDown },
			});
		}

		onMouseDown = function (event) {
			isMouseDown = event.which === 1;
			sendMouseEvent(event);
		};

		onMouseMove = throttle(function (event) {
			sendMouseEvent(event);
		}, 1000 / MAX_FPS);

		onMouseUp = function (event) {
			isMouseDown = false;
			sendMouseEvent(event);
		};

		initLocalStream(peer);
		audioEnabled.subscribe(() => initLocalStream(peer));
		videoEnabled.subscribe(() => initLocalStream(peer));
	});

	let lastStream = null;
	let videoStream = null;

	async function initLocalStream(peer) {
		const constraints = {
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
		};

		const { LocalStream } = await import('ion-sdk-js');
		const stream = await LocalStream.getUserMedia(constraints);
		if (!lastStream) {
			lastStream = stream;
			videoStream = stream;
			peer.publish(stream);
		}
	}
</script>

<Badge
	count={1 + topEntities.length}
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
			<EntityVideo srcObject={videoStream} muted={true} toggles={true} />

			{#each topEntities as [_, component, props]}
				<div>
					<svelte:component this={component} {...props} />
				</div>
			{/each}
		</header>

		{#each entities as [_, component, props]}
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
