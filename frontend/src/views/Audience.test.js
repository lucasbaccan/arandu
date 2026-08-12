import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import Audience from './Audience.svelte';

vi.mock('../lib/api.js', () => ({
  api: {
    public: {
      events: {
        live: {
          join: vi.fn(),
          state: vi.fn(),
          streamUrl: vi.fn((id, token) => `/api/public/events/${id}/live/stream?token=${token}`),
          react: vi.fn().mockResolvedValue({ ok: true }),
          submitQuestion: vi.fn().mockResolvedValue({ ok: true })
        }
      }
    }
  }
}));

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

import { api } from '../lib/api.js';
import { navigate } from '../lib/router.js';

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
  ],
  blanked: false,
  message: '',
  interactionsEnabled: true
};

class FakeEventSource {
  constructor(url) {
    this.url = url;
    this.onmessage = null;
    this.onerror = null;
    this.listeners = {};
    FakeEventSource.instances.push(this);
  }
  addEventListener(name, cb) {
    this.listeners[name] = cb;
  }
  emit(name, detail) {
    this.listeners[name] && this.listeners[name](detail);
  }
  close() {}
}
FakeEventSource.instances = [];

function mount(search = '') {
  window.history.pushState({}, '', `/audience/42${search}`);
  const target = document.createElement('div');
  document.body.appendChild(target);
  new Audience({ target, props: { id: '42' } });
  return within(target);
}

class ApiErrorLike extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

describe('Tela pública da apresentação (Audience)', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    vi.clearAllMocks();
    FakeEventSource.instances = [];
    global.EventSource = FakeEventSource;
    sessionStorage.clear();
  });

  it('sem PIN na URL, redireciona pra Home sem chamar a API', () => {
    mount();

    expect(navigate).toHaveBeenCalledWith('/');
    expect(api.public.events.live.join).not.toHaveBeenCalled();
    expect(api.public.events.live.state).not.toHaveBeenCalled();
  });

  it('com PIN na URL, mostra direto a tela de identificação (e-mail ou convidado)', () => {
    const view = mount('?pin=dev-team');

    expect(navigate).not.toHaveBeenCalled();
    expect(view.getByLabelText('E-mail')).toBeInTheDocument();
    expect(view.getByRole('button', { name: 'Entrar como convidado' })).toBeInTheDocument();
    expect(view.queryByLabelText('Código do evento')).not.toBeInTheDocument();
  });

  it('valida e-mail inválido antes de enviar', async () => {
    const view = mount('?pin=dev-team');

    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'nao-e-email' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));
    expect(await view.findByText('Informe um e-mail válido.')).toBeInTheDocument();
    expect(api.public.events.live.join).not.toHaveBeenCalled();
  });

  it('entra com sucesso via e-mail e mostra a pergunta atual sem nenhum controle de admin', async () => {
    api.public.events.live.join.mockResolvedValue({ token: 'tok-123', role: 'participante' });
    api.public.events.live.state.mockResolvedValue(snapshot);
    const view = mount('?pin=dev-team');

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
    expect(sessionStorage.length).toBe(0);
  });

  it('entra como convidado sem informar e-mail', async () => {
    api.public.events.live.join.mockResolvedValue({ token: 'tok-guest', role: 'observador' });
    api.public.events.live.state.mockResolvedValue(snapshot);
    const view = mount('?pin=dev-team');

    await fireEvent.click(view.getByRole('button', { name: 'Entrar como convidado' }));

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(api.public.events.live.join).toHaveBeenCalledWith('42', {
      pinCode: 'dev-team',
      email: ''
    });
  });

  it('atualiza a tela sozinha quando chega uma mensagem do SSE', async () => {
    api.public.events.live.join.mockResolvedValue({ token: 'tok-123', role: 'participante' });
    api.public.events.live.state.mockResolvedValue(snapshot);
    const view = mount('?pin=dev-team');

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

  it('mostra erro quando o PIN está errado, sem sair da tela de identificação', async () => {
    api.public.events.live.join.mockRejectedValue(new ApiErrorLike(401, 'PIN inválido.'));
    const view = mount('?pin=errado');

    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));

    expect(await view.findByText('PIN inválido.')).toBeInTheDocument();
    expect(view.getByLabelText('E-mail')).toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();
  });

  async function joinAndWatch(view, extraSnapshot = {}) {
    api.public.events.live.join.mockResolvedValue({ token: 'tok-123', role: 'participante' });
    api.public.events.live.state.mockResolvedValue({ ...snapshot, ...extraSnapshot });
    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));
    await view.findByText('Evento Live');
  }

  it('mostra a mensagem transmitida em vez da pergunta quando o organizador definiu uma', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view, { message: 'Voltamos em 5 min' });

    expect(await view.findByText('Voltamos em 5 min')).toBeInTheDocument();
    expect(view.queryByText('Qual sua linguagem favorita?')).not.toBeInTheDocument();
  });

  it('mostra tela em branco quando não há mensagem', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view, { blanked: true });

    expect(await view.findByText('Aguarde, já voltamos…')).toBeInTheDocument();
    expect(view.queryByText('Qual sua linguagem favorita?')).not.toBeInTheDocument();
  });

  it('esconde as opções quando o organizador ativa esconder respostas, mas mantém a pergunta e os pendentes', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view, { answersHidden: true });

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(view.getByText('O organizador escondeu as respostas por enquanto.')).toBeInTheDocument();
    expect(view.queryByText('Go')).not.toBeInTheDocument();
    expect(view.queryByText('JS')).not.toBeInTheDocument();
  });

  it('mostra a pergunta normalmente quando não está em branco nem tem mensagem', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
  });

  it('mostra o nome sob cada rosto por padrão (pendente e revelado), e esconde quando o organizador ativa "Ocultar nomes"', async () => {
    const withNames = {
      pending: [{ id: 'p1', name: 'Ana Ribeiro', email: 'ana@exemplo.com', photo: '' }],
      groups: [
        { label: 'Go', participants: [{ id: 'p2', name: 'Bob Souza', email: 'bob@exemplo.com', photo: '' }] },
        { label: 'JS', participants: [] }
      ]
    };
    const view = mount('?pin=dev-team');
    await joinAndWatch(view, withNames);

    expect(await view.findByText('Ana')).toBeInTheDocument();
    expect(view.getByText('Bob')).toBeInTheDocument();

    FakeEventSource.instances[0].onmessage({
      data: JSON.stringify({ ...snapshot, ...withNames, namesHidden: true })
    });

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(view.queryByText('Ana')).not.toBeInTheDocument();
    expect(view.queryByText('Bob')).not.toBeInTheDocument();
  });

  it('mostra a barra de reações e envia ao clicar num emoji', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    await fireEvent.click(view.getByRole('button', { name: 'Reagir com 👍' }));

    expect(api.public.events.live.react).toHaveBeenCalledWith('42', 'tok-123', '👍');
  });

  it('esconde reações e Q&A quando interactionsEnabled é falso', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view, { interactionsEnabled: false });

    await view.findByText('Qual sua linguagem favorita?');
    expect(view.queryByRole('button', { name: 'Reagir com 👍' })).not.toBeInTheDocument();
    expect(view.queryByLabelText('Pergunta ou recado pro organizador')).not.toBeInTheDocument();
  });

  it('envia uma pergunta e mostra confirmação', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    await fireEvent.input(view.getByLabelText('Pergunta ou recado pro organizador'), {
      target: { value: 'Posso ir embora mais cedo?' }
    });
    await fireEvent.click(view.getByRole('button', { name: 'Enviar' }));

    expect(api.public.events.live.submitQuestion).toHaveBeenCalledWith(
      '42',
      'tok-123',
      'Posso ir embora mais cedo?'
    );
    expect(await view.findByText('Enviado!')).toBeInTheDocument();
  });

  it('mostra reação recebida via evento SSE nomeado', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    FakeEventSource.instances[0].emit('reaction', { data: JSON.stringify({ emoji: '🎉' }) });

    await waitFor(() => expect(view.getByText('🎉')).toBeInTheDocument());
  });
});
