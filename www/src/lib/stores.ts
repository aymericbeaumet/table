import { writable } from 'svelte/store';

// Browser settings

export const audioInputDeviceID = writable(undefined as string | undefined);

export const videoInputDeviceID = writable(undefined as string | undefined);

// Table settings

export const backgroundImageID = writable('fantasy');

export const backgroundSoundID = writable('medieval');
