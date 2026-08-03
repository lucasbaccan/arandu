import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import App from './App.svelte';

vi.mock('./lib/authStore.js', async () => {
  const { writable } = await import('svelte/store');
  return {
    user: writable(null),
    authReady: writable(false),
    authConfig: writable({ minPasswordLength: 3 }),
    initAuth: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn()
  };
});

vi.mock('./lib/router.js', async () => {
  const { writable } = await import('svelte/store');
  return {
    route: writable('/'),
    navigate: vi.fn()
  };
});

vi.mock('./lib/api.js', () => ({
  api: {
    events: {
      create: vi.fn(),
      list: vi.fn().mockResolvedValue({ events: [] }),
      get: vi.fn().mockResolvedValue({
        event: {
          id: '42',
          title: 'X',
          pinCode: '123456',
          status: 'PREPARATION',
          configShowRanking: false,
          createdAt: '2026-01-01T00:00:00Z'
        }
      }),
      update: vi.fn()
    },
    logout: vi.fn()
  }
}));

import { user, authReady } from './lib/authStore.js';
import { route, navigate } from './lib/router.js';

describe('App (guardas de rota)', () => {
  beforeEach(() => {
    user.set(null);
    authReady.set(false);
    route.set('/');
    vi.clearAllMocks();
  });

  it('não mostra a tela inicial enquanto a autenticação não termina (regressão: F5)', () => {
    route.set('/dashboard');
    render(App);

    expect(screen.queryByRole('button', { name: 'Entrar' })).not.toBeInTheDocument();
    expect(screen.getByText('Carregando…')).toBeInTheDocument();
  });

  it('redireciona para o dashboard quando já está logado', async () => {
    route.set('/');
    render(App);

    user.set({ id: '1', name: 'Ana' });
    authReady.set(true);

    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/dashboard'));
  });

  it('redireciona para a home quando não está logado', async () => {
    route.set('/dashboard');
    render(App);

    authReady.set(true);

    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/'));
  });

  it('mantém a tela de login quando não logado (regressão: botão Entrar)', async () => {
    route.set('/login');
    render(App);

    authReady.set(true);

    await new Promise((r) => setTimeout(r, 20));
    expect(navigate).not.toHaveBeenCalled();
  });

  it('redireciona para o dashboard ao abrir login/registro já logado', async () => {
    route.set('/login');
    render(App);

    user.set({ id: '1', name: 'Ana' });
    authReady.set(true);

    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/dashboard'));
  });

  it('redireciona para a home em /events/new sem sessão', async () => {
    route.set('/events/new');
    render(App);

    authReady.set(true);

    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/'));
  });

  it('redireciona para a home em /events/123 sem sessão', async () => {
    route.set('/events/123');
    render(App);

    authReady.set(true);

    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/'));
  });

  it('exibe o dashboard quando logado e autenticado', async () => {
    route.set('/dashboard');
    render(App);

    user.set({ id: '1', name: 'Ana' });
    authReady.set(true);

    expect(await screen.findByText('Olá, Ana')).toBeInTheDocument();
  });
});
