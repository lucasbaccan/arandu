import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import userEvent from '@testing-library/user-event';
import { get } from 'svelte/store';
import EventEdit from './EventEdit.svelte';
import { toast } from '../lib/toastStore.js';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    events: { get: vi.fn(), update: vi.fn() }
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

const event = {
  id: '42',
  title: 'Conecta DevOps',
  pinCode: '123456',
  status: 'PREPARATION',
  configShowRanking: true,
  createdAt: '2026-08-02T00:00:00Z'
};

function mount() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new EventEdit({ target, props: { id: '42' } });
  return within(target);
}

describe('Editar evento', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    toast.set(null);
    vi.clearAllMocks();
  });

  it('carrega o evento e preenche o formulário', async () => {
    api.events.get.mockResolvedValue({ event });
    const view = mount();

    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    expect(view.getByText('123456')).toBeInTheDocument();
    expect(view.getByText('Em preparação')).toBeInTheDocument();
    expect(view.getByLabelText('Exibir ranking de pontos').checked).toBe(true);
  });

  it('exige título ao salvar', async () => {
    api.events.get.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.input(view.getByLabelText('Título'), { target: { value: '' } });
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    expect(await view.findByText('Informe o título do evento.')).toBeInTheDocument();
    expect(api.events.update).not.toHaveBeenCalled();
  });

  it('salva mantendo o PIN quando a troca não é marcada', async () => {
    api.events.get.mockResolvedValue({ event });
    api.events.update.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.input(view.getByLabelText('Título'), { target: { value: 'Título novo' } });
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    await waitFor(() =>
      expect(api.events.update).toHaveBeenCalledWith('42', {
        title: 'Título novo',
        pinCode: '',
        configShowRanking: true
      })
    );
    await waitFor(() => expect(get(toast)?.message).toBe('Alterações salvas!'));
  });

  it('salva com novo PIN quando marcado', async () => {
    api.events.get.mockResolvedValue({ event });
    api.events.update.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByLabelText('Definir novo PIN'));
    const pinInput = await view.findByLabelText('Novo PIN');
    await userEvent.clear(pinInput);
    await userEvent.type(pinInput, 'dev-team');
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    await waitFor(() =>
      expect(api.events.update).toHaveBeenCalledWith('42', {
        title: 'Conecta DevOps',
        pinCode: 'dev-team',
        configShowRanking: true
      })
    );
    await waitFor(() =>
      expect(view.queryByLabelText('Novo PIN')).not.toBeInTheDocument()
    );
    expect(view.getByText('dev-team')).toBeInTheDocument();
  });

  it('rejeita novo PIN inválido', async () => {
    api.events.get.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByLabelText('Definir novo PIN'));
    await userEvent.type(await view.findByLabelText('Novo PIN'), 'com espaço');
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    expect(await view.findByText(/O PIN deve ter 1 a 25 caracteres/)).toBeInTheDocument();
    expect(api.events.update).not.toHaveBeenCalled();
  });

  it('mostra erro do servidor', async () => {
    api.events.get.mockResolvedValue({ event });
    api.events.update.mockRejectedValue(new Error('Este PIN já está em uso. Escolha outro.'));
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    expect(
      await view.findByText('Este PIN já está em uso. Escolha outro.')
    ).toBeInTheDocument();
  });

  it('mostra tela de não encontrado quando o evento não existe', async () => {
    const err = new Error('Evento não encontrado.');
    err.status = 404;
    api.events.get.mockRejectedValue(err);
    const view = mount();

    expect(await view.findByText('Evento não encontrado')).toBeInTheDocument();
  });
});
