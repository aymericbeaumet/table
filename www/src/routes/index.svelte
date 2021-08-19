<script>
	import { onMount, onDestroy } from 'svelte';
	import Container from '$lib/components/Container.svelte';

	const goals = [
		'meetings',
		'all hands',
		'1:1s',
		'daily standups',
		'code reviews',
		'coffee breaks',
		'roleplay nights',
	];

	let goalsIndex = 0;
	let tableName = '';

	let interval: ReturnType<typeof setTimeout>;
	let inputTextElement: HTMLInputElement;

	onMount(() => {
		setTimeout(() => inputTextElement.focus(), 0); // https://github.com/sveltejs/sapper/issues/619#issuecomment-480616597
		interval = setInterval(() => (goalsIndex = (goalsIndex + 1) % goals.length), 3000);
	});

	onDestroy(() => {
		clearInterval(interval);
	});
</script>

<svelte:head>
	<title>Ephemeral Discussions - Table</title>
</svelte:head>

<Container>
	<section>
		<h1>A place for your <span>{goals[goalsIndex]}</span></h1>

		<form action={`/${tableName || 'table'}`}>
			<input
				type="text"
				placeholder="Choose a table name"
				bind:this={inputTextElement}
				bind:value={tableName}
			/>
			<input type="submit" value="Join" />
		</form>
	</section>
</Container>

<style>
	section {
		width: 400px;
		margin: 0 auto;
	}

	h1 {
		margin-top: 80px;
		margin-bottom: 25px;
	}

	span {
		color: orange;
	}

	form {
		display: flex;
	}

	input[type='text'] {
		flex: 1;
		font-size: 14pt;
		padding: 3px;
	}

	input[type='submit'] {
		font-size: 14pt;
		padding: 3px 6px;
	}
</style>
