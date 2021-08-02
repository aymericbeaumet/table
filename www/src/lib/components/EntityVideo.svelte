<script>
	export let label;
	export let srcObject;
	export let muted = false;

	let video;
	$: if (srcObject && video && video.srcObject !== srcObject) {
		video.srcObject = srcObject;
	}
</script>

<video bind:this={video} bind:muted autoplay playsinline />
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
