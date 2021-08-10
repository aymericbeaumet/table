<script>
	import { createEventDispatcher } from 'svelte';
	import {
		audioInputDeviceID,
		videoInputDeviceID,
		backgroundImageID,
		backgroundSoundID,
	} from '$lib/stores';
	import { backgroundImages, backgroundSounds } from '$lib/resources';

	const dispatch = createEventDispatcher();

	const onCloseButtonClick: svelte.JSX.MouseEventHandler<HTMLButtonElement> = () => {
		dispatch('close');
	};

	const devicesPromise = navigator.mediaDevices.enumerateDevices();
</script>

<div>
	<h2>Browser Settings</h2>

	<section>
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

	<h2>Table Appearance (synced)</h2>

	<section>
		<h3>Background image</h3>
		{#each Object.entries(backgroundImages) as [id, resource]}
			<label>
				<input type="radio" bind:group={$backgroundImageID} value={id} />
				<img src={resource.url} alt={resource.name} />
			</label>
		{/each}
	</section>

	<section>
		<h3>Background sound</h3>
		{#each Object.entries(backgroundSounds) as [id, resource]}
			<label>
				<input type="radio" bind:group={$backgroundSoundID} value={id} />
				{resource.name}
			</label>
		{/each}
	</section>

	<footer>
		<button on:click={onCloseButtonClick}>Close</button>
	</footer>
</div>

<style>
	div {
		margin: 0 auto;
		margin-top: 100px;
		background-color: white;
		width: 800px;
		border: 1px solid black;
	}

	label {
		cursor: pointer;
	}

	label > img {
		border: 1px solid grey;
		width: 64px;
		height: 64px;
		overflow: hidden;
	}
</style>
