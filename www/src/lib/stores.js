import { writable } from 'svelte/store';
import * as resources from '$lib/resources';

// Room settings

export const backgroundImage = writable(resources.backgroundImages.fantasy);

export const backgroundSound = writable(resources.backgroundSounds.medieval);
