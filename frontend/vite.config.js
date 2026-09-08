import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { execSync } from 'node:child_process';
import { readFileSync } from 'node:fs';

const backendPort = process.env.BACKEND_PORT || '8080';
// Mesma env do frontend: aponta o proxy dev /api para onde o backend roda.
// Sem env, assume o backend local (BACKEND_PORT ou 8080).
const backendUrl = process.env.VITE_BACKEND_URL || `http://localhost:${backendPort}`;

// Versão + commit embutidos no build (exibidos no "Como funciona" e no popover
// de ajuda). Lê o version do package.json e o hash curto do git; sem repo
// (Docker/Vercel sem .git), cai em '0.0.0-dev' / 'dev'.
// No build do Docker o .git não vai pra dentro da imagem, então o commit é
// injetado via build-arg GIT_COMMIT -> env VITE_APP_COMMIT (ver Dockerfile /
// docker-compose.yml / deploy.yml).
const pkg = JSON.parse(readFileSync(new URL('./package.json', import.meta.url), 'utf-8'));
let appCommit = process.env.VITE_APP_COMMIT;
if (appCommit === undefined) {
  try {
    appCommit = execSync('git rev-parse --short HEAD', { encoding: 'utf-8' }).trim();
  } catch {
    // sem .git / outro ambiente
    appCommit = 'dev';
  }
}
if (!appCommit) appCommit = 'dev';

export default defineConfig({
  plugins: [svelte()],
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version || '0.0.0-dev'),
    __APP_COMMIT__: JSON.stringify(appCommit || 'dev')
  },
  server: {
    host: '0.0.0.0',
    allowedHosts: true,
    proxy: {
      '/api': backendUrl
    }
  },
  build: {
    // Na Vercel (env VERCEL=1) o dist fica dentro do frontend; localmente vai
    // para ../backend/web/dist para o binário Go embutir via //go:embed.
    outDir: process.env.VERCEL ? 'dist' : '../backend/web/dist',
    emptyOutDir: true
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test-setup.js'
  }
});
