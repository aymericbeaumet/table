<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import SmartBar from './SmartBar.svelte';
	import Peer from '$lib/peer';
	import { throttle } from 'lodash';
	import EntityCursor from './EntityCursor.svelte';
	import { removeFromArray } from '$lib/utils';
	import * as audio from '$lib/audio';
	import murlocOGG from '$lib/assets/murloc.ogg';
	import stealthOGG from '$lib/assets/stealth.ogg';
	import gauntletPNG from '$lib/assets/gauntlet.png';
	import gauntletGlowPNG from '$lib/assets/gauntlet-glow.png';

	const MAX_FPS = 30;

	export let name;

	let topEntities = [];
	let entities = [];

	function onSmartBarEntity(event) {
		const entity = event.detail;
		entities = [...entities, entity];
		// TODO: send entity through network
	}

	let isMouseDown = false;
	let onMouseMove, onMouseDown, onMouseUp;

	onMount(async function () {
		// TODO: find a way to use the top imports for the 'ion-sdk-js' library
		// It fails server-side (SSR) because of `ReferenceError: MediaStream is not defined`
		const { Client, LocalStream } = await import('ion-sdk-js');

		// Some hashmaps indexed by peer ids
		const latestKnownStatus = {};
		const cursorEntities = {};
		const videoEntities = {};

		const peer = new Peer(name, Client, {
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
					audio.play(murlocOGG);
					return;
				}

				if (data.status === 'offline') {
					audio.play(stealthOGG);

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

			onTrack: function (peerID, stream) {
				if (!videoEntities[peerID]) {
					const entity = ['peer', EntityVideo, { srcObject: stream }];
					videoEntities[peerID] = entity;
					topEntities = [...topEntities, entity];
				}
			},
		});

		function sendMouseEvent(event) {
			peer.sendUnreliable({
				MOUSE: { clientX: event.clientX, clientY: event.clientY, isDown: isMouseDown },
			});
		}

		onMouseDown = throttle(function (event) {
			isMouseDown = event.which === 1;
			sendMouseEvent(event);
		}, 1000 / MAX_FPS);

		onMouseMove = throttle(function (event) {
			sendMouseEvent(event);
		}, 1000 / MAX_FPS);

		onMouseUp = throttle(function (event) {
			isMouseDown = false;
			sendMouseEvent(event);
		}, 1000 / MAX_FPS);

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
		peer.publishStream(stream);
		topEntities = [...topEntities, ['self', EntityVideo, { srcObject: stream, muted: true }]];
	});
</script>

<SmartBar on:entity={onSmartBarEntity}>
	<div
		class="container"
		on:mousedown={onMouseDown}
		on:mousemove={onMouseMove}
		on:mouseup={onMouseUp}
		style="cursor: url({isMouseDown ? gauntletGlowPNG : gauntletPNG}), auto;"
	>
		<header>
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
		background-color: lightyellow;
	}

	header {
		display: flex;
		justify-content: center;
		padding: 10px;
	}

	.entity {
		position: absolute;
		border: 1px solid grey;
		background-color: lightgrey;
	}
</style>
