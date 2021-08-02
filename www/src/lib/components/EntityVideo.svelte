<script>
	import videoPosterPng from '$lib/assets/video-poster.png';
	import { createEventDispatcher } from 'svelte';

	export let label = '';
	export let muted = false;
	export let poster = videoPosterPng;
	export let srcObject = null;
	export let toggles = false;

	const dispatch = createEventDispatcher();

	// toggle states
	let audioEnabled = true;
	let videoEnabled = true;

	let video;
	$: if (srcObject && video && video.srcObject !== srcObject) {
		video.srcObject = srcObject;
	}

	function onAudioToggleClick(event) {
		event.preventDefault();
		audioEnabled = !audioEnabled;
		dispatch('audiotoggle', audioEnabled);
	}

	function onVideoToggleClick(event) {
		event.preventDefault();
		videoEnabled = !videoEnabled;
		dispatch('videotoggle', videoEnabled);
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
		{poster}
	/>

	<legend>
		{#if toggles}
			<span on:click={onAudioToggleClick} class:active={audioEnabled}>A</span>
			<span on:click={onVideoToggleClick} class:active={videoEnabled}>V</span>
		{/if}
		{#if label}
			{label}
		{/if}
	</legend>
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
		border: 1px solid #997447;
	}

	legend {
		display: flex;
		justify-content: center;
		border-radius: 30px;
		margin-top: 3px;
		padding: 3px;
		background: #997447;
		text-align: center;
	}

	span {
		cursor: pointer;
		text-decoration: line-through;
		user-select: none;
		border: 1px solid black;
		padding: 0 3px;
		margin: 0 3px;
	}

	span.active {
		text-decoration: none;
	}
</style>
