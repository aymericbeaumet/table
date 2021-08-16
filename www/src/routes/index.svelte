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
</Container>

<style>
	main {
		max-width: 600px;
		padding: 30px;
		margin: 0 auto;
		margin-top: 200px;
		padding: 30px;
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
