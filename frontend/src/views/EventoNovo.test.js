import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { get } from 'svelte/store';
import EventoNovo from './EventoNovo.svelte';
import { toast } from '../lib/toastStore.js';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    eventos: { criar: vi.fn() }
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

describe('Novo evento', () => {
  beforeEach(() => {
    toast.set(null);
    vi.clearAllMocks();
  });

  it('exige título', async () => {
    render(EventoNovo);
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    expect(await screen.findByText('Informe o título do evento.')).toBeInTheDocument();
    expect(api.eventos.criar).not.toHaveBeenCalled();
  });

  it('cria com PIN automático quando não personalizado', async () => {
    api.eventos.criar.mockResolvedValue({ event: {} });
    render(EventoNovo);

    await userEvent.type(screen.getByLabelText('Título do evento'), 'Conecta DevOps');
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    await waitFor(() =>
      expect(api.eventos.criar).toHaveBeenCalledWith({
        title: 'Conecta DevOps',
        pinCode: ''
      })
    );
    await waitFor(() =>
      expect(get(toast)?.message).toBe('Evento criado com sucesso!')
    );
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/painel'));
  });

  it('envia PIN personalizado quando marcado', async () => {
    api.eventos.criar.mockResolvedValue({ event: {} });
    render(EventoNovo);

    await userEvent.type(screen.getByLabelText('Título do evento'), 'Retro');
    await fireEvent.click(screen.getByRole('button', { name: /Personalizado/ }));
    await userEvent.type(await screen.findByLabelText('PIN personalizado'), 'meu-pin_1');
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    await waitFor(() =>
      expect(api.eventos.criar).toHaveBeenCalledWith({
        title: 'Retro',
        pinCode: 'meu-pin_1'
      })
    );
  });

  it('rejeita PIN personalizado inválido', async () => {
    render(EventoNovo);

    await userEvent.type(screen.getByLabelText('Título do evento'), 'Retro');
    await fireEvent.click(screen.getByRole('button', { name: /Personalizado/ }));
    await userEvent.type(await screen.findByLabelText('PIN personalizado'), 'espaço inválido');
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    expect(
      await screen.findByText(/O PIN deve ter 1 a 25 caracteres/)
    ).toBeInTheDocument();
    expect(api.eventos.criar).not.toHaveBeenCalled();
  });

  it('mostra erro do servidor', async () => {
    api.eventos.criar.mockRejectedValue(new Error('Este PIN já está em uso. Escolha outro.'));
    render(EventoNovo);

    await userEvent.type(screen.getByLabelText('Título do evento'), 'Retro');
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    expect(
      await screen.findByText('Este PIN já está em uso. Escolha outro.')
    ).toBeInTheDocument();
  });
});
