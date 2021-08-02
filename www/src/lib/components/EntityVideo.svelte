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
	<!-- svelte-ignore a11y-media-has-caption -->
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
		background: #997447;
	}

	legend {
		border-radius: 30px;
		margin-top: 3px;
		padding: 3px;
		background: #997447;
		text-align: center;
	}
</style>
