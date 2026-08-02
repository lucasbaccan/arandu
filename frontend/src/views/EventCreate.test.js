import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import EventCreate from './EventCreate.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    events: { create: vi.fn() }
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

describe('Novo evento', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('exige título', async () => {
    render(EventCreate);
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    expect(await screen.findByText('Informe o título do evento.')).toBeInTheDocument();
    expect(api.events.create).not.toHaveBeenCalled();
  });

  it('cria com PIN automático quando não personalizado', async () => {
    api.events.create.mockResolvedValue({ event: {} });
    render(EventCreate);

    await userEvent.type(screen.getByLabelText('Título'), 'Conecta DevOps');
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    await waitFor(() =>
      expect(api.events.create).toHaveBeenCalledWith({
        title: 'Conecta DevOps',
        pinCode: ''
      })
    );
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/dashboard'));
  });

  it('envia PIN personalizado quando marcado', async () => {
    api.events.create.mockResolvedValue({ event: {} });
    render(EventCreate);

    await userEvent.type(screen.getByLabelText('Título'), 'Retro');
    await fireEvent.click(screen.getByLabelText('Definir PIN personalizado'));
    await userEvent.type(await screen.findByLabelText('PIN'), 'meu-pin_1');
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    await waitFor(() =>
      expect(api.events.create).toHaveBeenCalledWith({
        title: 'Retro',
        pinCode: 'meu-pin_1'
      })
    );
  });

  it('rejeita PIN personalizado inválido', async () => {
    render(EventCreate);

    await userEvent.type(screen.getByLabelText('Título'), 'Retro');
    await fireEvent.click(screen.getByLabelText('Definir PIN personalizado'));
    await userEvent.type(await screen.findByLabelText('PIN'), 'espaço inválido');
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    expect(
      await screen.findByText(/O PIN deve ter 1 a 25 caracteres/)
    ).toBeInTheDocument();
    expect(api.events.create).not.toHaveBeenCalled();
  });

  it('mostra erro do servidor', async () => {
    api.events.create.mockRejectedValue(new Error('Este PIN já está em uso. Escolha outro.'));
    render(EventCreate);

    await userEvent.type(screen.getByLabelText('Título'), 'Retro');
    await fireEvent.click(screen.getByRole('button', { name: 'Criar evento' }));

    expect(
      await screen.findByText('Este PIN já está em uso. Escolha outro.')
    ).toBeInTheDocument();
  });
});
