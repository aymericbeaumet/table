// We have to re-export from here to avoid a build problem.
// https://timdeschryver.dev/blog/environment-variables-with-sveltekit#the-problem

export const MAX_FPS = import.meta.env.MAX_FPS || 30;
export const OUTPOST_BASE_URL = import.meta.env.VITE_OUTPOST_BASE_URL;
export const STUN_SERVERS_URLS = import.meta.env.VITE_STUN_SERVERS_URLS;
