import { writable } from 'svelte/store';
import backgroundPNG from '$lib/assets/background.png';

export const audioEnabled = writable(true);

export const videoEnabled = writable(true);

export const backgroundURL = writable(backgroundPNG);
