import { writable } from 'svelte/store';

const STORAGE_KEY = 'arandu-theme';

// O toggle de tema é um elemento fixo dos dois shells (canto da tela pública,
// topbar do organizador), então a preferência precisa sobreviver à navegação e
// ao F5. É a única coisa que o app guarda no navegador — sessão de plateia
// continua vivendo só na URL.
function readStored() {
  try {
    const saved = window.localStorage.getItem(STORAGE_KEY);
    if (saved === 'light' || saved === 'dark') return saved;
  } catch {
    // localStorage indisponível (modo privado, iframe restrito): segue no padrão
  }
  return null;
}

function systemPrefersDark() {
  return Boolean(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches);
}

const initial = readStored() || (systemPrefersDark() ? 'dark' : 'light');

export const theme = writable(initial);

export function applyTheme(value) {
  document.documentElement.dataset.theme = value;
}

export function toggleTheme() {
  theme.update((current) => {
    const next = current === 'dark' ? 'light' : 'dark';
    try {
      window.localStorage.setItem(STORAGE_KEY, next);
    } catch {
      // sem persistência: o tema ainda vale para esta sessão
    }
    return next;
  });
}

theme.subscribe((value) => {
  if (typeof document !== 'undefined') applyTheme(value);
});
