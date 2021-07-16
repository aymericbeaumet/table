import adapter from '@sveltejs/adapter-static';

export default {
	kit: {
		target: '#svelte',
		adapter: adapter(),
		vite: {
			define: {
				'process.env.WEBSOCKET_BASE_URL': 'process.env.WEBSOCKET_BASE_URL'
			}
		}
	}
};
