import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/svelte';
import MapaDoSite from './MapaDoSite.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

import { navigate } from '../lib/router.js';

describe('Mapa do site', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('mostra o título e os grupos de fluxo', () => {
    render(MapaDoSite);
    expect(screen.getByRole('heading', { name: 'Mapa do site — Arandu' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Públicas — entrada' })).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { name: 'Participantes — evento ao vivo' })
    ).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Organizador — autenticado' })).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { name: 'Internas — desenvolvimento' })
    ).toBeInTheDocument();
  });

  it('lista as rotas dinâmicas como padrão (não clicáveis)', () => {
    render(MapaDoSite);
    expect(screen.getByText('/responder/{id}')).toBeInTheDocument();
    expect(screen.getByText('/evento/{id}')).toBeInTheDocument();
    expect(screen.getByText('/palco/{id}')).toBeInTheDocument();
    expect(screen.getByText('/palco/{id}/apresentar')).toBeInTheDocument();
    expect(screen.getByText('/plateia/{id}')).toBeInTheDocument();
  });

  it('transforma as rotas estáticas em links', () => {
    render(MapaDoSite);
    for (const path of ['/', '/entrar', '/criar-conta', '/painel', '/evento/novo', '/como-funciona', '/privacidade', '/tela', '/mapa-do-site', '/debug']) {
      expect(screen.getByRole('link', { name: path })).toBeInTheDocument();
    }
  });

  it('navega ao clicar em uma rota estática', async () => {
    render(MapaDoSite);
    await fireEvent.click(screen.getByRole('link', { name: '/entrar' }));
    expect(navigate).toHaveBeenCalledWith('/entrar');
  });

  it('navega para /debug e /tela pelo rodapé', async () => {
    render(MapaDoSite);
    await fireEvent.click(screen.getByRole('link', { name: '‹ Debug' }));
    expect(navigate).toHaveBeenCalledWith('/debug');
    await fireEvent.click(screen.getByRole('link', { name: 'Guia de componentes (/tela)' }));
    expect(navigate).toHaveBeenCalledWith('/tela');
  });
});
