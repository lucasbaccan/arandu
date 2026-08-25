import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import Plateia from './Plateia.svelte';

vi.mock('../lib/api.js', () => ({
  api: {
    publico: {
      eventos: {
        aoVivo: {
          entrar: vi.fn(),
          estado: vi.fn(),
          urlFluxo: vi.fn((id, token) => `/api/publico/eventos/${id}/ao-vivo/fluxo?token=${token}`),
          reagir: vi.fn().mockResolvedValue({ ok: true }),
          enviarPergunta: vi.fn().mockResolvedValue({ ok: true, messageId: 'm1' }),
          removerPergunta: vi.fn().mockResolvedValue({ ok: true })
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
  window.history.pushState({}, '', `/plateia/42${search}`);
  const target = document.createElement('div');
  document.body.appendChild(target);
  new Plateia({ target, props: { id: '42' } });
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
    expect(api.publico.eventos.aoVivo.entrar).not.toHaveBeenCalled();
    expect(api.publico.eventos.aoVivo.estado).not.toHaveBeenCalled();
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
    expect(api.publico.eventos.aoVivo.entrar).not.toHaveBeenCalled();
  });

  it('entra com sucesso via e-mail e mostra a tela da plateia sem nenhum controle de admin', async () => {
    api.publico.eventos.aoVivo.entrar.mockResolvedValue({ token: 'tok-123', role: 'participante' });
    api.publico.eventos.aoVivo.estado.mockResolvedValue(snapshot);
    const view = mount('?pin=dev-team');

    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));

    expect(await view.findByText('Evento Live')).toBeInTheDocument();
    expect(view.getByText('DEV-TEAM')).toBeInTheDocument();
    expect(view.getByLabelText('Alternar tema')).toBeInTheDocument();

    expect(api.publico.eventos.aoVivo.entrar).toHaveBeenCalledWith('42', {
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
    api.publico.eventos.aoVivo.entrar.mockResolvedValue({ token: 'tok-guest', role: 'observador' });
    api.publico.eventos.aoVivo.estado.mockResolvedValue(snapshot);
    const view = mount('?pin=dev-team');

    await fireEvent.click(view.getByRole('button', { name: 'Entrar como convidado' }));

    expect(await view.findByText('Evento Live')).toBeInTheDocument();
    expect(api.publico.eventos.aoVivo.entrar).toHaveBeenCalledWith('42', {
      pinCode: 'dev-team',
      email: ''
    });
  });

  it('atualiza a tela sozinha quando chega um snapshot do SSE', async () => {
    api.publico.eventos.aoVivo.entrar.mockResolvedValue({ token: 'tok-123', role: 'participante' });
    api.publico.eventos.aoVivo.estado.mockResolvedValue(snapshot);
    const view = mount('?pin=dev-team');

    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));
    await view.findByText('Evento Live');

    FakeEventSource.instances[0].onmessage({
      data: JSON.stringify({ ...snapshot, message: 'Voltamos em 5 min' })
    });

    await waitFor(() => expect(view.getByText('Voltamos em 5 min')).toBeInTheDocument());
  });

  it('mostra erro quando o PIN está errado, sem sair da tela de identificação', async () => {
    api.publico.eventos.aoVivo.entrar.mockRejectedValue(new ApiErrorLike(401, 'PIN inválido.'));
    const view = mount('?pin=errado');

    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));

    expect(await view.findByText('PIN inválido.')).toBeInTheDocument();
    expect(view.getByLabelText('E-mail')).toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();
  });

  async function joinAndWatch(view, extraSnapshot = {}) {
    api.publico.eventos.aoVivo.entrar.mockResolvedValue({ token: 'tok-123', role: 'participante' });
    api.publico.eventos.aoVivo.estado.mockResolvedValue({ ...snapshot, ...extraSnapshot });
    await fireEvent.input(view.getByLabelText('E-mail'), { target: { value: 'ana@exemplo.com' } });
    await fireEvent.click(view.getByRole('button', { name: 'Entrar' }));
    await view.findByText('Evento Live');
  }

  it('mostra a mensagem transmitida quando o organizador definiu uma', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view, { message: 'Voltamos em 5 min' });

    expect(await view.findByText('Voltamos em 5 min')).toBeInTheDocument();
  });

  it('mostra tela em branco quando não há mensagem', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view, { blanked: true });

    expect(await view.findByText('Aguarde, já voltamos…')).toBeInTheDocument();
  });

  it('mostra a dica e o rodapé de interação quando não está em branco nem tem mensagem', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    expect(
      await view.findByText('Reaja à apresentação ou mande uma pergunta pro apresentador.')
    ).toBeInTheDocument();
    expect(view.getByRole('button', { name: 'Reagir com 👍' })).toBeInTheDocument();
  });

  it('mostra a barra de reações e envia ao clicar num emoji', async () => {
    // Tela unificada: reações e Q&A ficam no rodapé da MESMA tela.
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    await fireEvent.click(view.getByRole('button', { name: 'Reagir com 👍' }));

    expect(api.publico.eventos.aoVivo.reagir).toHaveBeenCalledWith('42', 'tok-123', '👍');
  });

  it('tela unificada: mesmo sem ?view, mostra o menu no topo (PIN + tema), as interações e nenhum botão de modo', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    expect(await view.findByText('Evento Live')).toBeInTheDocument();
    // menu do topo: PIN e alternador de tema sempre visíveis
    expect(view.getByText('DEV-TEAM')).toBeInTheDocument();
    expect(view.getByLabelText('Alternar tema')).toBeInTheDocument();
    // interações disponíveis na mesma tela (sem depender de ?view=celular)
    expect(view.getByRole('button', { name: 'Reagir com 👍' })).toBeInTheDocument();
    // sem botões de alternância entre telão/celular
    expect(view.queryByRole('button', { name: 'Interagir' })).not.toBeInTheDocument();
    expect(view.queryByRole('button', { name: 'Modo telão' })).not.toBeInTheDocument();
  });

  it('esconde reações e Q&A quando interactionsEnabled é falso', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view, { interactionsEnabled: false });

    await view.findByText('Evento Live');
    expect(view.queryByRole('button', { name: 'Reagir com 👍' })).not.toBeInTheDocument();
    expect(view.queryByLabelText('Pergunta pro apresentador')).not.toBeInTheDocument();
  });

  it('envia uma pergunta, mostra confirmação e lista a pergunta como "minha"', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    await fireEvent.input(view.getByLabelText('Pergunta pro apresentador'), {
      target: { value: 'Posso ir embora mais cedo?' }
    });
    await fireEvent.click(view.getByRole('button', { name: 'Enviar' }));

    expect(api.publico.eventos.aoVivo.enviarPergunta).toHaveBeenCalledWith(
      '42',
      'tok-123',
      'Posso ir embora mais cedo?',
      expect.any(String) // identificador do navegador
    );
    expect(await view.findByText('Enviado!')).toBeInTheDocument();
    // a pergunta aparece na lista "Minhas perguntas" com botão de remover
    expect(await view.findByText('Posso ir embora mais cedo?')).toBeInTheDocument();
    expect(view.getByLabelText('Remover minha pergunta')).toBeInTheDocument();
  });

  it('remove a própria pergunta pela lista "Minhas perguntas"', async () => {
    api.publico.eventos.aoVivo.enviarPergunta.mockResolvedValue({ ok: true, messageId: 'm42' });
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    await fireEvent.input(view.getByLabelText('Pergunta pro apresentador'), {
      target: { value: 'Quero remover esta' }
    });
    await fireEvent.click(view.getByRole('button', { name: 'Enviar' }));
    await view.findByText('Quero remover esta');

    await fireEvent.click(view.getByLabelText('Remover minha pergunta'));

    expect(api.publico.eventos.aoVivo.removerPergunta).toHaveBeenCalledWith(
      '42',
      'tok-123',
      'm42',
      expect.any(String)
    );
    await waitFor(() => expect(view.queryByText('Quero remover esta')).not.toBeInTheDocument());
  });

  it('mostra reação recebida via evento SSE nomeado', async () => {
    const view = mount('?pin=dev-team');
    await joinAndWatch(view);

    FakeEventSource.instances[0].emit('reaction', { data: JSON.stringify({ emoji: '🎉' }) });

    await waitFor(() => expect(view.getByText('🎉')).toBeInTheDocument());
  });
});
