import adapterStatic from '@sveltejs/adapter-static';
import sveltePreprocess from 'svelte-preprocess';

export default {
	kit: {
		target: '#svelte',
		adapter: adapterStatic(),
	},

	preprocess: sveltePreprocess({
		defaults: {
			script: 'typescript',
		},
	}),
};
