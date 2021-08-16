<script>
	import Container from '$lib/components/Container.svelte';

	let tableName = '';
</script>

<svelte:head>
	<title>Ephemeral Discussions - Table</title>
</svelte:head>

<Container>
	<form action={`/${tableName}`}>
		<!-- svelte-ignore a11y-autofocus -->
		<input
			type="text"
			placeholder="Choose a name"
			bind:value={tableName}
			required
			minlength="3"
			autofocus={true}
		/>
		<input type="submit" value="Join table" />
	</form>
</Container>

<style>
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
