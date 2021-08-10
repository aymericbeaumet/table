module.exports = {
	root: true,
	parser: '@typescript-eslint/parser',
	parserOptions: {
		sourceType: 'module',
	},
	extends: ['eslint:recommended', 'plugin:@typescript-eslint/recommended', 'prettier'],
	plugins: ['svelte3', '@typescript-eslint'],
	overrides: [{ files: ['*.svelte'], processor: 'svelte3/svelte3' }],
	settings: {
		'svelte3/typescript': require('typescript'),
	},
	env: {
		browser: true,
		es6: true,
	},
	// TODO: can we somehow avoid to define those globals here?
	globals: {
		svelte: true,
		MediaProvider: true,
	},
};
