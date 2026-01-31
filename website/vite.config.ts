import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import path from "path";

export default defineConfig({
  resolve: {
    alias: {
      "@lib": path.resolve(__dirname, "./src/lib"),
      "@components": path.resolve(__dirname, "./src/components"),
      "@routes": path.resolve(__dirname, "./src/routes"),
    },
  },
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('@mui/x-charts') || id.includes('@mui/x-data-grid')) {
              return 'mui-x';
            }
            if (id.includes('@mui') || id.includes('@emotion') || id.includes('react') || id.includes('scheduler') || id.includes('react-router')) {
              return 'vendor-core';
            }
          }
        },
      },
    },
  },
})
