if (!import.meta.env.VITE_API_BASE_URL) {
	throw new Error('missing VITE_API_BASE_URL');
}
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

export const MAX_FPS = parseInt(import.meta.env.VITE_MAX_FPS || '30', 10);

if (!import.meta.env.VITE_OUTPOST_BASE_URL) {
	throw new Error('missing VITE_OUTPOST_BASE_URL');
}
export const OUTPOST_BASE_URL = import.meta.env.VITE_OUTPOST_BASE_URL as string;

if (!import.meta.env.VITE_STUN_SERVERS_URLS) {
	throw new Error('missing VITE_STUN_SERVERS_URLS');
}
export const STUN_SERVERS_URLS = import.meta.env.VITE_STUN_SERVERS_URLS as string;
