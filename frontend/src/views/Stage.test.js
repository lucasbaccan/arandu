import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import Stage from './Stage.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    events: {
      get: vi.fn(),
      questions: { list: vi.fn() },
      responses: { list: vi.fn() },
      live: {
        setQuestion: vi.fn().mockResolvedValue({ ok: true }),
        reveal: vi.fn().mockResolvedValue({ ok: true }),
        unreveal: vi.fn().mockResolvedValue({ ok: true }),
        revealAll: vi.fn().mockResolvedValue({ ok: true }),
        reset: vi.fn().mockResolvedValue({ ok: true }),
        setBlanked: vi.fn().mockResolvedValue({ ok: true }),
        setAnswersHidden: vi.fn().mockResolvedValue({ ok: true }),
        setMessage: vi.fn().mockResolvedValue({ ok: true }),
        setInteractionsEnabled: vi.fn().mockResolvedValue({ ok: true }),
        adminStreamUrl: vi.fn((id) => `/api/events/${id}/live/stream`),
        dismissQA: vi.fn().mockResolvedValue({ ok: true })
      }
    }
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

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

const adminSnapshot = {
  blanked: false,
  answersHidden: false,
  message: '',
  interactionsEnabled: true,
  qaInbox: []
};

const event = { id: '42', title: 'Conecta DevOps', pinCode: '123456', status: 'PRESENTING' };

const questions = [
  {
    id: 'q1',
    title: 'Qual sua linguagem favorita?',
    type: 'GROUP',
    layoutView: 'TIMELINE',
    orderIndex: 0,
    options: [
      { id: 'o1', text: 'Go' },
      { id: 'o2', text: 'JS' }
    ]
  },
  {
    id: 'q2',
    title: 'Deixe um recado',
    type: 'OPEN_TEXT',
    layoutView: 'TIMELINE',
    orderIndex: 1,
    options: []
  }
];

const participants = [
  {
    id: 'p1',
    email: 'ana@exemplo.com',
    photo: '',
    editToken: 't1',
    createdAt: '2026-08-03T00:00:00Z',
    answers: [
      { questionId: 'q1', questionTitle: questions[0].title, questionType: 'GROUP', optionId: 'o1', optionText: 'Go', text: '' },
      { questionId: 'q2', questionTitle: questions[1].title, questionType: 'OPEN_TEXT', optionId: '', optionText: '', text: 'Boa sorte!' }
    ]
  },
  {
    id: 'p2',
    email: 'bob@exemplo.com',
    photo: '',
    editToken: 't2',
    createdAt: '2026-08-03T00:05:00Z',
    answers: [
      { questionId: 'q1', questionTitle: questions[0].title, questionType: 'GROUP', optionId: 'o2', optionText: 'JS', text: '' },
      { questionId: 'q2', questionTitle: questions[1].title, questionType: 'OPEN_TEXT', optionId: '', optionText: '', text: 'Valeu!' }
    ]
  },
  {
    id: 'p3',
    email: 'carla@exemplo.com',
    photo: '',
    editToken: 't3',
    createdAt: '2026-08-03T00:10:00Z',
    answers: [
      { questionId: 'q1', questionTitle: questions[0].title, questionType: 'GROUP', optionId: 'o1', optionText: 'Go', text: '' },
      { questionId: 'q2', questionTitle: questions[1].title, questionType: 'OPEN_TEXT', optionId: '', optionText: '', text: '  boa sorte!  ' }
    ]
  }
];

function mount() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new Stage({ target, props: { id: '42' } });
  return within(target);
}

describe('Preview da apresentação', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    vi.clearAllMocks();
    FakeEventSource.instances = [];
    global.EventSource = FakeEventSource;
    api.events.get.mockResolvedValue({ event });
    api.events.questions.list.mockResolvedValue({ questions });
    api.events.responses.list.mockResolvedValue({ participantCount: participants.length, participants });
  });

  it('carrega a primeira pergunta com todos os participantes pendentes de revelar', async () => {
    const view = mount();

    expect(await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' })).toBeInTheDocument();
    expect(view.getByText('1 / 2')).toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de bob@exemplo.com')).toBeInTheDocument();
    expect(view.getByText('Go')).toBeInTheDocument();
    expect(view.getByText('JS')).toBeInTheDocument();
  });

  it('pula direto pra uma pergunta clicando nela no trilho', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByText('Deixe um recado'));

    expect(await view.findByText('2 / 2')).toBeInTheDocument();
    expect(api.events.live.setQuestion).toHaveBeenCalledWith('42', 'q2');
  });

  // jsdom não roda requestAnimationFrame por padrão, então a transição out:fade
  // do rosto pendente nunca conclui e o nó nunca some do DOM aqui (some normalmente
  // em um navegador real). Por isso as asserções abaixo verificam o estado revelado
  // (participante aparece na zona certa) em vez da ausência no grupo "pendente".
  it('revela um participante clicando no rosto, movendo para a opção que ele escolheu', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));

    const goZone = view.getByText('Go').closest('.zone');
    await waitFor(() => expect(within(goZone).getByTitle('ana@exemplo.com')).toBeInTheDocument());
    expect(within(goZone).getByText('1')).toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de bob@exemplo.com')).toBeInTheDocument();
  });

  it('clicar de novo no rosto revelado desrevela e volta pra pendente', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));
    const revealedFace = await view.findByLabelText('Desrevelar resposta de ana@exemplo.com');

    await fireEvent.click(revealedFace);

    expect(api.events.live.unreveal).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(String),
      expect.any(String)
    );
    expect(await view.findByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
    expect(view.queryByLabelText('Desrevelar resposta de ana@exemplo.com')).not.toBeInTheDocument();
  });

  it('revelar todos move todo mundo para as respectivas opções', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByRole('button', { name: 'Revelar tudo' }));

    const goZone = view.getByText('Go').closest('.zone');
    const jsZone = view.getByText('JS').closest('.zone');
    await waitFor(() => expect(within(goZone).getByTitle('ana@exemplo.com')).toBeInTheDocument());
    await waitFor(() => expect(within(jsZone).getByTitle('bob@exemplo.com')).toBeInTheDocument());
    // drena o timer escalonado da 3ª pessoa (carla, delay 300ms) antes de
    // sair do teste — senão ele dispara durante um teste seguinte, contra um
    // componente já órfão.
    await waitFor(() => expect(within(goZone).getByTitle('carla@exemplo.com')).toBeInTheDocument());
  });

  it('reiniciar revelação volta todos para pendentes', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });
    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));
    await waitFor(() => expect(view.getByRole('button', { name: 'Reiniciar' })).not.toBeDisabled());

    await fireEvent.click(view.getByRole('button', { name: 'Reiniciar' }));

    expect(await view.findByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
  });

  it('navega para a próxima pergunta, de resposta aberta', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByLabelText('Próxima pergunta'));

    expect(await view.findByRole('heading', { name: 'Deixe um recado' })).toBeInTheDocument();
    expect(view.getByText('2 / 2')).toBeInTheDocument();

    // as caixas de resposta já aparecem antes de qualquer revelação
    const bubble = view.getByText('Boa sorte!').closest('.zone');
    expect(within(bubble).getByText('0')).toBeInTheDocument();
    expect(within(bubble).queryByTitle('ana@exemplo.com')).not.toBeInTheDocument();

    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));
    await waitFor(() => expect(within(bubble).getByTitle('ana@exemplo.com')).toBeInTheDocument());
    expect(within(bubble).getByText('1')).toBeInTheDocument();
  });

  it('agrupa respostas abertas iguais (após trim e capitalização) num único balão', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });
    await fireEvent.click(view.getByLabelText('Próxima pergunta'));
    await view.findByRole('heading', { name: 'Deixe um recado' });

    // ana respondeu "Boa sorte!" e carla respondeu "  boa sorte!  " —
    // devem cair no mesmo balão, já normalizado.
    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));
    await fireEvent.click(view.getByLabelText('Revelar resposta de carla@exemplo.com'));

    const bubble = (await view.findByText('Boa sorte!')).closest('.zone');
    expect(within(bubble).getByTitle('ana@exemplo.com')).toBeInTheDocument();
    expect(within(bubble).getByTitle('carla@exemplo.com')).toBeInTheDocument();
    expect(view.getAllByText('Boa sorte!')).toHaveLength(1);
  });

  it('sai do preview voltando para a tela do evento', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByLabelText('Voltar para o evento'));

    expect(navigate).toHaveBeenCalledWith('/events/42');
  });

  it('mostra estado vazio quando o evento não tem perguntas', async () => {
    api.events.questions.list.mockResolvedValue({ questions: [] });
    const view = mount();

    expect(await view.findByText('Este evento ainda não tem perguntas.')).toBeInTheDocument();
  });

  it('modo apresentação esconde os controles e números de admin, mas mantém pergunta e respostas', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));

    expect(view.queryByLabelText('Voltar para o evento')).not.toBeInTheDocument();
    expect(view.queryByText('pessoas responderam', { exact: false })).not.toBeInTheDocument();
    expect(view.queryByText('1 / 2')).not.toBeInTheDocument();
    expect(view.queryByRole('button', { name: 'Revelar tudo' })).not.toBeInTheDocument();
    expect(view.queryByRole('button', { name: 'Reiniciar' })).not.toBeInTheDocument();
    expect(view.queryByLabelText('Pergunta anterior')).not.toBeInTheDocument();
    expect(view.queryByLabelText('Próxima pergunta')).not.toBeInTheDocument();

    // conteúdo essencial continua visível e clicável
    expect(view.getByRole('heading', { name: 'Qual sua linguagem favorita?' })).toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
    expect(view.getByText('Go')).toBeInTheDocument();
  });

  it('Esc sai do modo apresentação e volta os controles', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });
    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));
    expect(view.queryByLabelText('Próxima pergunta')).not.toBeInTheDocument();

    await fireEvent.keyDown(window, { key: 'Escape' });

    expect(await view.findByLabelText('Próxima pergunta')).toBeInTheDocument();
  });

  it('setas do teclado navegam entre perguntas mesmo no modo apresentação', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });
    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));

    await fireEvent.keyDown(window, { key: 'ArrowRight' });
    expect(await view.findByText('Deixe um recado')).toBeInTheDocument();

    await fireEvent.keyDown(window, { key: 'ArrowLeft' });
    expect(await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' })).toBeInTheDocument();
  });

  it('tecla R revela todos mesmo no modo apresentação', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });
    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));

    await fireEvent.keyDown(window, { key: 'r' });

    const goZone = view.getByText('Go').closest('.zone');
    const jsZone = view.getByText('JS').closest('.zone');
    await waitFor(() => expect(within(goZone).getByTitle('ana@exemplo.com')).toBeInTheDocument());
    // drena os timers escalonados de bob (150ms) e carla (300ms) antes de
    // sair do teste — senão disparam durante um teste seguinte.
    await waitFor(() => expect(within(jsZone).getByTitle('bob@exemplo.com')).toBeInTheDocument());
    await waitFor(() => expect(within(goZone).getByTitle('carla@exemplo.com')).toBeInTheDocument());
  });

  it('abre a tela de apresentação numa nova aba com o PIN na URL', async () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => {});
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByRole('button', { name: 'Abrir tela de apresentação' }));

    expect(openSpy).toHaveBeenCalledWith('/audience/42?pin=123456', '_blank');
    openSpy.mockRestore();
  });

  it('alterna tela em branco e chama a API', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getAllByRole('switch')[0]);

    expect(api.events.live.setBlanked).toHaveBeenCalledWith('42', true);
  });

  it('alterna interações e chama a API', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const switches = view.getAllByRole('switch');
    await fireEvent.click(switches[2]);

    expect(api.events.live.setInteractionsEnabled).toHaveBeenCalledWith('42', false);
  });

  it('alterna esconder respostas da plateia e chama a API', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const switches = view.getAllByRole('switch');
    await fireEvent.click(switches[1]);

    expect(api.events.live.setAnswersHidden).toHaveBeenCalledWith('42', true);
  });

  it('envia e limpa uma mensagem de aviso', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.input(view.getByLabelText('Aviso pra tela dos participantes'), {
      target: { value: 'Voltamos em 5 min' }
    });
    await fireEvent.click(view.getByRole('button', { name: 'Enviar aviso' }));
    expect(api.events.live.setMessage).toHaveBeenCalledWith('42', 'Voltamos em 5 min');

    await fireEvent.click(view.getByRole('button', { name: 'Limpar' }));
    expect(api.events.live.setMessage).toHaveBeenCalledWith('42', '');
  });

  it('espaço, R e P digitados no campo de aviso não disparam os atalhos globais', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const input = view.getByLabelText('Aviso pra tela dos participantes');
    await fireEvent.keyDown(input, { key: ' ' });
    await fireEvent.keyDown(input, { key: 'r' });
    await fireEvent.keyDown(input, { key: 'p' });

    // espaço não deveria ter avançado pra próxima pergunta (goNext)
    expect(view.getByRole('heading', { name: 'Qual sua linguagem favorita?' })).toBeInTheDocument();
    expect(view.queryByRole('heading', { name: 'Deixe um recado' })).not.toBeInTheDocument();
    // R não deveria ter revelado ninguém
    expect(view.getByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
    // P não deveria ter entrado em modo apresentação
    expect(view.getByLabelText('Voltar para o evento')).toBeInTheDocument();
  });

  it('mostra e dispensa mensagens da caixa de Q&A', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    FakeEventSource.instances[0].onmessage({
      data: JSON.stringify({
        ...adminSnapshot,
        qaInbox: [{ id: 'm1', email: 'ana@exemplo.com', text: 'oi', createdAt: '2026-08-03T00:00:00Z' }]
      })
    });

    expect(await view.findByText('oi')).toBeInTheDocument();
    expect(view.getByText('ana@exemplo.com')).toBeInTheDocument();

    await fireEvent.click(view.getByRole('button', { name: 'Dispensar' }));

    expect(api.events.live.dismissQA).toHaveBeenCalledWith('42', 'm1');
    expect(view.queryByText('oi')).not.toBeInTheDocument();
  });

  it('recebe reações via SSE', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    FakeEventSource.instances[0].emit('reaction', { data: JSON.stringify({ emoji: '🎉' }) });

    await waitFor(() => expect(view.getByText('🎉')).toBeInTheDocument());
  });
});
