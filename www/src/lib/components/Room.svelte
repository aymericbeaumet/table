<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';
	import ContextMenu from './ContextMenu.svelte';

	const contextMenuIcons = [
		['togglecam', 'cam', 'nocam'],
		['togglemic', 'mic', 'nomic'],
		['settings', 'set', 'set'],
	];

	const contextMenuItems = [
		['pintext', 'Pin text'],
		['pinmedia', 'Pin media'],
	];

	function onContextMenuAction(event) {
		console.log('action', event.detail);
	}

	let fixedEntities = [];
	let movableEntities = [];

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
		fixedEntities = [...fixedEntities, [EntityVideo, { srcObject: stream }]];
	});
</script>

<ContextMenu icons={contextMenuIcons} items={contextMenuItems} on:action={onContextMenuAction}>
	<div class="container">
		<header>
			{#each fixedEntities as [entity, props]}
				<div class="entity">
					<svelte:component this={entity} {...props} />
				</div>
			{/each}
		</header>

		{#each movableEntities as [entity, props]}
			<div class="entity">
				<svelte:component this={entity} {...props} />
			</div>
		{/each}
	</div>
</ContextMenu>

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
		height: 150px;
		width: 150px;
	}
</style>
