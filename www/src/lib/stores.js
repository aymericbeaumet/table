import { writable } from 'svelte/store';
import backgroundPNG from '$lib/assets/background.png';

// Room settings

export const backgroundURL = writable(backgroundPNG);
