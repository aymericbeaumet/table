<script>
	import { prefetchRoutes } from '$app/navigation';
	import { onMount } from 'svelte';

	let tableName: string;

	onMount(prefetchRoutes);
</script>

<svelte:head>
	<title>Ephemeral Discussions - Table</title>
</svelte:head>

<div>
	<header>
		<h1>Table</h1>
	</header>

	<main>
		<form action={`/${tableName}`}>
			<!-- svelte-ignore a11y-autofocus -->
			<input
				type="text"
				placeholder="Pick a table"
				bind:value={tableName}
				required
				minlength="3"
				autofocus={true}
			/>
			<input type="submit" value="Join" />
		</form>
	</main>
</div>

<style>
	:global(body) {
		background: sandybrown;
	}

	div {
		background: lightyellow;
		max-width: 600px;
		padding: 30px;
		margin: 0 auto;
		margin-top: 200px;
	}

	h1 {
		margin-bottom: 10px;
	}

	input[type='text'] {
		width: 300px;
		font-size: 14pt;
		padding: 3px;
	}

	input[type='submit'] {
		font-size: 14pt;
		padding: 3px;
	}
</style>
