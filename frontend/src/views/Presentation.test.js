import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import Presentation from './Presentation.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    events: {
      get: vi.fn(),
      questions: { list: vi.fn() },
      responses: { list: vi.fn() }
    }
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

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
  new Presentation({ target, props: { id: '42' } });
  return within(target);
}

describe('Preview da apresentação', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    vi.clearAllMocks();
    api.events.get.mockResolvedValue({ event });
    api.events.questions.list.mockResolvedValue({ questions });
    api.events.responses.list.mockResolvedValue({ participantCount: participants.length, participants });
  });

  it('carrega a primeira pergunta com todos os participantes pendentes de revelar', async () => {
    const view = mount();

    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(view.getByText('Pergunta 1 de 2')).toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de bob@exemplo.com')).toBeInTheDocument();
    expect(view.getByText('Go')).toBeInTheDocument();
    expect(view.getByText('JS')).toBeInTheDocument();
  });

  // jsdom não roda requestAnimationFrame por padrão, então a transição out:fade
  // do rosto pendente nunca conclui e o nó nunca some do DOM aqui (some normalmente
  // em um navegador real). Por isso as asserções abaixo verificam o estado revelado
  // (participante aparece na zona certa) em vez da ausência no grupo "pendente".
  it('revela um participante clicando no rosto, movendo para a opção que ele escolheu', async () => {
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');

    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));

    const goZone = view.getByText('Go').closest('.zone');
    await waitFor(() => expect(within(goZone).getByTitle('ana@exemplo.com')).toBeInTheDocument());
    expect(within(goZone).getByText('1')).toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de bob@exemplo.com')).toBeInTheDocument();
  });

  it('revelar todos move todo mundo para as respectivas opções', async () => {
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');

    await fireEvent.click(view.getByRole('button', { name: 'Revelar todos' }));

    const goZone = view.getByText('Go').closest('.zone');
    const jsZone = view.getByText('JS').closest('.zone');
    await waitFor(() => expect(within(goZone).getByTitle('ana@exemplo.com')).toBeInTheDocument());
    await waitFor(() => expect(within(jsZone).getByTitle('bob@exemplo.com')).toBeInTheDocument());
  });

  it('reiniciar revelação volta todos para pendentes', async () => {
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');
    await fireEvent.click(view.getByLabelText('Revelar resposta de ana@exemplo.com'));
    await waitFor(() => expect(view.getByRole('button', { name: 'Reiniciar revelação' })).not.toBeDisabled());

    await fireEvent.click(view.getByRole('button', { name: 'Reiniciar revelação' }));

    expect(await view.findByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
  });

  it('navega para a próxima pergunta, de resposta aberta', async () => {
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');

    await fireEvent.click(view.getByRole('button', { name: 'Próxima' }));

    expect(await view.findByText('Deixe um recado')).toBeInTheDocument();
    expect(view.getByText('Pergunta 2 de 2')).toBeInTheDocument();

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
    await view.findByText('Qual sua linguagem favorita?');
    await fireEvent.click(view.getByRole('button', { name: 'Próxima' }));
    await view.findByText('Deixe um recado');

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
    await view.findByText('Qual sua linguagem favorita?');

    await fireEvent.click(view.getByRole('button', { name: 'Sair do preview' }));

    expect(navigate).toHaveBeenCalledWith('/events/42');
  });

  it('mostra estado vazio quando o evento não tem perguntas', async () => {
    api.events.questions.list.mockResolvedValue({ questions: [] });
    const view = mount();

    expect(await view.findByText('Este evento ainda não tem perguntas.')).toBeInTheDocument();
  });

  it('modo apresentação esconde os controles e números de admin, mas mantém pergunta e respostas', async () => {
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');

    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));

    expect(view.queryByRole('button', { name: 'Sair do preview' })).not.toBeInTheDocument();
    expect(view.queryByText('pessoas responderam', { exact: false })).not.toBeInTheDocument();
    expect(view.queryByText('Pergunta 1 de 2')).not.toBeInTheDocument();
    expect(view.queryByRole('button', { name: 'Revelar todos' })).not.toBeInTheDocument();
    expect(view.queryByRole('button', { name: 'Reiniciar revelação' })).not.toBeInTheDocument();
    expect(view.queryByRole('button', { name: 'Anterior' })).not.toBeInTheDocument();
    expect(view.queryByRole('button', { name: 'Próxima' })).not.toBeInTheDocument();

    // conteúdo essencial continua visível e clicável
    expect(view.getByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(view.getByLabelText('Revelar resposta de ana@exemplo.com')).toBeInTheDocument();
    expect(view.getByText('Go')).toBeInTheDocument();
  });

  it('Esc sai do modo apresentação e volta os controles', async () => {
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');
    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));
    expect(view.queryByRole('button', { name: 'Próxima' })).not.toBeInTheDocument();

    await fireEvent.keyDown(window, { key: 'Escape' });

    expect(await view.findByRole('button', { name: 'Próxima' })).toBeInTheDocument();
  });

  it('setas do teclado navegam entre perguntas mesmo no modo apresentação', async () => {
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');
    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));

    await fireEvent.keyDown(window, { key: 'ArrowRight' });
    expect(await view.findByText('Deixe um recado')).toBeInTheDocument();

    await fireEvent.keyDown(window, { key: 'ArrowLeft' });
    expect(await view.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
  });

  it('tecla R revela todos mesmo no modo apresentação', async () => {
    const view = mount();
    await view.findByText('Qual sua linguagem favorita?');
    await fireEvent.click(view.getByRole('button', { name: 'Modo apresentação' }));

    await fireEvent.keyDown(window, { key: 'r' });

    const goZone = view.getByText('Go').closest('.zone');
    await waitFor(() => expect(within(goZone).getByTitle('ana@exemplo.com')).toBeInTheDocument());
  });
});
