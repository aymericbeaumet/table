<script>
	import Entity from '$lib/entity';
	import EntityCursor from '$lib/components/EntityCursor.svelte';
	import EntityIframe from '$lib/components/EntityIframe.svelte';
	import EntityText from '$lib/components/EntityText.svelte';

	export let entity: Entity;
</script>

{#if entity.kind === 'cursor'}
	<EntityCursor {...entity.props} />
{:else if entity.kind === 'iframe'}
	<EntityIframe {...entity.props} />
{:else if entity.kind === 'text'}
	<EntityText {...entity.props} />
{/if}
