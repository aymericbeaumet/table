import { writable } from 'svelte/store';

// Browser settings

export const audioInputDeviceID = writable('default');

export const videoInputDeviceID = writable('default');

// Table settings

export const backgroundImageID = writable('fantasy');

export const backgroundSoundID = writable('medieval');
