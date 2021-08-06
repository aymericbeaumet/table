<script>
	import videoPosterPng from '$lib/assets/video-poster.png';
	import { audioEnabled, videoEnabled } from '$lib/stores';

	export let muted = false;
	export let srcObject = null;
	export let toggles = false;

	let video;
	$: if (video && video.srcObject !== srcObject) {
		video.srcObject = srcObject;
	}

	function onAudioToggleClick(event) {
		event.preventDefault();
		audioEnabled.update((v) => !v);
	}

	function onVideoToggleClick(event) {
		event.preventDefault();
		videoEnabled.update((v) => !v);
	}
</script>

<div>
	<!-- svelte-ignore a11y-media-has-caption -->
	<video
		bind:this={video}
		autoplay
		disablepictureinpicture
		disableremoteplayback
		bind:muted
		playsinline
		style="background: repeat url({videoPosterPng});"
	/>

	<footer>
		{#if toggles}
			<button on:click={onAudioToggleClick} class:active={$audioEnabled}>A</button>
			<button on:click={onVideoToggleClick} class:active={$videoEnabled}>V</button>
		{/if}
	</footer>
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

	footer {
		display: flex;
		justify-content: center;
		margin-top: 3px;
		padding: 3px;
		text-align: center;
	}

	button {
		border: 1px solid black;
		padding: 0 6px;
		margin: 0 3px;
		cursor: pointer;
		text-decoration: none;
	}

	button:not(.active) {
		text-decoration: line-through;
	}
</style>
