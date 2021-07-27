<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import PinBar from './PinBar.svelte';
	import Peer from '$lib/peer';
	import { throttle } from 'lodash';
	import EntityCursor from './EntityCursor.svelte';

	const MAX_FPS = 30;

	export let name;

	let topEntities = [];
	let entities = [];

	function onPinBarEntity(event) {
		const entity = event.detail;
		entities = [...entities, entity];
		// TODO: send entity through network
	}

	let onMouseMove, onMouseDown, onMouseUp;

	onMount(async function () {
		const cursorEntities = {};

		const ws = new WebSocket(
			`${import.meta.env.VITE_WEBSOCKET_BASE_URL}/rooms/${encodeURIComponent(name)}`,
		);
		await new Promise((resolve) => (ws.onopen = resolve));

		const peer = new Peer(ws, {
			onMouseDown: throttle(function (peerID, data) {
				console.log('mouse down', peerID, data);
			}, 1000 / MAX_FPS),

			onMouseMove: throttle(function (peerID, data) {
				let entity = cursorEntities[peerID];
				if (!entity) {
					entity = [EntityCursor, { fps: MAX_FPS }];
					cursorEntities[peerID] = entity;
					entities = [...entities, entity];
				}
				entity[1].x = data.clientX;
				entity[1].y = data.clientY;
				entities = entities;
			}, 1000 / MAX_FPS),

			onMouseUp: throttle(function (peerID, data) {
				console.log('mouse up', peerID, data);
			}, 1000 / MAX_FPS),

			onPeer: function (peerID, data) {
				if (data.status === 'offline') {
					const cursorEntity = cursorEntities[peerID];
					if (cursorEntity) {
						delete cursorEntities[peerID];
						entities.splice(entities.indexOf(cursorEntity), 1);
						entities = entities;
					}
				}
			},

			onTrack: function (event) {
				const entity = [EntityVideo, { srcObject: event.streams[0] }];
				topEntities = [...topEntities, entity];
			},
		});

		onMouseDown = throttle(function (event) {
			peer.unreliable({ MOUSE_DOWN: { clientX: event.clientX, clientY: event.clientY } });
		}, 1000 / MAX_FPS);

		onMouseMove = throttle(function (event) {
			peer.unreliable({ MOUSE_MOVE: { clientX: event.clientX, clientY: event.clientY } });
		}, 1000 / MAX_FPS);

		onMouseUp = throttle(function (event) {
			peer.unreliable({ MOUSE_UP: { clientX: event.clientX, clientY: event.clientY } });
		}, 1000 / MAX_FPS);

		const stream = await navigator.mediaDevices.getUserMedia({
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
		});
		topEntities = [...topEntities, [EntityVideo, { srcObject: stream, muted: true }]];
		peer.addLocalStream(stream);
	});
</script>

<PinBar on:entity={onPinBarEntity}>
	<div
		class="container"
		on:mousedown={onMouseDown}
		on:mousemove={onMouseMove}
		on:mouseup={onMouseUp}
	>
		<header>
			{#each topEntities as [component, props]}
				<div>
					<svelte:component this={component} {...props} />
				</div>
			{/each}
		</header>

		{#each entities as [component, props]}
			<div class="entity">
				<svelte:component this={component} {...props} />
			</div>
		{/each}
	</div>
</PinBar>

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
		z-index: 100;
	}

	.entity {
		position: absolute;
		border: 1px solid grey;
		background-color: lightgrey;
	}
</style>
