<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import PinBar from './PinBar.svelte';
	import Network from '$lib/network';
	import { throttle } from 'lodash';
	import EntityCursor from './EntityCursor.svelte';

	const MAX_FPS = 30;

	export let name;

	let topEntities = [];
	let entities = [];

	function onPinBarEntity(event) {
		entities = [...entities, event.detail];
		// TODO: send entity through network
	}

	let cursorEntities = {};

	function onNetworkMouseClick(peerID, data) {
		console.log(peerID, data);
	}

	function onNetworkMouseMove(peerID, data) {
		let entity = cursorEntities[peerID];
		if (!entity) {
			entity = [EntityCursor, { fps: MAX_FPS }];
			cursorEntities[peerID] = entity;
			entities = [...entities, entity];
		}
		entity[1].x = data.clientX;
		entity[1].y = data.clientY;
		entities = entities;
	}

	function onNetworkPeer(peerID, data) {
		if (data.status === 'offline') {
			const cursorEntity = cursorEntities[peerID];
			if (cursorEntity) {
				delete cursorEntities[peerID];
				entities.splice(entities.indexOf(cursorEntity), 1);
				entities = entities;
			}
		}
	}

	let onMouseClick;
	let onMouseMove;

	onMount(async function () {
		const stream = await navigator.mediaDevices.getUserMedia({
			audio: true,
			video: {
				width: { ideal: 360, max: 360 },
				height: { ideal: 360, max: 360 },
				facingMode: { ideal: 'user' },
				frameRate: { ideal: MAX_FPS, max: MAX_FPS },
			},
		});
		topEntities = [...topEntities, [EntityVideo, { srcObject: stream }]];

		const network = new Network(name, stream, {
			onMouseClick: throttle(onNetworkMouseClick, 1000 / MAX_FPS),
			onMouseMove: throttle(onNetworkMouseMove, 1000 / MAX_FPS),
			onPeer: onNetworkPeer,
		});
		await network.init();

		onMouseClick = function (event) {
			network.unreliable({
				MOUSE_CLICK: {
					clientX: event.clientX,
					clientY: event.clientY,
				},
			});
		};

		onMouseMove = throttle(function (event) {
			network.unreliable({
				MOUSE_MOVE: {
					clientX: event.clientX,
					clientY: event.clientY,
				},
			});
		}, 1000 / MAX_FPS);
	});
</script>

<PinBar on:entity={onPinBarEntity}>
	<div class="container" on:click={onMouseClick} on:mousemove={onMouseMove}>
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
