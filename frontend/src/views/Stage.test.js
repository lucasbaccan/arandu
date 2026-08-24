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
        resetAll: vi.fn().mockResolvedValue({ ok: true }),
        setBlanked: vi.fn().mockResolvedValue({ ok: true }),
        setAnswersHidden: vi.fn().mockResolvedValue({ ok: true }),
        setNamesHidden: vi.fn().mockResolvedValue({ ok: true }),
        setDensityMode: vi.fn().mockResolvedValue({ ok: true }),
        setMessage: vi.fn().mockResolvedValue({ ok: true }),
        setInteractionsEnabled: vi.fn().mockResolvedValue({ ok: true }),
        adminState: vi.fn(),
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
  namesHidden: false,
  message: '',
  interactionsEnabled: true,
  currentQuestionId: '',
  revealed: {},
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
    api.events.live.adminState.mockResolvedValue(adminSnapshot);
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

  it('retoma na pergunta e na revelação salvas no servidor, sem forçar a pergunta 1', async () => {
    api.events.live.adminState.mockResolvedValue({
      ...adminSnapshot,
      currentQuestionId: 'q2',
      revealed: { q2: ['p1'] }
    });
    const view = mount();

    expect(await view.findByRole('heading', { name: 'Deixe um recado' })).toBeInTheDocument();
    expect(view.getByText('2 / 2')).toBeInTheDocument();
    // ana (p1) já revelada — não deveria mais estar como pendente
    expect(view.queryByLabelText('Revelar resposta de ana@exemplo.com')).not.toBeInTheDocument();
    // ao retomar de um estado já em andamento, não deve forçar de volta pra pergunta 1
    expect(api.events.live.setQuestion).not.toHaveBeenCalled();
  });

  it('reinicia a revelação de todas as perguntas ao clicar em "Reiniciar tudo"', async () => {
    api.events.live.adminState.mockResolvedValue({
      ...adminSnapshot,
      currentQuestionId: 'q1',
      revealed: { q1: ['p1'] }
    });
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    expect(view.getByRole('button', { name: 'Reiniciar tudo' })).not.toBeDisabled();

    await fireEvent.click(view.getByRole('button', { name: 'Reiniciar tudo' }));

    expect(api.events.live.resetAll).toHaveBeenCalledWith('42');
    expect(await view.findByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
  });

  it('desabilita "Reiniciar tudo" quando ninguém foi revelado em nenhuma pergunta', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    expect(view.getByRole('button', { name: 'Reiniciar tudo' })).toBeDisabled();
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
    await waitFor(() => expect(within(goZone).getByLabelText('Desrevelar resposta de ana@exemplo.com')).toBeInTheDocument());
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
    await waitFor(() => expect(within(goZone).getByLabelText('Desrevelar resposta de ana@exemplo.com')).toBeInTheDocument());
    await waitFor(() => expect(within(jsZone).getByLabelText('Desrevelar resposta de bob@exemplo.com')).toBeInTheDocument());
    // drena o timer escalonado da 3ª pessoa (carla, delay 300ms) antes de
    // sair do teste — senão ele dispara durante um teste seguinte, contra um
    // componente já órfão.
    await waitFor(() => expect(within(goZone).getByLabelText('Desrevelar resposta de carla@exemplo.com')).toBeInTheDocument());
  });

  it('reiniciar revelação volta todos para pendentes', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });
    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));
    await waitFor(() =>
      expect(view.getByRole('button', { name: 'Reiniciar pergunta' })).not.toBeDisabled()
    );

    await fireEvent.click(view.getByRole('button', { name: 'Reiniciar pergunta' }));

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
    expect(within(bubble).queryByLabelText('Revelar resposta de ana@exemplo.com')).not.toBeInTheDocument();

    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));
    await waitFor(() => expect(within(bubble).getByLabelText('Desrevelar resposta de ana@exemplo.com')).toBeInTheDocument());
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
    expect(within(bubble).getByLabelText('Desrevelar resposta de ana@exemplo.com')).toBeInTheDocument();
    expect(within(bubble).getByLabelText('Desrevelar resposta de carla@exemplo.com')).toBeInTheDocument();
    expect(view.getAllByText('Boa sorte!')).toHaveLength(1);
  });

  it('sai do preview voltando para a tela do evento', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    // O caminho de volta agora é o breadcrumb do shell do organizador.
    await fireEvent.click(view.getByRole('link', { name: 'Conecta DevOps' }));

    expect(navigate).toHaveBeenCalledWith('/events/42');
  });

  it('mostra estado vazio quando o evento não tem perguntas', async () => {
    api.events.questions.list.mockResolvedValue({ questions: [] });
    const view = mount();

    expect(await view.findByText('Este evento ainda não tem perguntas.')).toBeInTheDocument();
  });

  it('abre a janela de apresentação somente leitura numa janela nomeada', async () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => ({ closed: false }));
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));

    expect(openSpy).toHaveBeenCalledWith('/stage/42/present', 'arandu-present-42', 'width=1366,height=768');
    // não é toggle de estado local — os controles de admin continuam aqui
    expect(view.getByLabelText('Próxima pergunta')).toBeInTheDocument();
    openSpy.mockRestore();
  });

  it('os botões de modo no rodapé mudam a densidade ao vivo (sem navegar) e abrem/focam a janela de apresentação', async () => {
    const fakeWindow = { closed: false, focus: vi.fn() };
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => fakeWindow);
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const smartBtn = view.getByRole('button', { name: 'Modo apresentação — Smart' });
    await fireEvent.click(smartBtn);

    // muda o estado no servidor — quem já tiver /present ou /audience
    // abertos vê ao vivo, sem precisar recarregar nem trocar de URL.
    expect(api.events.live.setDensityMode).toHaveBeenCalledWith('42', 'smart');
    expect(openSpy).toHaveBeenCalledWith('/stage/42/present', 'arandu-present-42', 'width=1366,height=768');
    expect(openSpy).toHaveBeenCalledTimes(1);
    expect(smartBtn).toHaveAttribute('aria-pressed', 'true');

    const cols2Btn = view.getByRole('button', { name: 'Modo apresentação — 2 colunas' });
    await fireEvent.click(cols2Btn);

    expect(api.events.live.setDensityMode).toHaveBeenCalledWith('42', '2');
    // janela já aberta: só foca, não abre de novo
    expect(openSpy).toHaveBeenCalledTimes(1);
    expect(fakeWindow.focus).toHaveBeenCalledTimes(1);
    expect(cols2Btn).toHaveAttribute('aria-pressed', 'true');
    expect(smartBtn).toHaveAttribute('aria-pressed', 'false');

    openSpy.mockRestore();
  });

  it('setas do teclado navegam entre perguntas', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.keyDown(window, { key: 'ArrowRight' });
    expect(await view.findByRole('heading', { name: 'Deixe um recado' })).toBeInTheDocument();

    await fireEvent.keyDown(window, { key: 'ArrowLeft' });
    expect(await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' })).toBeInTheDocument();
  });

  it('tecla R revela todos', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.keyDown(window, { key: 'r' });

    const goZone = view.getByText('Go').closest('.zone');
    const jsZone = view.getByText('JS').closest('.zone');
    await waitFor(() => expect(within(goZone).getByLabelText('Desrevelar resposta de ana@exemplo.com')).toBeInTheDocument());
    // drena os timers escalonados de bob (150ms) e carla (300ms) antes de
    // sair do teste — senão disparam durante um teste seguinte.
    await waitFor(() => expect(within(jsZone).getByLabelText('Desrevelar resposta de bob@exemplo.com')).toBeInTheDocument());
    await waitFor(() => expect(within(goZone).getByLabelText('Desrevelar resposta de carla@exemplo.com')).toBeInTheDocument());
  });

  it('abre a tela de apresentação numa nova aba com o PIN na URL', async () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => {});
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByRole('button', { name: 'Abrir tela da plateia' }));

    expect(openSpy).toHaveBeenCalledWith('/audience/42?pin=123456&view=telao', '_blank');
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
    await fireEvent.click(switches[3]);

    expect(api.events.live.setInteractionsEnabled).toHaveBeenCalledWith('42', false);
  });

  it('alterna esconder respostas da plateia e chama a API', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const switches = view.getAllByRole('switch');
    await fireEvent.click(switches[1]);

    expect(api.events.live.setAnswersHidden).toHaveBeenCalledWith('42', true);
  });

  it('alterna ocultar nomes e chama a API, mas mantém a legenda de nome visível em /stage', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const pendingWrap = view.getByLabelText('Revelar resposta de ana@exemplo.com').closest('.face-wrap');
    expect(within(pendingWrap).getByText('ana@exemplo.com')).toBeInTheDocument();

    const switches = view.getAllByRole('switch');
    await fireEvent.click(switches[2]);

    // é estado do servidor agora (pra sincronizar com /present), mas /stage
    // sempre mostra os nomes — o toggle não afeta a própria tela do admin
    expect(api.events.live.setNamesHidden).toHaveBeenCalledWith('42', true);
    expect(within(pendingWrap).getByText('ana@exemplo.com')).toBeInTheDocument();
  });

  it('envia e limpa uma mensagem de aviso', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByRole('tab', { name: 'Aviso' }));
    await fireEvent.input(view.getByLabelText('Aviso pra tela dos participantes'), {
      target: { value: 'Voltamos em 5 min' }
    });
    await fireEvent.click(view.getByRole('button', { name: 'Enviar aviso' }));
    expect(api.events.live.setMessage).toHaveBeenCalledWith('42', 'Voltamos em 5 min');

    await fireEvent.click(view.getByRole('button', { name: 'Limpar' }));
    expect(api.events.live.setMessage).toHaveBeenCalledWith('42', '');
  });

  it('espaço e R digitados no campo de aviso não disparam os atalhos globais', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByRole('tab', { name: 'Aviso' }));
    const input = view.getByLabelText('Aviso pra tela dos participantes');
    await fireEvent.keyDown(input, { key: ' ' });
    await fireEvent.keyDown(input, { key: 'r' });

    // espaço não deveria ter avançado pra próxima pergunta (goNext)
    expect(view.getByRole('heading', { name: 'Qual sua linguagem favorita?' })).toBeInTheDocument();
    expect(view.queryByRole('heading', { name: 'Deixe um recado' })).not.toBeInTheDocument();
    // R não deveria ter revelado ninguém
    expect(view.getByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
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

    await fireEvent.click(view.getByRole('tab', { name: 'Q&A' }));
    const qaItem = (await view.findByText('oi')).closest('.qa-item');
    expect(within(qaItem).getByText('ana@exemplo.com')).toBeInTheDocument();

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

  it('alterna as dicas de hover pelo botão "Dicas" do CrumbBar', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const toggle = view.getByRole('button', { name: 'Dicas' });
    expect(toggle).toHaveAttribute('aria-pressed', 'true');

    await fireEvent.click(toggle);
    expect(toggle).toHaveAttribute('aria-pressed', 'false');
  });

  it('hover no rosto pendente mostra a resposta da pessoa após ~0,6s (e some ao sair)', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const wrap = view.getByLabelText('Revelar resposta de ana@exemplo.com').closest('.face-wrap');
    await fireEvent.mouseEnter(wrap);

    // antes do atraso de 0,6s nada aparece
    expect(view.queryByRole('tooltip')).not.toBeInTheDocument();

    const tip = await view.findByRole('tooltip', {}, { timeout: 1500 });
    expect(within(tip).getByText('ana@exemplo.com')).toBeInTheDocument();
    expect(within(tip).getByText('Go')).toBeInTheDocument();

    await fireEvent.mouseLeave(wrap);
    await waitFor(() => expect(view.queryByRole('tooltip')).not.toBeInTheDocument());
  });

  it('hover no rosto revelado mostra a resposta da pessoa', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));
    const revealed = await view.findByLabelText('Desrevelar resposta de ana@exemplo.com');

    await fireEvent.mouseEnter(revealed.closest('.person-wrap'));

    const tip = await view.findByRole('tooltip', {}, { timeout: 1500 });
    expect(within(tip).getByText('ana@exemplo.com')).toBeInTheDocument();
    expect(within(tip).getByText('Go')).toBeInTheDocument();
  });

  it('hover no rótulo de uma resposta mostra quem está (e quem ainda vai cair) nela', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const label = view.getByText('Go').closest('.zone-label');
    await fireEvent.mouseEnter(label);

    // ana e carla escolheram Go; nenhuma revelada ainda → ambas aparecem como
    // pendentes no tooltip e a fila lá em cima ganha o anel de destaque.
    await waitFor(
      () => expect(view.getByText('2 pessoas nesta resposta · 2 ainda pendentes')).toBeInTheDocument(),
      { timeout: 1500 }
    );
    const tip = view.getByRole('tooltip');
    expect(within(tip).getByText('ana@exemplo.com')).toBeInTheDocument();
    expect(within(tip).getByText('carla@exemplo.com')).toBeInTheDocument();
    // ninguém revelado ainda → ambas estão com o estado "pendente" no popup
    // (borda tracejada + nome esmaecido)
    expect(
      within(tip)
        .getByRole('button', { name: 'Revelar resposta de ana@exemplo.com' })
        .classList.contains('pending')
    ).toBe(true);
    expect(
      within(tip)
        .getByRole('button', { name: 'Revelar resposta de carla@exemplo.com' })
        .classList.contains('pending')
    ).toBe(true);

    const pendingRow = document.body.querySelector('.pending-row');
    const faceWrap = within(pendingRow)
      .getByLabelText('Revelar resposta de ana@exemplo.com')
      .closest('.face-wrap');
    expect(faceWrap.classList.contains('zone-hinted')).toBe(true);

    await fireEvent.mouseLeave(label);
    await waitFor(() => expect(view.queryByRole('tooltip')).not.toBeInTheDocument());
  });


  it('revela uma pessoa clicando nela no popup da resposta (hover da pergunta)', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const label = view.getByText('Go').closest('.zone-label');
    await fireEvent.mouseEnter(label);

    const tip = await view.findByRole('tooltip', {}, { timeout: 1500 });
    await fireEvent.click(
      within(tip).getByRole('button', { name: 'Revelar resposta de ana@exemplo.com' })
    );

    expect(api.events.live.reveal).toHaveBeenCalledWith('42', 'q1', 'p1');

    // revelou: ana sai da fila e entra na zona; o popup continua aberto e o
    // botão dela agora vira "desrevelar", com o rodapé atualizado.
    const goZone = within(document.body.querySelector('.zones')).getByText('Go').closest('.zone');
    await waitFor(() =>
      expect(within(goZone).getByLabelText('Desrevelar resposta de ana@exemplo.com')).toBeInTheDocument()
    );
    expect(view.getByRole('tooltip')).toBeInTheDocument();
    expect(
      within(view.getByRole('tooltip')).getByRole('button', {
        name: 'Desrevelar resposta de ana@exemplo.com'
      })
    ).toBeInTheDocument();
    expect(view.getByText('2 pessoas nesta resposta · 1 ainda pendente')).toBeInTheDocument();

    // o popup reflete a revelação: ana sai do estado "pendente" (nome cheio),
    // carla continua pendente (borda tracejada + nome esmaecido)
    const tip2 = view.getByRole('tooltip');
    expect(
      within(tip2)
        .getByRole('button', { name: 'Desrevelar resposta de ana@exemplo.com' })
        .classList.contains('pending')
    ).toBe(false);
    expect(
      within(tip2)
        .getByRole('button', { name: 'Revelar resposta de carla@exemplo.com' })
        .classList.contains('pending')
    ).toBe(true);
  });


  it('revela todos DAQUELA resposta pelo mini botão do popup (sem tocar nas outras)', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    const label = view.getByText('Go').closest('.zone-label');
    await fireEvent.mouseEnter(label);

    const tip = await view.findByRole('tooltip', {}, { timeout: 1500 });
    const revealBtn = within(tip).getByRole('button', { name: 'Revelar todos desta resposta' });
    expect(revealBtn).not.toBeDisabled();

    await fireEvent.click(revealBtn);

    // só quem respondeu "Go" (ana e carla) é revelado; bob (JS) continua pendente
    const zonesEl = document.body.querySelector('.zones');
    const goZone = within(zonesEl).getByText('Go').closest('.zone');
    const jsZone = within(zonesEl).getByText('JS').closest('.zone');
    await waitFor(() =>
      expect(within(goZone).getByLabelText('Desrevelar resposta de ana@exemplo.com')).toBeInTheDocument()
    );
    await waitFor(() =>
      expect(within(goZone).getByLabelText('Desrevelar resposta de carla@exemplo.com')).toBeInTheDocument()
    );
    expect(
      within(jsZone).queryByLabelText('Desrevelar resposta de bob@exemplo.com')
    ).not.toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de bob@exemplo.com')).toBeInTheDocument();

    // popup continua aberto, sem pendentes na resposta → botão desabilitado
    const tip2 = view.getByRole('tooltip');
    expect(within(tip2).getByText('2 pessoas nesta resposta')).toBeInTheDocument();
    await waitFor(() =>
      expect(within(tip2).getByRole('button', { name: 'Revelar todos desta resposta' })).toBeDisabled()
    );
  });

  it('com as dicas desligadas, o hover não mostra tooltip', async () => {
    const view = mount();
    await view.findByRole('heading', { name: 'Qual sua linguagem favorita?' });

    await fireEvent.click(view.getByRole('button', { name: 'Dicas' }));

    const wrap = view.getByLabelText('Revelar resposta de ana@exemplo.com').closest('.face-wrap');
    await fireEvent.mouseEnter(wrap);

    await new Promise((r) => setTimeout(r, 700));
    expect(view.queryByRole('tooltip')).not.toBeInTheDocument();
  });
});
