import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  // The locale modules and the app must configure the same Moment instance.
  resolve: { alias: [{ find: /^moment$/, replacement: 'moment/moment.js' }] },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
        ws: true,
        timeout: 240000,
        proxyTimeout: 240000,
      },
      '/healthz': 'http://localhost:8090',
    },
  },
})
