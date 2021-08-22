<script>
	import ambiances from '$lib/assets/ambiances';
	import { ambianceID } from '$lib/stores';

	export let active = false;
	export let muted = false;
	export let srcObject: HTMLMediaElement['srcObject'] | null = null;

	let video: HTMLVideoElement;
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
		bind:muted
		playsinline
		style="background: url({ambiances[$ambianceID].posterURL});"
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
		border: 2px solid #997447;
		object-fit: cover;
	}

	video.active {
		border: 2px solid lightgrey;
	}
</style>
