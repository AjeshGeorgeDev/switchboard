import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

const apiPort = process.env.PORT || '8080'
const apiTarget = `http://localhost:${apiPort}`

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': apiTarget,
      '/webhooks': apiTarget,
    },
  },
  build: {
    outDir: '../backend/internal/static/dist',
    emptyOutDir: true,
  },
})
