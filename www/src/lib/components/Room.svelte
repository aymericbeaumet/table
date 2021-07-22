<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import PinBar from './PinBar.svelte';
	import Network from '$lib/network';
	import { throttle } from 'lodash';
	import EntityText from './EntityText.svelte';
	import EntityIframe from './EntityIframe.svelte';

	const MAX_FPS = 30;

	export let name;

	let network;
	let topEntities = [];
	let entities = [];

	function getEntityByname(name) {
		switch (name) {
			case 'iframe':
				return EntityIframe;
			case 'text':
				return EntityText;
			default:
				throw new Error(`unknown entity ${name}`);
		}
	}

	function onEntity(event) {
		const [name, props] = event.detail;
		network.send(
			{
				NEW_ENTITY: {
					name,
					props,
				},
			},
			{ reliable: true },
		);
		const entity = getEntityByname(name);
		entities = [...entities, [entity, props]];
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

	function onMouseClick(event) {
		network.send(
			{
				MOUSE_CLICK: {
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
	<div
		class="container"
		on:mousemove={network && throttle(onMouseMove, 1000 / MAX_FPS)}
		on:click={network && onMouseClick}
	>
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
