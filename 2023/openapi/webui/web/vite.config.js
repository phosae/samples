import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    port: 5173,
    proxy: {
      // http://localhost:5173/view -> http://localhost:8000/view
      "/view": "http://localhost:8000",
      "/uis": "http://localhost:8000",
      "/apifiles": "http://localhost:8000",
    }
  },
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})
