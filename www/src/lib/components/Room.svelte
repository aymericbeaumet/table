<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import PinBar from './PinBar.svelte';
	import Network from '$lib/network';
	import { throttle } from 'lodash';

	const MAX_FPS = 30;

	export let name;

	let topEntities = [];
	let entities = [];

	function onPinBarEntity(event) {
		entities = [...entities, event.detail];
		// TODO: send entity through network
	}

	function onNetworkMessage({ peerID, data }) {
		console.log(peerID, data);
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

		const network = new Network(name, stream, onNetworkMessage);
		await network.init();

		onMouseClick = function (event) {
			network.reliable({
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
