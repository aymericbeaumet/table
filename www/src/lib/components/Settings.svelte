<script>
	import { createEventDispatcher } from 'svelte';
	import {
		audioInputDeviceID,
		videoInputDeviceID,
		backgroundImageID,
		backgroundSoundID,
		backgroundSoundVolume,
	} from '$lib/stores';
	import { backgroundImages, backgroundSounds } from '$lib/resources';

	const dispatch = createEventDispatcher();

	const devicesPromise = navigator.mediaDevices.enumerateDevices();

	const onWindowKeyDown = (event: KeyboardEvent) => {
		if (event.key === 'Escape' || (event.ctrlKey && (event.key === 'c' || event.key === '\\'))) {
			dispatch('close');
			return;
		}
	};
</script>

<svelte:window on:keydown={onWindowKeyDown} />

<div>
	<section>
		<h2>Browser settings</h2>

		<h3>Audio & Video</h3>
		<table>
			<tr>
				<td>Microphone</td>
				<td>Camera</td>
			</tr>
			<tr>
				{#await devicesPromise}
					<td colspan="2">loading...</td>
				{:then devices}
					<td>
						<select bind:value={$audioInputDeviceID}>
							{#each devices as device}
								{#if device.kind === 'audioinput'}
									<option value={device.deviceId}>{device.label}</option>
								{/if}
							{/each}
						</select>
					</td>

					<td>
						<select bind:value={$videoInputDeviceID}>
							{#each devices as device}
								{#if device.kind === 'videoinput'}
									<option value={device.deviceId}>{device.label}</option>
								{/if}
							{/each}
						</select>
					</td>
				{/await}
			</tr>
		</table>
	</section>

	<section>
		<h2>Table settings (synced)</h2>

		<h3>Atmosphere</h3>
		{#each Object.entries(backgroundImages) as [id, resource]}
			<label class="bg-image" style="background-image: url({resource.url})">
				<input type="radio" bind:group={$backgroundImageID} value={id} />
			</label>
		{/each}

		<h3>Ambiance</h3>
		{#each Object.entries(backgroundSounds) as [id, resource]}
			<label>
				<input type="radio" bind:group={$backgroundSoundID} value={id} />
				{resource.name}
			</label>
		{/each}

		<h3>Ambiance sound level</h3>

		<label>
			<input type="range" bind:value={$backgroundSoundVolume} min="0" max="100" />
			<input type="number" bind:value={$backgroundSoundVolume} min="0" max="100" />%
		</label>
	</section>

	<footer>
		<button on:click={() => dispatch('close')}>Close</button>
	</footer>
</div>

<style>
	div {
		background-color: white;
		width: 800px;
		padding: 10px;
		border: 1px solid black;
	}

	section {
		margin: 20px 0;
		border: 1px solid grey;
		padding: 10px;
	}

	h2,
	h3 {
		margin-top: 5px;
		margin-bottom: 5px;
	}

	label {
		cursor: pointer;
	}

	label.bg-image {
		display: inline-block;
		border: 1px solid grey;
		width: 64px;
		height: 64px;
		background-position: center;
		background-size: cover;
	}

	label.bg-image + label.bg-image {
		margin-left: 10px;
	}

	input[type='range'] {
		width: 400px;
	}

	button {
		padding: 5px;
	}
</style>
