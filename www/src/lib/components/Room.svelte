<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import PinBar from './PinBar.svelte';
	import Network from '$lib/network';
	import { throttle } from 'lodash';

	const MAX_FPS = 30;

	export let name;

	let network;
	let topEntities = [];
	let entities = [];

	function onEntity(event) {
		const [component, props] = event.detail;
		entities = [...entities, [component, props]];
	}

	function onMouseMove(event) {
		network.send(
			{
				MOUSE_MOVE: {
					clientX: event.clientX,
					clientY: event.clientY,
				},
			},
			{ reliable: false },
		);
	}

	onMount(async () => {
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

		network = new Network(name, stream);
		await network.init();
	});
</script>

<PinBar on:entity={onEntity}>
	<div class="container" on:mousemove={network && throttle(onMouseMove, 1000 / MAX_FPS)}>
		<header>
			{#each topEntities as [entity, props]}
				<div>
					<svelte:component this={entity} {...props} />
				</div>
			{/each}
		</header>

		{#each entities as [entity, props]}
			<div class="entity">
				<svelte:component this={entity} {...props} />
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
