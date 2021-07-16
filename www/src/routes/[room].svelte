<script>
	import { onMount } from 'svelte';

	let name = '';

	function normalize(name) {
		return name.trim().replace(/\s+/g, ' ');
	}

	onMount(() => {
		name = normalize(decodeURIComponent(window.location.pathname.substring(1)));
		window.history.replaceState(null, name, `/${name}`);
	});
</script>

<svelte:head>
	<title>{name}</title>
</svelte:head>

<header>
	<div>
		<h1>{name}</h1>
		<span>[copy]</span>
	</div>

	<div>
		<span>[settings]</span>
	</div>
</header>

<main />

<style>
	header {
		border-bottom: 1px solid black;
		display: flex;
		justify-content: space-between;
		padding: 10px;
	}

	h1 {
		display: inline;
	}
</style>
