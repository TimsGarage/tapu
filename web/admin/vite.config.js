import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

export default defineConfig(({ command }) => ({
	base: command === "serve" ? "/" : "/ui/",
	plugins: [svelte()],
	server: {
		host: "127.0.0.1",
		port: 5173,
		strictPort: true,

		proxy: {
			"/api": {
				target: "http://127.0.0.1:8080",
				changeOrigin: false,
			},

			// Needed when login continues an OAuth request.
			"/oauth": {
				target: "http://127.0.0.1:8080",
				changeOrigin: false,
			},
		},
	},
	build: {
		outDir: "../../internal/adminui/dist",
		emptyOutDir: true,
		sourcemap: false,
	},
}));
