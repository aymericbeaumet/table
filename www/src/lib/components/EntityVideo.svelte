<script>
	export let label = '';
	export let srcObject = null;
	export let muted = false;

	let video;
	$: if (srcObject && video && video.srcObject !== srcObject) {
		video.srcObject = srcObject;
	}
</script>

<div>
	<video bind:this={video} bind:muted autoplay playsinline />
	<legend>{label}</legend>
</div>

<style>
	div {
		padding: 5px;
	}

	video {
		transform: scale(-1, 1);
		border-radius: 50%;
		width: 150px;
		height: 150px;
		border: 1px solid grey;
		background-color: lightgrey;
	}

	legend {
		border-radius: 30px;
		padding: 3px;
		background: lightgrey;
		text-align: center;
	}
</style>
