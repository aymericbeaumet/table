<script>
	import Container from '$lib/components/Container.svelte';

	let tableName = '';
</script>

<svelte:head>
	<title>Ephemeral Discussions - Table</title>
</svelte:head>

<Container>
	<h1>A table for your <span>1:1s</span></h1>

	<form action={`/${tableName}`}>
		<!-- svelte-ignore a11y-autofocus -->
		<input
			type="text"
			placeholder="Choose a table name"
			bind:value={tableName}
			required
			minlength="3"
			autofocus={true}
		/>
		<input type="submit" value="Join" />
	</form>
</Container>

<style>
	span {
		color: orange;
	}

	input[type='text'] {
		width: 300px;
		font-size: 14pt;
		padding: 3px;
	}

	input[type='submit'] {
		font-size: 14pt;
		padding: 3px 6px;
	}
</style>
