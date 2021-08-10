interface ImportMeta {
	env: {
		VITE_MAX_FPS?: string;
		VITE_OUTPOST_BASE_URL?: string;
		VITE_STUN_SERVERS_URLS?: string;
	};
}

declare module 'svelte-favicon-badge';

// JPEG, OGG, MP3, PNG, SVG files are imported as urls (string)
declare module '*.jpeg' {
	const content: string;
	export default content;
}
declare module '*.mp3' {
	const content: string;
	export default content;
}
declare module '*.ogg' {
	const content: string;
	export default content;
}
declare module '*.png' {
	const content: string;
	export default content;
}
declare module '*.svg' {
	const content: string;
	export default content;
}
