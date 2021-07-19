<script>
	import { onMount } from 'svelte';

	export let srcObject;

	let video;

	onMount(() => {
		video.muted = true;
		video.srcObject = srcObject;
		video.play();
	});
</script>

<video bind:this={video} muted="muted" />

<style>
	video {
		transform: scale(-1, 1);
		border-radius: 50%;
		width: 100%;
		height: 100%;
	}
</style>
