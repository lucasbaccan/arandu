import { writable } from 'svelte/store';
import { api } from './api.js';

export const user = writable(null);
export const authConfig = writable({ minPasswordLength: 3 });
export const authReady = writable(false);

export async function initAuth() {
  try {
    const { user: me } = await api.me();
    user.set(me);
  } catch {
    user.set(null);
  }
  try {
    authConfig.set(await api.config());
  } catch {
    // mantém defaults
  }
  authReady.set(true);
}

export async function login(email, password) {
  const { user: u } = await api.login({ email, password });
  user.set(u);
  return u;
}

export async function register(name, email, password) {
  const { user: u } = await api.register({ name, email, password });
  user.set(u);
  return u;
}

export async function logout() {
  try {
    await api.logout();
  } catch {
    // segue mesmo sem confirmação do servidor
  }
  user.set(null);
}
