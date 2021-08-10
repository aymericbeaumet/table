<script>
	import { onMount } from 'svelte';
	import Table from '$lib/components/Table.svelte';

	let tableName = '';

	function normalize(input: string): string {
		return input.trim().toLowerCase().replace(/\s+/g, ' ');
	}

	onMount(() => {
		tableName = normalize(decodeURIComponent(window.location.pathname.substring(1)));
		window.history.replaceState(null, tableName, `/${tableName}`);
	});
</script>

<svelte:head>
	<title>Table - {tableName}</title>
</svelte:head>

{#if tableName}
	<Table {tableName} />
{/if}
