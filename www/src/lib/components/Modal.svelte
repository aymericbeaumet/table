<script>
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher();

	const onKeyDown = (event: KeyboardEvent) => {
		if (event.key === 'Escape' || (event.ctrlKey && (event.key === 'c' || event.key === '\\'))) {
			event.preventDefault();
			event.stopPropagation();
			dispatch('close');
			return;
		}
	};

	const onClick = (event: MouseEvent) => {
		if (event.currentTarget === event.target) {
			event.preventDefault();
			event.stopPropagation();
			dispatch('close');
		}
	};
</script>

<svelte:window on:keydown={onKeyDown} />

<div on:click={onClick}>
	<slot />
</div>

<style>
	div {
		z-index: 1000;
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgb(1, 1, 1, 0.5);
	}
</style>
