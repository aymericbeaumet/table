<script>
	import { onMount } from 'svelte';
	import EntityVideo from './EntityVideo.svelte';

	export let name;

	function copyJoiningInfo() {
		navigator.clipboard.writeText(window.location.toString());
	}

	let entities = [];

	onMount(async () => {
		const stream = await navigator.mediaDevices.getUserMedia({
			audio: true,
			video: {
				width: { ideal: 360, max: 360 },
				height: { ideal: 360, max: 360 },
				facingMode: { ideal: 'user' },
				frameRate: { ideal: 30, max: 30 }
			}
		});
		entities = [...entities, [EntityVideo, { srcObject: stream }]];
	});
</script>

<div class="container">
	<main>
		{#each entities as [entity, props]}
			<div class="entity">
				<svelte:component this={entity} {...props} />
			</div>
		{/each}
	</main>

	<footer>
		<div>
			<h1 class="clickable" title="Copy room URL" on:click={copyJoiningInfo}>{name}</h1>
		</div>

		<div>
			<button title="Turn off microphone">mic</button>
			<button title="Turn off camera">cam</button>
		</div>

		<div>
			<button title="Settings">settings</button>
		</div>
	</footer>
</div>

<style>
	.container {
		height: 100vh;
		display: flex;
		flex-direction: column;
	}

	main {
		height: 100%;
		background-color: lightyellow;
	}

	.entity {
		height: 150px;
		width: 150px;
	}

	footer {
		border-top: 1px solid black;
		background-color: lightgrey;
		padding: 10px;
		display: flex;
		justify-content: space-between;
	}

	h1 {
		display: inline-block;
	}

	.clickable:hover {
		cursor: pointer;
	}
</style>
