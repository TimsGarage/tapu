import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

export default defineConfig({
	base: "/admin/",
	plugins: [svelte()],
	build: {
		outDir: "../../internal/adminui/dist",
		emptyOutDir: true,
		sourcemap: false,
	},
});
