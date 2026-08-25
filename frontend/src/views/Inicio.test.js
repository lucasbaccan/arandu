import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import Inicio from './Inicio.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    publico: {
      eventos: {
        resolverPin: vi.fn()
      }
    }
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

class ApiErrorLike extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

describe('Tela inicial', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('navega para o login ao clicar em Entrar na conta', async () => {
    render(Inicio);
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar na conta' }));
    expect(navigate).toHaveBeenCalledWith('/entrar');
  });

  it('navega para o registro ao clicar em Criar conta', async () => {
    render(Inicio);
    await fireEvent.click(screen.getByRole('button', { name: 'Criar conta' }));
    expect(navigate).toHaveBeenCalledWith('/criar-conta');
  });

  it('desabilita o botão Entrar enquanto o código estiver vazio', async () => {
    render(Inicio);
    expect(screen.getByRole('button', { name: 'Entrar' })).toBeDisabled();
    expect(api.publico.eventos.resolverPin).not.toHaveBeenCalled();
  });

  it('navega pra live com o código já preenchido quando o código existe', async () => {
    api.publico.eventos.resolverPin.mockResolvedValue({ id: '42' });
    render(Inicio);

    await fireEvent.input(screen.getByPlaceholderText('DEV-TEAM'), { target: { value: 'dev-team' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    expect(api.publico.eventos.resolverPin).toHaveBeenCalledWith('dev-team');
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/plateia/42?pin=dev-team'));
  });

  it('mostra erro quando o código não existe', async () => {
    api.publico.eventos.resolverPin.mockRejectedValue(new ApiErrorLike(404, 'Código não encontrado.'));
    render(Inicio);

    await fireEvent.input(screen.getByPlaceholderText('DEV-TEAM'), { target: { value: 'naoexiste' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    expect(await screen.findByText('Código não encontrado.')).toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();
  });
});
