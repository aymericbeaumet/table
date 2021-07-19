<script>
	import { onMount } from 'svelte';

	export let name;

	function copyJoiningInfo() {
		navigator.clipboard.writeText(window.location.toString());
	}

	function displayQrcode() {
		console.log('todo');
	}

	let srcObject;
	let video;

	$: if (video && srcObject && video.srcObject !== srcObject) {
		video.srcObject = srcObject;
		video.play();
	}

	onMount(async () => {
		const stream = await navigator.mediaDevices.getUserMedia({
			audio: true,
			video: {
				width: { ideal: 360, max: 360 },
				height: { ideal: 360, max: 360 },
				facingMode: { ideal: 'user' },
				frameRate: { ideal: 30, max: 30 }
			}
		});
		srcObject = stream;
	});
</script>

<div class="container">
	<main>
		<div>
			<video class="primary" bind:this={video} muted="muted" />
		</div>
	</main>

	<footer>
		<div>
			<h1>{name}</h1>
			<div class="toggle" on:click={copyJoiningInfo}>copy</div>
			<div class="toggle" on:click={displayQrcode}>qrcode</div>
		</div>

		<div>
			<div class="toggle">audio</div>
			<div class="toggle">video</div>
		</div>

		<div />
	</footer>
</div>

<style>
	.container {
		height: 100vh;
		display: flex;
		flex-direction: column;
	}

	main {
		height: 100%;
		background-color: red;
	}

	video {
		transform: scale(-1, 1);
		border-radius: 50%;
		height: 150px;
	}

	footer {
		border-top: 1px solid black;
		display: flex;
		justify-content: space-between;
		padding: 10px;
	}

	.toggle {
		border: 1px solid black;
		border-radius: 50%;
		display: inline-block;
		width: 48px;
		height: 48px;
	}

	h1 {
		display: inline;
	}
</style>
