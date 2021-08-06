<script>
	import videoPosterPng from '$lib/assets/video-poster.png';

	export let active = false;
	export let muted = false;
	export let srcObject = null;

	let video;
	$: if (video && video.srcObject !== srcObject) {
		video.srcObject = srcObject;
	}
</script>

<div>
	<!-- svelte-ignore a11y-media-has-caption -->
	<video
		bind:this={video}
		class:active
		autoplay
		disablepictureinpicture
		disableremoteplayback
		bind:muted
		playsinline
		style="background: repeat url({videoPosterPng});"
	/>
</div>

<style>
	div {
		padding: 5px;
		width: 150px;
		height: 150px;
	}

	video {
		transform: scale(-1, 1);
		border-radius: 50%;
		width: 100%;
		height: 100%;
		border: 1px solid #997447;
	}

	video.active {
		border: 1px solid red;
	}
</style>
