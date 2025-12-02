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
            const normalizedId = id.split(path.sep).join('/');
            switch (true) {
              case normalizedId.includes('@mui/icons-material'):
                return 'mui-icons';
              case normalizedId.includes('@mui'):
              case normalizedId.includes('@emotion'):
                return 'mui-vendor';
              case normalizedId.includes('react'):
              case normalizedId.includes('scheduler'):
              case normalizedId.includes('remix-run'):
                return 'react-vendor';
              case normalizedId.includes('zod'):
                return 'utils';
              default:
                return 'vendor';
            }
          }
        },
      },
    },
  },
})
