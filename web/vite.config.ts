import { svelte } from '@sveltejs/vite-plugin-svelte'
import { fileURLToPath } from 'url'
import { defineConfig } from 'vite'

export default defineConfig({
  // GitHub Pages serves the project at /<repo>/, so the build needs a base path.
  // Set BASE_PATH=/Nurtura/ there; everywhere else (local dev, Vercel) stays at root.
  base: process.env.BASE_PATH ?? '/',
  plugins: [svelte()],
  resolve: {
    alias: {
      $src: fileURLToPath(new URL('./src/', import.meta.url))
    }
  }
})
