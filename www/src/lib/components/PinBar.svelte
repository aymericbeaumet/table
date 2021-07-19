<script>
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher();

	let inputTextElement;
	let inputTextValue;
	let visible = false;

	function onContextMenu(event) {
		// allow to bypass custom logic by holding "option"
		if (event.altKey) {
			return;
		}
		event.preventDefault();
		show();
	}

	function onClick(event) {
		if (event.target === inputTextElement) {
			return;
		}
		hide();
	}

	function onSubmit(event) {
		event.preventDefault();

		const value = inputTextValue.trim();
		if (value.length > 0) {
			dispatch('entity', {
				type: 'text',
				value,
			});
		}

		hide();
	}

	function onKeyPress(event) {
		if (event.ctrlKey && (event.key === 'c' || event.key === '\\')) {
			hide();
			return;
		}
	}

	function onWindowKeyDown(event) {
		if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
			toggle();
			return;
		}

		if (event.key === 'Escape') {
			hide();
			return;
		}
	}

	function show() {
		setTimeout(() => inputTextElement.focus(), 0); // https://github.com/sveltejs/sapper/issues/619#issuecomment-480616597
		if (!visible) {
			inputTextValue = '';
			visible = true;
		}
	}

	function hide() {
		visible = false;
	}

	function toggle() {
		!visible ? show() : hide();
	}
</script>

<svelte:window on:keydown={onWindowKeyDown} />

<div class="wrapper" on:contextmenu={onContextMenu} on:click={onClick}>
	<div class="pin-bar" class:visible>
		<form on:submit={onSubmit}>
			<input
				type="text"
				required
				minlength="1"
				on:keypress={onKeyPress}
				bind:this={inputTextElement}
				bind:value={inputTextValue}
			/>
			<input type="submit" value="Pin" />
		</form>
	</div>

	<slot />
</div>

<style>
	.wrapper {
		width: 100%;
		height: 100%;
	}

	.pin-bar {
		position: fixed;
		z-index: 100;
		width: 100%;
		height: 100%;
		background: rgb(1, 1, 1, 0.5);
		display: none;
		justify-content: center;
		padding-top: 20%;
	}

	.pin-bar.visible {
		display: flex;
	}
</style>
