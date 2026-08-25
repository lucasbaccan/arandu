import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import CriarConta from './CriarConta.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    register: vi.fn()
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

async function fillAndSubmit({ name, email, password, confirm }) {
  if (name) await userEvent.type(screen.getByLabelText('Nome completo'), name);
  if (email) await userEvent.type(screen.getByLabelText('E-mail'), email);
  if (password) await userEvent.type(screen.getByLabelText('Senha'), password);
  if (confirm !== undefined) {
    await userEvent.type(screen.getByLabelText('Confirmar senha'), confirm);
  }
  await fireEvent.click(screen.getByRole('button', { name: 'Criar conta' }));
}

describe('Tela de Criar conta', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('exibe erro de validação ao enviar vazio', async () => {
    render(CriarConta);
    await fireEvent.click(screen.getByRole('button', { name: 'Criar conta' }));
    expect(await screen.findByText('Informe seu nome.')).toBeInTheDocument();
  });

  it('rejeita senha abaixo do mínimo configurado', async () => {
    render(CriarConta);
    await fillAndSubmit({
      name: 'Ana',
      email: 'ana@x.com',
      password: 'ab',
      confirm: 'ab'
    });
    expect(
      await screen.findByText('A senha deve ter pelo menos 3 caracteres.')
    ).toBeInTheDocument();
    expect(api.register).not.toHaveBeenCalled();
  });

  it('rejeita senhas diferentes na confirmação', async () => {
    render(CriarConta);
    await fillAndSubmit({
      name: 'Ana',
      email: 'ana@x.com',
      password: 'segredo',
      confirm: 'segredo-diferente'
    });
    expect(await screen.findByText('As senhas não conferem.')).toBeInTheDocument();
  });

  it('cria conta com sucesso e navega para o painel', async () => {
    api.register.mockResolvedValue({
      user: { id: '1', email: 'ana@x.com', name: 'Ana' }
    });
    render(CriarConta);

    await fillAndSubmit({
      name: 'Ana',
      email: 'ana@x.com',
      password: 'segredo',
      confirm: 'segredo'
    });

    await waitFor(() =>
      expect(api.register).toHaveBeenCalledWith({
        name: 'Ana',
        email: 'ana@x.com',
        password: 'segredo'
      })
    );
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/painel'));
  });

  it('exibe mensagem de erro do servidor', async () => {
    api.register.mockRejectedValue(new Error('Este e-mail já está cadastrado.'));
    render(CriarConta);

    await fillAndSubmit({
      name: 'Ana',
      email: 'ana@x.com',
      password: 'segredo',
      confirm: 'segredo'
    });

    expect(
      await screen.findByText('Este e-mail já está cadastrado.')
    ).toBeInTheDocument();
  });
});
