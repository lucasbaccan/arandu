import { writable } from 'svelte/store';

export const toast = writable(null);

let timer = null;

export function showToast(message, type = 'success', duration = 3000) {
  toast.set({ message, type, id: Date.now() });
  clearTimeout(timer);
  timer = setTimeout(() => toast.set(null), duration);
}
