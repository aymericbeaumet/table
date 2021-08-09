<script>
	import { createEventDispatcher } from 'svelte';
	import {
		audioInputDeviceID,
		audioOutputDeviceID,
		videoInputDeviceID,
		backgroundImage,
		backgroundSound,
	} from '$lib/stores';
	import { backgroundImages, backgroundSounds } from '$lib/resources';

	const dispatch = createEventDispatcher();

	const devicesPromise = navigator.mediaDevices.enumerateDevices();
</script>

<div>
	<h2>Local Settings</h2>

	<section>
		<h3>Audio & Video</h3>
		<table>
			<tr>
				<td>Microphone</td>
				<td>Speakers</td>
				<td>Camera</td>
			</tr>
			<tr>
				{#await devicesPromise}
					<td colspan="3">loading...</td>
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
						<select bind:value={$audioOutputDeviceID}>
							{#each devices as device}
								{#if device.kind === 'audiooutput'}
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

	<h2>Room Settings</h2>

	<section>
		<h3>Background image</h3>
		{#each backgroundImages as resource}
			<label>
				<input type="radio" bind:group={$backgroundImage} value={resource} />
				<img src={resource.url} alt={resource.name} />
			</label>
		{/each}
	</section>

	<section>
		<h3>Background sound</h3>
		{#each backgroundSounds as resource}
			<label>
				<input type="radio" bind:group={$backgroundSound} value={resource} />
				{resource.name}
			</label>
		{/each}
	</section>

	<footer>
		<button on:click={() => dispatch('close')}>Close</button>
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
