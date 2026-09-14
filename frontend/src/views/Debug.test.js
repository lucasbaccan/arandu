import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/svelte';
import Debug from './Debug.svelte';

vi.mock('../lib/router.js', async () => {
  const { writable } = await import('svelte/store');
  return {
    route: writable('/debug'),
    navigate: vi.fn()
  };
});

vi.mock('../lib/authStore.js', async () => {
  const { writable } = await import('svelte/store');
  return {
    user: writable(null),
    logout: vi.fn()
  };
});

vi.mock('../lib/themeStore.js', async () => {
  const { writable } = await import('svelte/store');
  return {
    theme: writable('light'),
    toggleTheme: vi.fn()
  };
});

import { user } from '../lib/authStore.js';
import { navigate } from '../lib/router.js';

describe('Debug', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    user.set(null);
  });

  it('mostra o título e os atalhos internos', () => {
    render(Debug);
    expect(screen.getByRole('heading', { name: 'Debug — Arandu' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Atalhos internos' })).toBeInTheDocument();
  });

  it('tem link para o mapa do site e para o guia de componentes (/tela)', async () => {
    render(Debug);
    await fireEvent.click(screen.getByRole('link', { name: /Mapa do site/ }));
    expect(navigate).toHaveBeenCalledWith('/mapa-do-site');
    await fireEvent.click(screen.getByRole('link', { name: /Guia de componentes/ }));
    expect(navigate).toHaveBeenCalledWith('/tela');
  });

  it('tem link para a doc da API que abre em nova aba sem passar pelo router', async () => {
    render(Debug);
    const link = screen.getByRole('link', { name: /Doc da API/ });
    expect(link).toHaveAttribute('href', '/api/docs/');
    expect(link).toHaveAttribute('target', '_blank');
    await fireEvent.click(link);
    expect(navigate).not.toHaveBeenCalled();
  });

  it('mostra o estado da aplicação (rota, tema e usuário)', () => {
    render(Debug);
    expect(screen.getByText('Rota (store)')).toBeInTheDocument();
    expect(screen.getByText('/debug')).toBeInTheDocument();
    expect(screen.getByText('Tema')).toBeInTheDocument();
    expect(screen.getByText('light')).toBeInTheDocument();
    expect(screen.getByText('Usuário')).toBeInTheDocument();
    expect(screen.getByText('deslogado')).toBeInTheDocument();
  });

  it('mostra o usuário logado quando existe sessão', () => {
    user.set({ id: '1', name: 'Ana', email: 'ana@exemplo.com' });
    render(Debug);
    expect(screen.getByText('Ana <ana@exemplo.com>')).toBeInTheDocument();
  });

  it('navega para o mapa completo pelo rodapé', async () => {
    render(Debug);
    await fireEvent.click(screen.getByRole('link', { name: 'Ver mapa completo do site ›' }));
    expect(navigate).toHaveBeenCalledWith('/mapa-do-site');
  });
});
