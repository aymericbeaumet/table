import { writable } from 'svelte/store';

// Browser settings

export const audioInputDeviceID = writable(undefined as string | undefined);

export const videoInputDeviceID = writable(undefined as string | undefined);

// Table settings

export const ambianceID = writable('office');

export const ambianceSoundVolume = writable(20); // [0; 100]

export const ambianceEffectsVolume = writable(20); // [0; 100]
