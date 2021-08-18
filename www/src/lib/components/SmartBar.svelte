<script>
	import { createEventDispatcher, onMount } from 'svelte';
	import Entity from '$lib/entity';
	import { throttle } from 'lodash';
	import { MAX_FPS } from '$lib/env';
	import Modal from '$lib/components/Modal.svelte';

	export let initialValue = '';

	let entities = new Array<Entity>();
	let selectedIndex = 0;
	let inputTextElement: HTMLInputElement;

	const dispatch = createEventDispatcher();

	const onInput: svelte.JSX.FormEventHandler<HTMLInputElement> = (event) => {
		const value = event.currentTarget?.value.trim() || '';
		entities = Entity.infer(value);
		selectedIndex = 0;
	};

	const onKeyDown: svelte.JSX.KeyboardEventHandler<HTMLInputElement> = (event) => {
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
	};

	const onMouseMove: svelte.JSX.MouseEventHandler<HTMLUListElement> = (event) => {
		const target = event.target as HTMLElement;
		if (target?.nodeName === 'LI') {
			const i = Array.prototype.indexOf.call(target.parentNode?.children, event.target);
			if (i > -1) {
				selectedIndex = i;
			}
		}
	};

	const onSubmit: svelte.JSX.FormEventHandler<HTMLFormElement> = (event) => {
		event.preventDefault();
		dispatch('entity', entities[selectedIndex]);
		dispatch('close');
	};

	const onClick: svelte.JSX.MouseEventHandler<HTMLDivElement> = (event) => {
		const target = event.target as HTMLElement;
		if (target?.nodeName === 'LI') {
			const i = Array.prototype.indexOf.call(target.parentNode?.children, event.target);
			if (i > -1) {
				dispatch('entity', entities[selectedIndex]);
				dispatch('close');
			}
		}
	};

	onMount(() => {
		inputTextElement.value = initialValue;
		entities = Entity.infer(initialValue);
		setTimeout(() => inputTextElement.focus(), 0); // https://github.com/sveltejs/sapper/issues/619#issuecomment-480616597
	});
</script>

<Modal on:close>
	<div on:click={onClick}>
		<form on:submit={onSubmit}>
			<input type="text" on:input={onInput} on:keydown={onKeyDown} bind:this={inputTextElement} />

			<ul on:mousemove={throttle(onMouseMove, 1000 / MAX_FPS)}>
				{#each entities as entity, i}
					<li class:selected={i === selectedIndex}>
						<img src={entity.iconURL} alt={entity.kind} />
						<p>{entity.toString()}</p>
					</li>
				{/each}
			</ul>

			<footer>
				<p>Embed content within the page</p>
			</footer>
		</form>
	</div>
</Modal>

<style>
	div {
		margin: 20px auto 0;
		width: 50%;
	}

	form {
		margin-top: 10px;
		min-width: 700px;
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
		display: flex;
		align-items: center;
		width: 100%;
		padding: 15px;
	}

	li.selected {
		background-color: rgb(18, 100, 163);
	}

	li img {
		width: 32px;
		height: 32px;
	}

	li p {
		margin-left: 10px;
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
