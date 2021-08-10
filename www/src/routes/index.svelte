<script>
	import { goto, prefetchRoutes } from '$app/navigation';
	import { onMount } from 'svelte';

	let tableName: string;

	const onFormSubmit: svelte.JSX.FormEventHandler<HTMLFormElement> = (event) => {
		event.preventDefault();
		goto(`/${tableName}`);
	};

	onMount(prefetchRoutes);
</script>

<svelte:head>
	<title>Ephemeral Discussions - Table</title>
</svelte:head>

<h1>Home page</h1>

<form on:submit={onFormSubmit}>
	<input type="text" placeholder="Pick a table" bind:value={tableName} required minlength="3" />
	<input type="submit" value="Join" />
</form>
