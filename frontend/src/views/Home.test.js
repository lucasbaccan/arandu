import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import Home from './Home.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    public: {
      events: {
        resolvePin: vi.fn()
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
    render(Home);
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar na conta' }));
    expect(navigate).toHaveBeenCalledWith('/login');
  });

  it('navega para o registro ao clicar em Criar conta', async () => {
    render(Home);
    await fireEvent.click(screen.getByRole('button', { name: 'Criar conta' }));
    expect(navigate).toHaveBeenCalledWith('/register');
  });

  it('desabilita o botão Entrar enquanto o código estiver vazio', async () => {
    render(Home);
    expect(screen.getByRole('button', { name: 'Entrar' })).toBeDisabled();
    expect(api.public.events.resolvePin).not.toHaveBeenCalled();
  });

  it('navega pra live com o código já preenchido quando o código existe', async () => {
    api.public.events.resolvePin.mockResolvedValue({ id: '42' });
    render(Home);

    await fireEvent.input(screen.getByPlaceholderText('DEV-TEAM'), { target: { value: 'dev-team' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    expect(api.public.events.resolvePin).toHaveBeenCalledWith('dev-team');
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/audience/42?pin=dev-team'));
  });

  it('mostra erro quando o código não existe', async () => {
    api.public.events.resolvePin.mockRejectedValue(new ApiErrorLike(404, 'Código não encontrado.'));
    render(Home);

    await fireEvent.input(screen.getByPlaceholderText('DEV-TEAM'), { target: { value: 'naoexiste' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    expect(await screen.findByText('Código não encontrado.')).toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();
  });
});
