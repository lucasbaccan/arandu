import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import Live from './Live.svelte';

vi.mock('../lib/api.js', () => ({
  api: {
    public: {
      events: {
        live: {
          join: vi.fn(),
          state: vi.fn(),
          streamUrl: vi.fn((id, token) => `/api/public/events/${id}/live/stream?token=${token}`)
        }
      }
    }
  }
}));

import { api } from '../lib/api.js';

const snapshot = {
  eventTitle: 'Evento Live',
  questions: [
    {
      id: 'q1',
      title: 'Qual sua linguagem favorita?',
      type: 'GROUP',
      options: [
        { id: 'o1', text: 'Go' },
        { id: 'o2', text: 'JS' }
      ]
    }
  ],
  currentQuestionId: 'q1',
  pending: [{ id: 'p1', email: 'ana@exemplo.com', photo: '' }],
  groups: [
    { label: 'Go', participants: [] },
    { label: 'JS', participants: [] }
  ]
};

class FakeEventSource {
  constructor(url) {
    this.url = url;
    this.onmessage = null;
    this.onerror = null;
    FakeEventSource.instances.push(this);
  }
  close() {}
}
FakeEventSource.instances = [];

function mount() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new Live({ target, props: { id: '42' } });
  return within(target);
}

describe('Tela pública da apresentação (Live)', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    vi.clearAllMocks();
    FakeEventSource.instances = [];
    global.EventSource = FakeEventSource;
    sessionStorage.clear();
  });

  it('pede PIN e e-mail antes de mostrar qualquer coisa da apresentação', () => {
    const view = mount();

    expect(view.getByLabelText('PIN do evento')).toBeInTheDocument();
    expect(view.getByLabelText('E-mail')).toBeInTheDocument();
    expect(api.public.events.live.state).not.toHaveBeenCalled();
  });

  it('valida PIN vazio e e-mail inválido antes de enviar', async () => {
    const view = mount();

    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));
    expect(await view.findByText('Informe o PIN do evento.')).toBeInTheDocument();
    expect(api.public.events.live.join).not.toHaveBeenCalled();

    await fireEvent.input(view.getByLabelText('PIN do evento'), { target: { value: 'DEV-TEAM' } });
    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'nao-e-email' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));
    expect(await view.findByText('Informe um e-mail válido.')).toBeInTheDocument();
    expect(api.public.events.live.join).not.toHaveBeenCalled();
  });

  it('entra com sucesso e mostra a pergunta atual sem nenhum controle de admin', async () => {
    api.public.events.live.join.mockResolvedValue({ token: 'tok-123', role: 'participante' });
    api.public.events.live.state.mockResolvedValue(snapshot);
    const view = mount();

    await fireEvent.input(view.getByLabelText('PIN do evento'), { target: { value: 'dev-team' } });
    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(view.getByText('Evento Live')).toBeInTheDocument();
    expect(view.getByLabelText('ana@exemplo.com')).toBeInTheDocument();
    expect(view.getByLabelText('ana@exemplo.com')).toBeDisabled();

    expect(api.public.events.live.join).toHaveBeenCalledWith('42', {
      pinCode: 'dev-team',
      email: 'ana@exemplo.com'
    });
    expect(view.queryByRole('button', { name: 'Revelar todos' })).not.toBeInTheDocument();
    expect(view.queryByRole('button', { name: 'Sair do preview' })).not.toBeInTheDocument();
    expect(FakeEventSource.instances).toHaveLength(1);
    expect(FakeEventSource.instances[0].url).toContain('token=tok-123');
  });

  it('atualiza a tela sozinha quando chega uma mensagem do SSE', async () => {
    api.public.events.live.join.mockResolvedValue({ token: 'tok-123', role: 'participante' });
    api.public.events.live.state.mockResolvedValue(snapshot);
    const view = mount();

    await fireEvent.input(view.getByLabelText('PIN do evento'), { target: { value: 'dev-team' } });
    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));
    await view.findByText('Qual sua linguagem favorita?');

    const updated = {
      ...snapshot,
      pending: [],
      groups: [
        { label: 'Go', participants: [{ id: 'p1', email: 'ana@exemplo.com', photo: '' }] },
        { label: 'JS', participants: [] }
      ]
    };
    FakeEventSource.instances[0].onmessage({ data: JSON.stringify(updated) });

    await waitFor(() => {
      const goZone = view.getByText('Go').closest('.zone');
      expect(within(goZone).getByTitle('ana@exemplo.com')).toBeInTheDocument();
    });
  });

  it('mostra erro quando o PIN está errado, sem sair da tela de entrada', async () => {
    api.public.events.live.join.mockRejectedValue(new Error('PIN inválido.'));
    const view = mount();

    await fireEvent.input(view.getByLabelText('PIN do evento'), { target: { value: 'errado' } });
    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));

    expect(await view.findByText('PIN inválido.')).toBeInTheDocument();
    expect(view.getByLabelText('PIN do evento')).toBeInTheDocument();
  });

  it('restaura a sessão do sessionStorage sem pedir PIN de novo', async () => {
    sessionStorage.setItem('live-viewer-42', JSON.stringify({ token: 'tok-saved' }));
    api.public.events.live.state.mockResolvedValue(snapshot);

    const view = mount();

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(api.public.events.live.join).not.toHaveBeenCalled();
    expect(api.public.events.live.state).toHaveBeenCalledWith('42', 'tok-saved');
  });
});
