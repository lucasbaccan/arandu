import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import Administracao from './Administracao.svelte';
import { user } from '../lib/authStore.js';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    admin: {
      usuarios: {
        listar: vi.fn(),
        excluir: vi.fn(),
        gerarLinkRedefinicao: vi.fn()
      },
      eventos: { listar: vi.fn() },
      configuracoes: { buscar: vi.fn(), atualizar: vi.fn() }
    },
    sair: vi.fn()
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

function mountAdmin() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new Administracao({ target });
  return within(target);
}

describe('Administração (super admin)', () => {
  beforeEach(() => {
    user.set({ id: '1', name: 'Admin', role: 'super_admin' });
    document.body.innerHTML = '';
    vi.clearAllMocks();
    api.admin.eventos.listar.mockResolvedValue({ events: [] });
    api.admin.configuracoes.buscar.mockResolvedValue({ registrationEnabled: true });
  });

  it('lista usuários com papel e contagem de eventos', async () => {
    api.admin.usuarios.listar.mockResolvedValue({
      users: [
        { id: '1', name: 'Admin', email: 'admin@exemplo.com', role: 'super_admin', eventCount: 0, createdAt: '2026-08-01T00:00:00Z' },
        { id: '2', name: 'Bia', email: 'bia@exemplo.com', role: 'organizer', eventCount: 3, createdAt: '2026-08-02T00:00:00Z' }
      ]
    });
    const view = mountAdmin();

    expect(await view.findByText('Bia')).toBeInTheDocument();
    expect(view.getByText('Super admin')).toBeInTheDocument();
    expect(view.getByText('Organizador')).toBeInTheDocument();
    expect(view.getByText('3')).toBeInTheDocument();
  });

  it('não mostra botão de excluir para a própria conta', async () => {
    api.admin.usuarios.listar.mockResolvedValue({
      users: [
        { id: '1', name: 'Admin', email: 'admin@exemplo.com', role: 'super_admin', eventCount: 0, createdAt: '2026-08-01T00:00:00Z' },
        { id: '2', name: 'Bia', email: 'bia@exemplo.com', role: 'organizer', eventCount: 0, createdAt: '2026-08-02T00:00:00Z' }
      ]
    });
    const view = mountAdmin();
    await view.findByText('Bia');

    expect(view.getAllByRole('button', { name: 'Excluir' })).toHaveLength(1);
  });

  it('gera link de redefinição de senha e mostra a URL', async () => {
    api.admin.usuarios.listar.mockResolvedValue({
      users: [{ id: '2', name: 'Bia', email: 'bia@exemplo.com', role: 'organizer', eventCount: 0, createdAt: '2026-08-02T00:00:00Z' }]
    });
    api.admin.usuarios.gerarLinkRedefinicao.mockResolvedValue({
      token: 'abc123',
      expiresAt: '2026-08-02T01:00:00Z'
    });
    const view = mountAdmin();
    await view.findByText('Bia');

    await fireEvent.click(view.getByRole('button', { name: 'Gerar link de redefinição' }));

    expect(await view.findByText(/token=abc123/)).toBeInTheDocument();
  });

  it('exclui usuário após confirmação', async () => {
    window.confirm = vi.fn(() => true);
    api.admin.usuarios.listar.mockResolvedValue({
      users: [{ id: '2', name: 'Bia', email: 'bia@exemplo.com', role: 'organizer', eventCount: 0, createdAt: '2026-08-02T00:00:00Z' }]
    });
    api.admin.usuarios.excluir.mockResolvedValue({ ok: true });
    const view = mountAdmin();
    await view.findByText('Bia');

    await fireEvent.click(view.getByRole('button', { name: 'Excluir' }));

    await waitFor(() => expect(api.admin.usuarios.excluir).toHaveBeenCalledWith('2'));
    await waitFor(() => expect(view.queryByText('Bia')).not.toBeInTheDocument());
  });

  it('não exclui usuário se a confirmação for cancelada', async () => {
    window.confirm = vi.fn(() => false);
    api.admin.usuarios.listar.mockResolvedValue({
      users: [{ id: '2', name: 'Bia', email: 'bia@exemplo.com', role: 'organizer', eventCount: 0, createdAt: '2026-08-02T00:00:00Z' }]
    });
    const view = mountAdmin();
    await view.findByText('Bia');

    await fireEvent.click(view.getByRole('button', { name: 'Excluir' }));

    expect(api.admin.usuarios.excluir).not.toHaveBeenCalled();
  });

  it('lista eventos de todos os organizadores na aba Eventos', async () => {
    api.admin.usuarios.listar.mockResolvedValue({ users: [] });
    api.admin.eventos.listar.mockResolvedValue({
      events: [
        {
          id: '10',
          title: 'Evento da Bia',
          pinCode: '123456',
          status: 'PREPARATION',
          createdAt: '2026-08-02T00:00:00Z',
          ownerName: 'Bia',
          ownerEmail: 'bia@exemplo.com'
        }
      ]
    });
    const view = mountAdmin();

    await fireEvent.click(view.getByRole('tab', { name: 'Eventos' }));

    expect(await view.findByText('Evento da Bia')).toBeInTheDocument();
    expect(view.getByText('bia@exemplo.com')).toBeInTheDocument();

    await fireEvent.click(view.getByText('Evento da Bia'));
    expect(navigate).toHaveBeenCalledWith('/evento/10');
  });

  it('alterna se novos cadastros são permitidos na aba Configurações', async () => {
    api.admin.usuarios.listar.mockResolvedValue({ users: [] });
    api.admin.configuracoes.atualizar.mockResolvedValue({ registrationEnabled: false });
    const view = mountAdmin();

    await fireEvent.click(view.getByRole('tab', { name: 'Configurações' }));
    const toggle = await view.findByRole('switch');
    expect(toggle).toHaveAttribute('aria-checked', 'true');

    await fireEvent.click(toggle);

    await waitFor(() =>
      expect(api.admin.configuracoes.atualizar).toHaveBeenCalledWith({ registrationEnabled: false })
    );
  });
});
