import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

const backendPort = process.env.BACKEND_PORT || '8080';

export default defineConfig({
  plugins: [svelte()],
  server: {
    host: '0.0.0.0',
    allowedHosts: ['oracle.lucasbaccan.com.br'],
    proxy: {
      '/api': `http://localhost:${backendPort}`
    }
  },
  build: {
    outDir: '../backend/web/dist',
    emptyOutDir: true
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test-setup.js'
  }
});
