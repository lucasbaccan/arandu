import { describe, it, expect, vi, beforeEach } from 'vitest';
import { within } from '@testing-library/dom';
import StagePresentation from './StagePresentation.svelte';

vi.mock('../lib/api.js', () => ({
  api: {
    events: {
      live: {
        presentationState: vi.fn(),
        presentationStreamUrl: vi.fn((id) => `/api/events/${id}/live/presentation/stream`)
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
  pending: [{ id: 'p1', name: 'Ana Ribeiro', email: 'ana@exemplo.com', photo: '' }],
  groups: [
    { label: 'Go', participants: [{ id: 'p2', name: 'Bob Souza', email: 'bob@exemplo.com', photo: '' }] },
    { label: 'JS', participants: [] }
  ],
  blanked: false,
  message: '',
  interactionsEnabled: true,
  answersHidden: false,
  namesHidden: false
};

class FakeEventSource {
  constructor(url) {
    this.url = url;
    this.onmessage = null;
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

function mount() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new StagePresentation({ target, props: { id: '42' } });
  return within(target);
}

describe('Janela de apresentação somente leitura (StagePresentation)', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    vi.clearAllMocks();
    FakeEventSource.instances = [];
    global.EventSource = FakeEventSource;
  });

  it('busca o snapshot autenticado (sem PIN) e mostra a pergunta, as opções e os participantes', async () => {
    api.events.live.presentationState.mockResolvedValue(snapshot);
    const view = mount();

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(view.getByText('Go')).toBeInTheDocument();
    expect(view.getByText('JS')).toBeInTheDocument();
    expect(view.getByTitle('Bob Souza')).toBeInTheDocument();

    expect(api.events.live.presentationState).toHaveBeenCalledWith('42');
    expect(FakeEventSource.instances).toHaveLength(1);
    expect(FakeEventSource.instances[0].url).toBe('/api/events/42/live/presentation/stream');
  });

  it('não tem nenhum elemento clicável — é puramente leitura', async () => {
    api.events.live.presentationState.mockResolvedValue(snapshot);
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');

    // os rostos renderizam como <button> (mesmo componente da tela pública),
    // mas sem onFaceClick eles ficam sempre desabilitados — nada clicável.
    // O menu flutuante de variantes (PresentVariantMenu) é a única exceção
    // proposital: não mexe no estado da apresentação, só troca qual layout
    // é exibido, então fica de fora desta checagem.
    const faceButtons = view
      .getAllByRole('button')
      .filter((btn) => !btn.closest('.variant-menu'));
    expect(faceButtons.length).toBeGreaterThan(0);
    for (const btn of faceButtons) {
      expect(btn).toBeDisabled();
    }
    expect(view.queryByLabelText(/revelar|desrevelar/i)).not.toBeInTheDocument();
  });

  it('atualiza sozinha quando chega um snapshot novo via SSE', async () => {
    api.events.live.presentationState.mockResolvedValue(snapshot);
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');

    FakeEventSource.instances[0].onmessage({
      data: JSON.stringify({
        ...snapshot,
        groups: [
          { label: 'Go', participants: [{ id: 'p1', email: 'ana@exemplo.com', photo: '' }] },
          { label: 'JS', participants: [{ id: 'p2', email: 'bob@exemplo.com', photo: '' }] }
        ],
        pending: []
      })
    });

    expect(await view.findByTitle('ana@exemplo.com')).toBeInTheDocument();
  });

  it('mostra a mensagem transmitida em vez da pergunta quando o organizador definiu uma', async () => {
    api.events.live.presentationState.mockResolvedValue({ ...snapshot, message: 'Voltamos em 5 min' });
    const view = mount();

    expect(await view.findByText('Voltamos em 5 min')).toBeInTheDocument();
    expect(view.queryByText('Qual sua linguagem favorita?')).not.toBeInTheDocument();
  });

  it('mostra tela em branco quando o organizador ativa', async () => {
    api.events.live.presentationState.mockResolvedValue({ ...snapshot, blanked: true });
    const view = mount();

    expect(await view.findByText('Aguarde, já voltamos…')).toBeInTheDocument();
    expect(view.queryByText('Qual sua linguagem favorita?')).not.toBeInTheDocument();
  });

  it('esconde as opções quando "Ocultar respostas" está ativado, mas mantém a pergunta e os pendentes', async () => {
    api.events.live.presentationState.mockResolvedValue({ ...snapshot, answersHidden: true });
    const view = mount();

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(view.getByText('O organizador escondeu as respostas por enquanto.')).toBeInTheDocument();
    expect(view.queryByText('Go')).not.toBeInTheDocument();
    expect(view.getByTitle('Ana Ribeiro')).toBeInTheDocument();
  });

  it('mostra o nome sob cada rosto por padrão (pendente e revelado), e esconde quando "Ocultar nomes" está ativado', async () => {
    api.events.live.presentationState.mockResolvedValue(snapshot);
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');

    expect(view.getByText('Ana')).toBeInTheDocument();
    expect(view.getByText('Bob')).toBeInTheDocument();

    FakeEventSource.instances[0].onmessage({ data: JSON.stringify({ ...snapshot, namesHidden: true }) });

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(view.queryByText('Ana')).not.toBeInTheDocument();
    expect(view.queryByText('Bob')).not.toBeInTheDocument();
  });

  it('mostra uma reação recebida via evento SSE nomeado', async () => {
    api.events.live.presentationState.mockResolvedValue(snapshot);
    mount();
    await new Promise((r) => setTimeout(r, 0));

    expect(() =>
      FakeEventSource.instances[0].emit('reaction', { data: JSON.stringify({ emoji: '🎉' }) })
    ).not.toThrow();
  });

  it('mostra erro quando o snapshot falha ao carregar', async () => {
    api.events.live.presentationState.mockRejectedValue(new Error('Evento não encontrado.'));
    const view = mount();

    expect(await view.findByText('Evento não encontrado.')).toBeInTheDocument();
  });
});
