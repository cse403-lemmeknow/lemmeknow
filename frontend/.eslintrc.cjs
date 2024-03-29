module.exports = {
	root: true,
	extends: ['eslint:recommended', 'prettier', 'plugin:svelte/prettier'],
	overrides: [{ files: ['*.svelte'] }],
	rules: {
		'no-unused-vars': [
			'error',
			{
				varsIgnorePattern: '^_',
				argsIgnorePattern: '^_'
			}
		]
	},
	parserOptions: {
		sourceType: 'module',
		ecmaVersion: 2020
	},
	env: {
		browser: true,
		es2017: true,
		node: true
	}
};
