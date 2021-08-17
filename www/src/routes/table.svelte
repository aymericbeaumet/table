<script>
	import { onMount } from 'svelte';
	import Table from '$lib/components/Table.svelte';
	import { hri } from 'human-readable-ids';

	let tableName = '';

	function normalize(input: string): string {
		return input.trim().toLowerCase().replace(/\s+/g, ' ') || '';
	}

	onMount(() => {
		tableName = normalize(decodeURIComponent(window.location.pathname.substring(1)));
		if (!tableName || tableName === 'table') {
			tableName = hri.random();
		}
		window.history.replaceState(null, tableName, `/${tableName}`);
	});
</script>

<svelte:head>
	<title>Table - {tableName}</title>

	<meta name="robots" content="noindex" />
</svelte:head>

{#if tableName}
	<Table {tableName} />
{/if}
