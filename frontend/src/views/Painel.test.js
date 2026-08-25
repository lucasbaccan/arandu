import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import Painel from './Painel.svelte';
import { user } from '../lib/authStore.js';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    eventos: { listar: vi.fn(), criar: vi.fn() },
    sair: vi.fn()
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

function mountDashboard() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new Painel({ target });
  return within(target);
}

describe('Dashboard (meus eventos)', () => {
  beforeEach(() => {
    user.set(null);
    document.body.innerHTML = '';
    vi.clearAllMocks();
  });

  it('mostra o avatar com a inicial do usuário logado', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.eventos.listar.mockResolvedValue({ events: [] });
    const view = mountDashboard();

    const avatar = await view.findByTitle('Ana');
    expect(avatar).toHaveTextContent('A');
  });

  it('mostra estado vazio com atalho para criar evento', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.eventos.listar.mockResolvedValue({ events: [] });
    const view = mountDashboard();

    expect(await view.findByText('Nenhum evento ainda')).toBeInTheDocument();
    await fireEvent.click(view.getByRole('button', { name: 'Criar evento' }));
    expect(navigate).toHaveBeenCalledWith('/evento/novo');
  });

  it('lista os eventos com PIN, status e data em formato brasileiro', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.eventos.listar.mockResolvedValue({
      events: [
        {
          id: '1',
          title: 'Conecta DevOps',
          pinCode: '123456',
          status: 'PREPARATION',
          createdAt: '2026-08-02T12:00:00Z',
          questionCount: 3,
          participantCount: 0
        },
        {
          id: '2',
          title: 'Retro',
          pinCode: '654321',
          status: 'PRESENTING',
          createdAt: '2026-08-01T00:00:00Z',
          questionCount: 2,
          participantCount: 5
        }
      ]
    });
    const view = mountDashboard();

    expect(await view.findByText('Conecta DevOps')).toBeInTheDocument();
    expect(view.getByText('Retro')).toBeInTheDocument();
    expect(view.getByText('123456')).toBeInTheDocument();
    expect(view.getByText('Em preparação', { selector: '.chip' })).toBeInTheDocument();
    // "Ao vivo" também é um filtro no topo — aqui interessa o chip da linha.
    expect(view.getByText('Ao vivo', { selector: '.chip' })).toBeInTheDocument();
    expect(view.getByText('02/08/2026')).toBeInTheDocument();
  });

  it('mostra erro da API', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.eventos.listar.mockRejectedValue(new Error('Erro inesperado. Tente novamente.'));
    const view = mountDashboard();

    expect(
      await view.findByText('Erro inesperado. Tente novamente.')
    ).toBeInTheDocument();
  });

  it('abre o evento ao clicar na linha', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.eventos.listar.mockResolvedValue({
      events: [
        {
          id: '1',
          title: 'Conecta DevOps',
          pinCode: '123456',
          status: 'PREPARATION',
          createdAt: '2026-08-02T00:00:00Z',
          questionCount: 1,
          participantCount: 0
        }
      ]
    });
    const view = mountDashboard();
    await view.findByText('Conecta DevOps');

    await fireEvent.click(view.getByText('Conecta DevOps'));

    expect(navigate).toHaveBeenCalledWith('/evento/1');
  });

  it('filtra eventos pela busca de título ou PIN', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.eventos.listar.mockResolvedValue({
      events: [
        {
          id: '1',
          title: 'Conecta DevOps',
          pinCode: 'DEV123',
          status: 'PREPARATION',
          createdAt: '2026-08-02T00:00:00Z',
          questionCount: 1,
          participantCount: 0
        },
        {
          id: '2',
          title: 'Retro',
          pinCode: 'RETRO1',
          status: 'PRESENTING',
          createdAt: '2026-08-01T00:00:00Z',
          questionCount: 2,
          participantCount: 5
        }
      ]
    });
    const view = mountDashboard();
    await view.findByText('Conecta DevOps');

    await fireEvent.input(view.getByPlaceholderText('Buscar por título ou PIN'), {
      target: { value: 'retro' }
    });

    expect(view.queryByText('Conecta DevOps')).not.toBeInTheDocument();
    expect(view.getByText('Retro')).toBeInTheDocument();
  });

  it('sai da conta e volta para a home', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.eventos.listar.mockResolvedValue({ events: [] });
    const view = mountDashboard();
    await view.findByTitle('Ana');

    await fireEvent.click(view.getByTitle('Ana'));
    await fireEvent.click(view.getByRole('menuitem', { name: 'Sair' }));

    await waitFor(() => expect(api.sair).toHaveBeenCalled());
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/'));
  });
});
