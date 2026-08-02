import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import Login from './Login.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    login: vi.fn()
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

describe('Tela de Login', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('exibe erro de validação ao enviar vazio', async () => {
    render(Login);
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    expect(await screen.findByText('Informe seu e-mail.')).toBeInTheDocument();
  });

  it('faz login com sucesso e navega para o dashboard', async () => {
    api.login.mockResolvedValue({
      user: { id: '1', email: 'ana@x.com', name: 'Ana' }
    });
    render(Login);

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@x.com');
    await userEvent.type(screen.getByLabelText('Senha'), 'segredo');
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    await waitFor(() =>
      expect(api.login).toHaveBeenCalledWith({
        email: 'ana@x.com',
        password: 'segredo'
      })
    );
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/dashboard'));
  });

  it('exibe mensagem de erro do servidor', async () => {
    api.login.mockRejectedValue(new Error('E-mail ou senha inválidos.'));
    render(Login);

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@x.com');
    await userEvent.type(screen.getByLabelText('Senha'), 'errada');
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    expect(await screen.findByText('E-mail ou senha inválidos.')).toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();
  });
});
