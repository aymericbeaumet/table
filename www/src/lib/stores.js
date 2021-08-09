import { writable } from 'svelte/store';
import * as resources from '$lib/resources';

// Local settings

export const audioInputDeviceID = writable(undefined);
export const audioOutputDeviceID = writable(undefined);
export const videoInputDeviceID = writable(undefined);

// Room settings

export const backgroundImage = writable(resources.backgroundImages.find((r) => r.id === 'fantasy'));
export const backgroundSound = writable(
	resources.backgroundSounds.find((r) => r.id === 'medieval'),
);
