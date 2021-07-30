<script>
	import { onMount } from 'svelte';

	export let label;
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
<legend>{label}</legend>

<style>
	video {
		transform: scale(-1, 1);
		border-radius: 50%;
		width: 150px;
		height: 150px;
		border: 1px solid grey;
		background-color: lightgrey;
	}

	legend {
		border: 1px solid lightcoral;
		border-radius: 10px;
		padding: 3px;
		background: lightsalmon;
		text-align: center;
	}
</style>
