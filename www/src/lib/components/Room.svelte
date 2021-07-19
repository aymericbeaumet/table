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
			<video bind:this={video} muted="muted" />
		</div>
	</main>

	<footer>
		<div>
			<h1 class="clickable" title="Copy room URL" on:click={copyJoiningInfo}>{name}</h1>
		</div>

		<div>
			<button title="Turn off microphone">mic</button>
			<button title="Turn off camera">cam</button>
			<button title="Hang up">hang up</button>
		</div>

		<div>
			<button title="Settings">settings</button>
		</div>
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
		background-color: lightyellow;
	}

	video {
		transform: scale(-1, 1);
		border-radius: 50%;
		height: 150px;
	}

	footer {
		border-top: 1px solid black;
		background-color: lightgrey;
		padding: 10px;
		display: flex;
		justify-content: space-between;
	}

	h1 {
		display: inline-block;
	}

	.clickable:hover {
		cursor: pointer;
	}
</style>
