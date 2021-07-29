<script>
	import { onMount } from 'svelte';
	import Room from '$lib/components/Room.svelte';

	let name = '';

	function normalize(name) {
		return name.split('/')[0].trim().replace(/\s+/g, ' ');
	}

	onMount(() => {
		name = normalize(decodeURIComponent(window.location.pathname.substring(1)));
		window.history.replaceState(null, name, `/${name}`);
	});
</script>

<svelte:head>
	<title>Table - {name}</title>
</svelte:head>

{#if name}
	<Room {name} />
{/if}
