import { writable } from 'svelte/store';
import backgroundPNG from '$lib/assets/background.png';

// User settings

export const audioEnabled = writable(true);

export const videoEnabled = writable(true);

// Room settings

export const backgroundURL = writable(backgroundPNG);
