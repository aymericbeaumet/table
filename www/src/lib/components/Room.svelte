<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import PinBar from './PinBar.svelte';

	let topEntities = [];
	let entities = [];

	function onEntity(event) {
		const [component, props] = event.detail;
		entities = [...entities, [component, props]];
	}

	onMount(async () => {
		const stream = await navigator.mediaDevices.getUserMedia({
			audio: true,
			video: {
				width: { ideal: 360, max: 360 },
				height: { ideal: 360, max: 360 },
				facingMode: { ideal: 'user' },
				frameRate: { ideal: 30, max: 30 },
			},
		});
		topEntities = [...topEntities, [EntityVideo, { srcObject: stream }]];
	});
</script>

<PinBar on:entity={onEntity}>
	<div class="container">
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
	}

	.entity {
		position: absolute;
		border: 1px solid grey;
		background-color: lightgrey;
	}
</style>
