<script>
	import { createEventDispatcher } from 'svelte';
	import { inferEntities } from '$lib/entity';

	let inputTextElement;
	let visible = false;
	let entities = [];

	const dispatch = createEventDispatcher();

	function onClick(event) {
		if (event.target === inputTextElement) {
			return;
		}
		hide();
	}

	function onSubmit(event) {
		event.preventDefault();
		dispatch('entity', entities[0]);
		hide();
	}

	function onInput(event) {
		const value = event.target.value.trim();
		if (value.length > 0) {
			entities = inferEntities(value);
			return;
		}
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
		entities = [];
		setTimeout(() => inputTextElement.focus(), 0); // https://github.com/sveltejs/sapper/issues/619#issuecomment-480616597
		if (!visible) {
			/*inputTextValue = '';*/
			visible = true;
		}
	}

	function hide() {
		entities = [];
		visible = false;
	}

	function toggle() {
		!visible ? show() : hide();
	}
</script>

<svelte:window on:keydown={onWindowKeyDown} />

<div class="wrapper" on:click={onClick}>
	<div class="overlay" class:visible>
		<form on:submit={onSubmit}>
			<input type="text" on:input={onInput} on:keypress={onKeyPress} bind:this={inputTextElement} />
		</form>
		<ul>
			{#each entities as [_, props]}
				<li>{JSON.stringify(props)}</li>
			{/each}
		</ul>
	</div>

	<slot />
</div>

<style>
	.wrapper {
		width: 100%;
		height: 100%;
	}

	.overlay {
		position: fixed;
		z-index: 1000;
		width: 100%;
		height: 100%;
		background: rgb(1, 1, 1, 0.5);
		display: none;
		justify-content: center;
		padding-top: 20%;
	}

	.overlay.visible {
		display: flex;
	}

	form {
		width: 300px;
	}

	input {
		width: 100%;
	}
</style>
