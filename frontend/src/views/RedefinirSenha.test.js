import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import RedefinirSenha from './RedefinirSenha.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', async () => {
  const actual = await vi.importActual('../lib/api.js');
  return {
    ...actual,
    api: {
      publico: {
        redefinirSenha: {
          validarToken: vi.fn(),
          redefinir: vi.fn()
        }
      }
    }
  };
});

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

function setToken(token) {
  window.history.pushState({}, '', token ? `/redefinir-senha?token=${token}` : '/redefinir-senha');
}

describe('Redefinir senha (link do super admin)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('sem token na URL, mostra link inválido', async () => {
    setToken('');
    render(RedefinirSenha);

    expect(await screen.findByText('Link inválido')).toBeInTheDocument();
    expect(api.publico.redefinirSenha.validarToken).not.toHaveBeenCalled();
  });

  it('token expirado/inexistente mostra link inválido', async () => {
    setToken('token-ruim');
    api.publico.redefinirSenha.validarToken.mockRejectedValue(new Error('Link inválido ou expirado.'));
    render(RedefinirSenha);

    expect(await screen.findByText('Link inválido')).toBeInTheDocument();
  });

  it('token válido permite definir nova senha', async () => {
    setToken('token-bom');
    api.publico.redefinirSenha.validarToken.mockResolvedValue({ valid: true });
    api.publico.redefinirSenha.redefinir.mockResolvedValue({ ok: true });
    render(RedefinirSenha);

    await screen.findByText('Definir nova senha');

    await userEvent.type(screen.getByLabelText('Nova senha'), 'senha-nova-123');
    await userEvent.type(screen.getByLabelText('Confirmar nova senha'), 'senha-nova-123');
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar nova senha' }));

    await waitFor(() =>
      expect(api.publico.redefinirSenha.redefinir).toHaveBeenCalledWith('token-bom', 'senha-nova-123')
    );
    expect(await screen.findByText('Senha redefinida')).toBeInTheDocument();
  });

  it('confirmação divergente mostra erro sem chamar a API', async () => {
    setToken('token-bom');
    api.publico.redefinirSenha.validarToken.mockResolvedValue({ valid: true });
    render(RedefinirSenha);
    await screen.findByText('Definir nova senha');

    await userEvent.type(screen.getByLabelText('Nova senha'), 'senha-nova-123');
    await userEvent.type(screen.getByLabelText('Confirmar nova senha'), 'outra-coisa');
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar nova senha' }));

    expect(await screen.findByText('A confirmação não confere com a nova senha.')).toBeInTheDocument();
    expect(api.publico.redefinirSenha.redefinir).not.toHaveBeenCalled();
  });

  it('botão de voltar ao login navega para /entrar quando o link é inválido', async () => {
    setToken('');
    render(RedefinirSenha);
    await screen.findByText('Link inválido');

    await fireEvent.click(screen.getByRole('button', { name: 'Voltar ao login' }));
    expect(navigate).toHaveBeenCalledWith('/entrar');
  });
});
