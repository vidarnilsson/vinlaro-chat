import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// When running inside Docker, BACKEND_HOST is set to host.docker.internal so
// the Vite proxy can reach the Go API running on the host machine.
const backendHost = process.env.BACKEND_HOST ?? 'localhost'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': `http://${backendHost}:8000`,
      '/ws': {
        target: `ws://${backendHost}:8000`,
        ws: true,
      },
    },
  },
})
