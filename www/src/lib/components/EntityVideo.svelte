<script>
	import { onMount } from 'svelte';

	export let srcObject;
	export let muted = false;

	let video;

	onMount(() => {
		video.srcObject = srcObject;
		video.muted = muted;
		video.play();
	});
</script>

<video bind:this={video} />

<style>
	video {
		transform: scale(-1, 1);
		border-radius: 50%;
		width: 150px;
		height: 150px;
		border: 1px solid grey;
		background-color: lightgrey;
	}
</style>
