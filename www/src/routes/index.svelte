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

	<footer>
		<a
			href="https://slack.com/oauth/v2/authorize?client_id=2367199992290.2367416489235&scope=commands&user_scope="
			><img
				alt="Add to Slack"
				height="40"
				width="139"
				src="https://platform.slack-edge.com/img/add_to_slack.png"
				srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x"
			/></a
		>
	</footer>
</div>

<style>
	:global(body) {
		background: sandybrown;
	}

	div {
		max-width: 600px;
		padding: 30px;
		margin: 0 auto;
		margin-top: 200px;
	}

	header {
		background: lightcyan;
		padding: 30px;
	}

	main {
		background: lightyellow;
		padding: 30px;
	}

	footer {
		background: lightgreen;
		padding: 30px;
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
