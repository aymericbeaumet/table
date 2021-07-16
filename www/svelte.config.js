import adapter from '@sveltejs/adapter-static';

export default {
	kit: {
		target: '#svelte',
		adapter: adapter()
	}
};
