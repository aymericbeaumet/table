<script>
	import { createEventDispatcher } from 'svelte';

	export let items;

	const dispatch = createEventDispatcher();

	let contextMenuElement;
	let top = 0;
	let left = 0;
	let visible = false;

	function onContextMenu(event) {
		event.preventDefault();

		let x = event.clientX;
		const clientXDelta = x + contextMenuElement.offsetWidth - document.body.offsetWidth;
		if (clientXDelta > 0) {
			x -= clientXDelta;
		}

		let y = event.clientY;
		const clientYDelta = y + contextMenuElement.offsetHeight - document.body.offsetHeight;
		if (clientYDelta > 0) {
			y -= clientYDelta;
		}

		top = y;
		left = x;
		visible = true;
	}

	function onClick(event) {
		if (event.target.offsetParent === contextMenuElement) {
			dispatch('action', {
				action: event.target.getAttribute('data-action')
			});
		}
		visible = false;
	}
</script>

<div class="wrapper" on:contextmenu={onContextMenu} on:click={onClick}>
	<div
		class="context-menu"
		class:visible
		bind:this={contextMenuElement}
		style="top: {top}px; left: {left}px;"
	>
		{#each items as [action, desc]}
			<div class="item" data-action={action}>{desc}</div>
		{/each}
	</div>

	<slot />
</div>

<style>
	.wrapper {
		width: 100%;
		height: 100%;
	}

	.context-menu {
		position: fixed;
		z-index: 100;
		width: 150px;
		background: #1b1a1a;
		border-radius: 3px;
		display: none;
	}

	.context-menu.visible {
		display: block;
	}

	.context-menu .item {
		padding: 8px 10px;
		font-size: 15px;
		color: #eee;
		cursor: pointer;
	}

	.context-menu .item:hover {
		background: #343434;
	}
</style>
