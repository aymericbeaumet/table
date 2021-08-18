<script>
	import { onMount } from 'svelte';
	import Table from '$lib/components/Table.svelte';
	import { hri } from 'human-readable-ids';

	let tableName = '';

	onMount(() => {
		let normalized = normalize(decodeURIComponent(window.location.pathname.substring(1)));
		if (!normalized || ['new', 'rand', 'random', 'start', 'create', 'table'].includes(normalized)) {
			normalized = hri.random();
		}
		tableName = normalized;
		window.history.replaceState(null, tableName, `/${tableName}`);
	});

	function normalize(input: string): string {
		return input.trim().toLowerCase().replace(/\s+/g, ' ');
	}
</script>

<svelte:head>
	<title>Table{tableName ? ` - ${tableName}` : ''}</title>

	<meta name="robots" content="noindex" />
</svelte:head>

{#if tableName}
	<Table {tableName} />
{/if}
