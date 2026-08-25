import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import Entrar from './Entrar.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    entrar: vi.fn()
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

describe('Tela de Login', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('exibe erro de validação ao enviar vazio', async () => {
    render(Entrar);
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    expect(await screen.findByText('Informe seu e-mail.')).toBeInTheDocument();
  });

  it('faz login com sucesso e navega para o painel', async () => {
    api.entrar.mockResolvedValue({
      user: { id: '1', email: 'ana@x.com', name: 'Ana' }
    });
    render(Entrar);

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@x.com');
    await userEvent.type(screen.getByLabelText('Senha'), 'segredo');
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    await waitFor(() =>
      expect(api.entrar).toHaveBeenCalledWith({
        email: 'ana@x.com',
        password: 'segredo'
      })
    );
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/painel'));
  });

  it('exibe mensagem de erro do servidor', async () => {
    api.entrar.mockRejectedValue(new Error('E-mail ou senha inválidos.'));
    render(Entrar);

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@x.com');
    await userEvent.type(screen.getByLabelText('Senha'), 'errada');
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    expect(await screen.findByText('E-mail ou senha inválidos.')).toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();
  });
});
