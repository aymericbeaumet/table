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
	<span class="menutoggle">[&gt;]</span>
	<h1>{name}</h1>
	<span>[copy]</span>
</header>

<main />

<style>
	header {
		border-bottom: 1px solid black;
	}

	h1 {
		display: inline;
	}

	.menutoggle {
		cursor: pointer;
	}
</style>
