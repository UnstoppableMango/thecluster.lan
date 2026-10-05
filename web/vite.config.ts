import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  // The Go server reads dist/.vite/manifest.json to find the entry chunk.
  build: { manifest: true },
  server: {
    host: true,
    proxy: { "/ping": "http://localhost:8080" },
  },
});
