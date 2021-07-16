<script>
	import { onMount } from 'svelte';

	let name = '';

	onMount(() => {
		name = normalize(decodeURIComponent(window.location.pathname.substring(1)));
		window.history.replaceState(null, name, `/${name}`);
	});

	function normalize(name) {
		return name.trim().replace(/\s+/g, ' ');
	}
</script>

<svelte:head>
	<title>{name}</title>
</svelte:head>

<h1>{name}</h1>
