<script>
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import queryString from 'query-string';
	import Table from '$lib/components/Table.svelte';
	import { API_BASE_URL } from '$lib/env';

	let tableName = '';
	let token = '';

	onMount(() => {
		const pathname = decodeURIComponent(window.location.pathname.substring(1));
		const queryParams = queryString.parse(window.location.search);

		if (typeof queryParams['token'] === 'string' && queryParams['token']) {
			tableName = pathname;
			token = queryParams['token'];
			window.history.replaceState(null, '', window.location.pathname);
		} else {
			const origin = `${window.location.protocol}//${window.location.host}`;
			goto(
				`${API_BASE_URL}/auth/table/${encodeURIComponent(pathname)}?origin=${encodeURIComponent(
					origin,
				)}`,
			);
		}
	});
</script>

<svelte:head>
	<title>Table{tableName ? ` - ${tableName}` : ''}</title>

	<meta name="robots" content="noindex" />
</svelte:head>

{#if token}
	<Table {token} />
{/if}
