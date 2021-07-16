<script>
	import { onMount } from 'svelte';
	import Room from '$lib/components/Room.svelte';

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

<Room {name} />
