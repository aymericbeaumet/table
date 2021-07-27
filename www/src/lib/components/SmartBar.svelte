<script>
	import { createEventDispatcher } from 'svelte';
	import { inferEntities } from '$lib/entity';

	let inputTextElement;
	let visible = false;
	let entities = [];
	let selectedIndex = 0;

	const dispatch = createEventDispatcher();

	function onClick(event) {
		if (event.target === inputTextElement) {
			return;
		}
		hide();
	}

	function onSubmit(event) {
		event.preventDefault();
		dispatch('entity', entities[selectedIndex]);
		hide();
	}

	function onInput(event) {
		const value = event.target.value.trim();
		entities = inferEntities(value);
		selectedIndex = 0;
	}

	function onKeyDown(event) {
		if (event.key === 'ArrowUp') {
			selectedIndex = (selectedIndex + entities.length - 1) % entities.length;
		} else if (event.key === 'ArrowDown') {
			selectedIndex = (selectedIndex + entities.length + 1) % entities.length;
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
		if (!visible) {
			entities = [];
			selectedIndex = 0;
			inputTextElement.value = '';
			setTimeout(() => inputTextElement.focus(), 0); // https://github.com/sveltejs/sapper/issues/619#issuecomment-480616597
		}
		visible = true;
	}

	function hide() {
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
			<input
				type="text"
				on:input={onInput}
				on:keydown={onKeyDown}
				on:keypress={onKeyPress}
				bind:this={inputTextElement}
			/>
			<ul>
				{#each entities as [_, props], i}
					<li class:selected={i == selectedIndex}>{JSON.stringify(props)}</li>
				{/each}
			</ul>
		</form>
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
		background: rgb(1, 1, 1, 0.6);
		display: none;
		justify-content: center;
	}

	.overlay.visible {
		display: flex;
	}

	form {
		margin-top: 10px;
		min-width: 700px;
		width: 50%;
		background-color: rgb(26, 29, 33);
		color: rgb(209, 210, 211);
		border: 1px solid grey;
		border-radius: 10px;
	}

	input {
		width: 100%;
		background-color: inherit;
		color: inherit;
		border: none;
		border-bottom: 1px solid grey;
		border-radius: inherit;
		border-bottom-left-radius: 0;
		border-bottom-right-radius: 0;
		padding: 15px;
		font-size: 20px;
	}

	input:focus {
		outline: none;
	}

	ul {
		padding: 0;
	}

	li {
		display: block;
		width: 100%;
		height: 50px;
		line-height: 50px;
	}

	li.selected {
		background-color: blue;
	}
</style>
