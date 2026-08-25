import { writable } from 'svelte/store';
import { api } from './api.js';
import { setUnauthorizedHandler } from './sessionExpired.js';
import { navigate } from './router.js';
import { showToast } from './toastStore.js';

export const user = writable(null);
export const authConfig = writable({ minPasswordLength: 3 });
export const authReady = writable(false);

// Qualquer rota autenticada que responda 401 derruba a sessão local e manda pro
// login, em vez de deixar o erro cru na tela.
setUnauthorizedHandler(() => {
  let wasLogged = false;
  user.update((u) => {
    wasLogged = Boolean(u);
    return null;
  });
  if (!wasLogged) return;
  showToast('Sua sessão expirou. Entre de novo.', 'error');
  navigate('/entrar');
});

export async function initAuth() {
  try {
    const { user: me } = await api.eu();
    user.set(me);
  } catch {
    user.set(null);
  }
  try {
    authConfig.set(await api.configuracao());
  } catch {
    // mantém defaults
  }
  authReady.set(true);
}

export async function login(email, password) {
  const { user: u } = await api.entrar({ email, password });
  user.set(u);
  return u;
}

export async function register(name, email, password) {
  const { user: u } = await api.criarConta({ name, email, password });
  user.set(u);
  return u;
}

export async function logout() {
  try {
    await api.sair();
  } catch {
    // segue mesmo sem confirmação do servidor
  }
  user.set(null);
}
