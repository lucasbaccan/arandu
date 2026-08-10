import { writable } from 'svelte/store';

export const route = writable(currentPath());

function currentPath() {
  const path = window.location.pathname;
  return path.length > 1 && path.endsWith('/') ? path.slice(0, -1) : path;
}

export function navigate(to) {
  if (to === currentPath()) return;
  window.history.pushState({}, '', to);
  // route guarda só o pathname (igual ao popstate) — os regex de rota em
  // App.svelte não esperam query string. Quem precisa de query (ex.: ?pin=)
  // lê window.location.search direto no componente.
  route.set(currentPath());
}

window.addEventListener('popstate', () => route.set(currentPath()));
