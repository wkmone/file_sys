import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 10050,
    proxy: {
      '/api': {
        target: 'http://localhost:10040',
        changeOrigin: true,
      },
    },
  },
})
