import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

export default defineConfig(({ command }) => ({
	base: command === 'serve' ? '/' : '/_app/',
	server: {
		host: '127.0.0.1',
		port: 5173,
		strictPort: true,

		proxy: {
			'/api': {
				target: 'http://127.0.0.1:8080',
				changeOrigin: false
			},

			// Needed when login continues an OAuth request.
			'/oauth': {
				target: 'http://127.0.0.1:8080',
				changeOrigin: false
			}
		}
	},
	plugins: [
		sveltekit({
			preprocess: vitePreprocess(),
			compilerOptions: {
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			adapter: adapter({
				pages: '../../internal/adminui/dist',
				assets: '../../internal/adminui/dist',
				fallback: 'index.html',
				precompress: false
			}),

			paths: {
				base: '',
				assets: ''
			},

			csp: {
				mode: 'hash',

				directives: {
					'default-src': ['self'],
					'script-src': ['self'],
					'style-src': ['self', 'unsafe-inline'],
					'img-src': ['self', 'data:'],
					'connect-src': ['self'],
					'object-src': ['none'],
					'base-uri': ['self'],
					'form-action': ['self']
				}
			}
		})
	]
}));
