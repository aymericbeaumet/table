<script>
	import { createEventDispatcher } from 'svelte';
	import { inferEntities } from '$lib/entity';
	import { throttle } from 'lodash';

	let visible = false;
	let entities = [];
	let selectedIndex = 0;
	let inputTextElement;

	const dispatch = createEventDispatcher();

	function onClick(event) {
		if (event.target === inputTextElement) {
			return;
		}

		if (event.target.nodeName === 'LI') {
			const i = Array.prototype.indexOf.call(event.target.parentNode.children, event.target);
			if (i > -1) {
				selectedIndex = i;
				dispatch('entity', entities[selectedIndex]);
				hide();
				return;
			}
		}

		hide();
	}

	const onMouseMove = throttle((event) => {
		if (event.target.nodeName === 'LI') {
			const i = Array.prototype.indexOf.call(event.target.parentNode.children, event.target);
			if (i > -1) {
				selectedIndex = i;
			}
		}
	}, 1000 / 60);

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
			event.preventDefault();
			selectedIndex = (selectedIndex + entities.length - 1) % entities.length || 0;
			return;
		}

		if (event.key === 'ArrowDown') {
			event.preventDefault();
			selectedIndex = (selectedIndex + entities.length + 1) % entities.length || 0;
			return;
		}
	}

	function onKeyPress(event) {
		if (event.ctrlKey && (event.key === 'c' || event.key === 'l' || event.key === '\\')) {
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
		<div>
			<form on:submit={onSubmit}>
				<input
					type="text"
					on:input={onInput}
					on:keydown={onKeyDown}
					on:keypress={onKeyPress}
					bind:this={inputTextElement}
				/>

				<ul on:mousemove={onMouseMove}>
					{#each entities as [kind, _, props], i}
						<li class:selected={i === selectedIndex}>{kind} {JSON.stringify(props)}</li>
					{/each}
				</ul>

				<footer>
					<p>Embed content within the page</p>
				</footer>
			</form>
		</div>
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
		cursor: pointer;
	}

	li {
		display: block;
		width: 100%;
		padding: 15px;
	}

	li.selected {
		background-color: rgb(18, 100, 163);
	}

	footer {
		background-color: rgb(34, 37, 41);
		color: rgb(171, 171, 173);
		text-align: right;
		padding: 15px;
		border-radius: inherit;
		border-top-left-radius: 0;
		border-top-right-radius: 0;
	}
</style>
