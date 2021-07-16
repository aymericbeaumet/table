<script>
	import { goto, prefetchRoutes } from '$app/navigation';
	import { onMount } from 'svelte';

	let name;

	function handleOnSubmit(event) {
		event.preventDefault();
		goto(`/${name}`);
	}

	onMount(prefetchRoutes);
</script>

<h1>Home page</h1>

<form on:submit={handleOnSubmit}>
	<input type="text" placeholder="Name your room" autofocus bind:value={name} min="3" />
	<input type="submit" value="Join" />
</form>
