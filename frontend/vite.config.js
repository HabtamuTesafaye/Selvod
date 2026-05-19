import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  // Load env vars from frontend/.env (and .env.local overrides)
  const env = loadEnv(mode, process.cwd(), '')

  const apiTarget = env.VITE_API_TARGET || 'http://localhost:8081'
  const hlsTarget = env.VITE_HLS_TARGET || 'https://localhost:18443'

  return {
    plugins: [vue()],
    server: {
      proxy: {
        // /api and /health go directly to the Go backend (port from frontend/.env VITE_API_TARGET)
        '/api': {
          target: apiTarget,
          changeOrigin: true,
        },
        '/health': {
          target: apiTarget,
          changeOrigin: true,
        },
        // /hls goes through Nginx HTTPS (port from frontend/.env VITE_HLS_TARGET)
        // secure: false allows self-signed local dev certificates
        '/hls': {
          target: hlsTarget,
          changeOrigin: true,
          secure: false,
        }
      }
    }
  }
})

