<script>
	import { createEventDispatcher } from 'svelte';
	import { backgroundImage, backgroundSound } from '$lib/stores';
	import { backgroundImages, backgroundSounds } from '$lib/resources';

	const dispatch = createEventDispatcher();
</script>

<div>
	<header>
		<h2>settings</h2>
	</header>

	<section>
		<h3>Background image</h3>
		{#each Object.values(backgroundImages) as resource}
			<label>
				<input type="radio" bind:group={$backgroundImage} value={resource} />
				<img src={resource.url} alt={resource.name} />
			</label>
		{/each}
	</section>

	<section>
		<h3>Background sound</h3>
		{#each Object.values(backgroundSounds) as resource}
			<label>
				<input type="radio" bind:group={$backgroundSound} value={resource} />
				{resource.name}
			</label>
		{/each}
	</section>

	<footer>
		<button on:click={() => dispatch('close')}>Close</button>
	</footer>
</div>

<style>
	div {
		margin: 0 auto;
		margin-top: 100px;
		background-color: white;
		width: 800px;
		border: 1px solid black;
	}

	label {
		cursor: pointer;
	}

	label > img {
		border: 1px solid grey;
		width: 64px;
		height: 64px;
	}
</style>
