import { writable } from 'svelte/store';

// Browser settings

export const audioInputDeviceID = writable(undefined as string | undefined);

export const videoInputDeviceID = writable(undefined as string | undefined);

// Table settings

export const backgroundImageID = writable('desk');

export const backgroundSoundID = writable('silence');

export const backgroundSoundVolume = writable(25); // [0; 100]
