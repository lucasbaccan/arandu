import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { user, authConfig, authReady, initAuth, login, register, logout } from './authStore.js';
import { api } from './api.js';

vi.mock('./api.js', () => ({
  api: {
    me: vi.fn(),
    config: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn()
  }
}));

describe('authStore', () => {
  beforeEach(() => {
    user.set(null);
    authConfig.set({ minPasswordLength: 3 });
    authReady.set(false);
    vi.clearAllMocks();
  });

  it('initAuth carrega usuário logado e config', async () => {
    api.me.mockResolvedValue({ user: { id: '1', name: 'Ana' } });
    api.config.mockResolvedValue({ minPasswordLength: 8 });

    await initAuth();

    expect(get(user)).toEqual({ id: '1', name: 'Ana' });
    expect(get(authConfig)).toEqual({ minPasswordLength: 8 });
    expect(get(authReady)).toBe(true);
  });

  it('initAuth extrai o user do wrapper (regressão: F5 não pode mostrar undefined)', async () => {
    api.me.mockResolvedValue({ user: { id: '2', name: 'Bia' } });
    api.config.mockResolvedValue({ minPasswordLength: 3 });

    await initAuth();

    expect(get(user)?.name).toBe('Bia');
  });

  it('initAuth zera usuário quando sessão expirada', async () => {
    api.me.mockRejectedValue(new Error('401'));
    api.config.mockRejectedValue(new Error('offline'));

    await initAuth();

    expect(get(user)).toBeNull();
    expect(get(authConfig)).toEqual({ minPasswordLength: 3 });
    expect(get(authReady)).toBe(true);
  });

  it('login define o usuário no store', async () => {
    api.login.mockResolvedValue({ user: { id: '2', name: 'Bia' } });

    const u = await login('bia@x.com', '123');
    expect(u.name).toBe('Bia');
    expect(get(user)).toEqual({ id: '2', name: 'Bia' });
  });

  it('register define o usuário no store', async () => {
    api.register.mockResolvedValue({ user: { id: '3', name: 'Cadu' } });

    await register('Cadu', 'cadu@x.com', '123');
    expect(get(user).name).toBe('Cadu');
  });

  it('logout limpa o usuário mesmo com falha no servidor', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.logout.mockRejectedValue(new Error('offline'));

    await logout();

    expect(get(user)).toBeNull();
  });
});
