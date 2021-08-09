import { writable } from 'svelte/store';
import * as resources from '$lib/resources';

// Local settings

export const audioInputDeviceID = writable('default');
export const videoInputDeviceID = writable('default');

// Room settings

export const backgroundImage = writable(resources.backgroundImages.find((r) => r.id === 'fantasy'));
export const backgroundSound = writable(
	resources.backgroundSounds.find((r) => r.id === 'medieval'),
);
