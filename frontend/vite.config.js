import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

const backendPort = process.env.BACKEND_PORT || '8080';
// Mesma env do frontend: aponta o proxy dev /api para onde o backend roda.
// Sem env, assume o backend local (BACKEND_PORT ou 8080).
const backendUrl = process.env.VITE_BACKEND_URL || `http://localhost:${backendPort}`;

export default defineConfig({
  plugins: [svelte()],
  server: {
    host: '0.0.0.0',
    allowedHosts: true,
    proxy: {
      '/api': backendUrl
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
