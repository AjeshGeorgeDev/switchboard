import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const apiPort = process.env.PORT || '8080'
const apiTarget = `http://localhost:${apiPort}`
const distDir = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '../backend/internal/static/dist',
)

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    {
      // emptyOutDir wipes dist/; keep a file so //go:embed dist/* always compiles
      name: 'preserve-embed-placeholder',
      closeBundle() {
        fs.mkdirSync(distDir, { recursive: true })
        fs.writeFileSync(
          path.join(distDir, 'placeholder'),
          'placeholder for go:embed when frontend has not been built yet\n',
        )
      },
    },
  ],
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
