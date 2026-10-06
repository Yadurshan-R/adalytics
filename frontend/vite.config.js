import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// In development, requests to /api go to the Go backend on port 8080.
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
