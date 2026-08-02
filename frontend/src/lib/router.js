import { writable } from 'svelte/store';

export const route = writable(currentPath());

function currentPath() {
  const path = window.location.pathname;
  return path.length > 1 && path.endsWith('/') ? path.slice(0, -1) : path;
}

export function navigate(to) {
  if (to === currentPath()) return;
  window.history.pushState({}, '', to);
  route.set(to);
}

window.addEventListener('popstate', () => route.set(currentPath()));
