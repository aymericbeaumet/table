<script>
	import {
		audioInputDeviceID,
		videoInputDeviceID,
		ambianceID,
		ambianceSoundVolume,
		ambianceEffectsVolume,
	} from '$lib/stores';
	import ambiances from '$lib/assets/ambiances';
	import Modal from '$lib/components/Modal.svelte';

	const devicesPromise = navigator.mediaDevices.enumerateDevices();
</script>

<Modal on:close>
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

			<h3>Ambiance</h3>

			{#each Object.entries(ambiances) as [id, ambiance]}
				<label
					class="ambiance"
					style="background-image: url({ambiance.backgroundURL})"
					title={ambiance.name}
					><input type="radio" bind:group={$ambianceID} value={id} />{ambiance.name}</label
				>
			{/each}

			<h3>Ambiance sound volume</h3>

			<label>
				<input type="range" bind:value={$ambianceSoundVolume} min="0" max="100" />
				<input type="number" bind:value={$ambianceSoundVolume} min="0" max="100" />%
			</label>

			<h3>Ambiance effect volume</h3>

			<label>
				<input type="range" bind:value={$ambianceEffectsVolume} min="0" max="100" />
				<input type="number" bind:value={$ambianceEffectsVolume} min="0" max="100" />%
			</label>
		</section>
	</div>
</Modal>

<style>
	div {
		margin: 150px auto 0;
		background-color: white;
		width: 90%;
		max-width: 800px;
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
		color: black;
	}

	label.ambiance {
		display: inline-block;
		border: 1px solid grey;
		width: 90px;
		height: 90px;
		background-position: center;
		background-size: cover;
	}

	label.ambiance + label {
		margin-left: 10px;
	}

	input[type='range'] {
		width: 400px;
	}
</style>
