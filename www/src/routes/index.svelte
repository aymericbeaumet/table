<script>
	import { onMount } from 'svelte';
	import Container from '$lib/components/Container.svelte';

	let tableName = '';
	let inputTextElement: HTMLInputElement;

	onMount(() => {
		setTimeout(() => inputTextElement.focus(), 0); // https://github.com/sveltejs/sapper/issues/619#issuecomment-480616597
	});
</script>

<svelte:head>
	<title>Ephemeral Discussions - Table</title>
</svelte:head>

<Container>
	<h1>A table for your <span>1:1s</span></h1>

	<form action={`/${tableName}`}>
		<input
			type="text"
			placeholder="Choose a table name"
			bind:this={inputTextElement}
			bind:value={tableName}
			required
			minlength="3"
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
