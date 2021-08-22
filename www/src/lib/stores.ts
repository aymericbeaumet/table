import { writable } from 'svelte/store';
import { AmbianceID } from '$lib/assets/ambiances';

// Browser settings

export const audioInputDeviceID = writable(undefined as string | undefined);

export const videoInputDeviceID = writable(undefined as string | undefined);

// Table settings

export const ambianceID = writable(AmbianceID.COFFEE);

export const ambianceSoundVolume = writable(20); // [0; 100]

export const ambianceEffectsVolume = writable(20); // [0; 100]
