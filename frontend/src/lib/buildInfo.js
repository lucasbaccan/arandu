// Info de build do frontend, injetada pelo Vite (define em vite.config.js):
// versão (package.json) + commit curto do git. Link do repositório é fixo.
// Sem define (testes/ambiente), os valores caem em 'dev'.
export const buildInfo = {
  version: typeof __APP_VERSION__ !== 'undefined' ? __APP_VERSION__ : 'dev',
  commit: typeof __APP_COMMIT__ !== 'undefined' ? __APP_COMMIT__ : 'dev',
  github: 'https://github.com/lucasbaccan/arandu'
};
