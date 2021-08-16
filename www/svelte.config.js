import adapter from '@sveltejs/adapter-static';
import preprocess from 'svelte-preprocess';
import { mdsvex } from 'mdsvex';

export default {
	kit: {
		target: '#svelte',
		adapter: adapter(),
	},

	extensions: ['.svelte', '.md'],

	preprocess: [
		preprocess({
			defaults: { script: 'typescript' },
		}),
		mdsvex({
			extensions: ['.md'],
		}),
	],
};
