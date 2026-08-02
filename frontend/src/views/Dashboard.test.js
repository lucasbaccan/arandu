import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import Dashboard from './Dashboard.svelte';
import { user } from '../lib/authStore.js';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    events: { list: vi.fn(), create: vi.fn() },
    logout: vi.fn()
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

function mountDashboard() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new Dashboard({ target });
  return within(target);
}

describe('Dashboard (meus eventos)', () => {
  beforeEach(() => {
    user.set(null);
    document.body.innerHTML = '';
    vi.clearAllMocks();
  });

  it('mostra o nome do usuário logado', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.events.list.mockResolvedValue({ events: [] });
    const view = mountDashboard();

    expect(await view.findByText('Olá, Ana')).toBeInTheDocument();
  });

  it('mostra estado vazio com atalho para criar evento', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.events.list.mockResolvedValue({ events: [] });
    const view = mountDashboard();

    expect(await view.findByText('Nenhum evento ainda')).toBeInTheDocument();
    await fireEvent.click(view.getByRole('button', { name: 'Criar evento' }));
    expect(navigate).toHaveBeenCalledWith('/events/new');
  });

  it('lista os eventos com PIN e status', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.events.list.mockResolvedValue({
      events: [
        {
          id: '1',
          title: 'Conecta DevOps',
          pinCode: '123456',
          status: 'PREPARATION',
          createdAt: '2026-08-02T00:00:00Z'
        },
        {
          id: '2',
          title: 'Retro',
          pinCode: '654321',
          status: 'PRESENTING',
          createdAt: '2026-08-01T00:00:00Z'
        }
      ]
    });
    const view = mountDashboard();

    expect(await view.findByText('Conecta DevOps')).toBeInTheDocument();
    expect(view.getByText('Retro')).toBeInTheDocument();
    expect(view.getByText('123456')).toBeInTheDocument();
    expect(view.getByText('Em preparação')).toBeInTheDocument();
    expect(view.getByText('Ao vivo')).toBeInTheDocument();
  });

  it('mostra erro da API', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.events.list.mockRejectedValue(new Error('Erro inesperado. Tente novamente.'));
    const view = mountDashboard();

    expect(
      await view.findByText('Erro inesperado. Tente novamente.')
    ).toBeInTheDocument();
  });

  it('sai da conta e volta para a home', async () => {
    user.set({ id: '1', name: 'Ana' });
    api.events.list.mockResolvedValue({ events: [] });
    const view = mountDashboard();
    await view.findByText('Olá, Ana');

    await fireEvent.click(view.getByRole('button', { name: 'Sair' }));

    await waitFor(() => expect(api.logout).toHaveBeenCalled());
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/'));
  });
});
