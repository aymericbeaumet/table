<script>
	import Network from '$lib/network';
	import { onMount } from 'svelte';
	import Room from '$lib/components/Room.svelte';

	let name = '';

	function normalize(name) {
		return name.split('/')[0].trim().replace(/\s+/g, ' ');
	}

	onMount(() => {
		name = normalize(decodeURIComponent(window.location.pathname.substring(1)));
		window.history.replaceState(null, name, `/${name}`);

		const network = new Network(name);
		network.start();
	});
</script>

<svelte:head>
	<title>Chill - {name}</title>
</svelte:head>

<Room {name} />
