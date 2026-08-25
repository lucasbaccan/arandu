import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import { get } from 'svelte/store';
import ResponsesPanel from './ResponsesPanel.svelte';
import { toast } from '../lib/toastStore.js';

vi.mock('../lib/api.js', () => ({
  api: {
    eventos: {
      respostas: {
        listar: vi.fn(),
        atualizarResposta: vi.fn(),
        atualizarFoto: vi.fn()
      }
    }
  }
}));

import { api } from '../lib/api.js';

const questions = [
  {
    id: 'q1',
    title: 'Qual sua linguagem favorita?',
    type: 'GROUP',
    options: [
      { id: 'opt-go', text: 'Go' },
      { id: 'opt-js', text: 'JS' }
    ]
  },
  { id: 'q2', title: 'Qual sua comida favorita?', type: 'OPEN_TEXT', options: [] }
];

const participant = {
  id: 'p1',
  email: 'ana@exemplo.com',
  photo: '',
  editToken: 'tok123',
  createdAt: '2026-08-03T00:00:00Z',
  answers: [
    { questionId: 'q1', questionTitle: questions[0].title, questionType: 'GROUP', optionId: 'opt-go', optionText: 'Go', text: '' },
    { questionId: 'q2', questionTitle: questions[1].title, questionType: 'OPEN_TEXT', optionId: '', optionText: '', text: 'Pizza' }
  ]
};

function mount(props = {}) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new ResponsesPanel({ target, props: { eventId: '42', questions, ...props } });
  return within(target);
}

describe('ResponsesPanel', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    toast.set(null);
    vi.clearAllMocks();
  });

  it('mostra contagem zero e mensagem de vazio', async () => {
    api.eventos.respostas.listar.mockResolvedValue({ participantCount: 0, participants: [] });
    const view = mount();

    expect(await view.findByText('0 pessoas responderam')).toBeInTheDocument();
    expect(view.getByText('Ninguém respondeu ainda.')).toBeInTheDocument();
  });

  it('lista participantes e mostra respostas ao expandir', async () => {
    api.eventos.respostas.listar.mockResolvedValue({ participantCount: 1, participants: [participant] });
    const view = mount();

    expect(await view.findByText('1 pessoa respondeu')).toBeInTheDocument();
    expect(view.getByText('ana@exemplo.com')).toBeInTheDocument();
    expect(view.queryByText('Go')).not.toBeInTheDocument();

    await fireEvent.click(view.getByText('ana@exemplo.com'));

    expect(await view.findByText('Go')).toBeInTheDocument();
    expect(view.getByText('Pizza')).toBeInTheDocument();
  });

  it('permite editar uma resposta de texto aberto', async () => {
    api.eventos.respostas.listar.mockResolvedValueOnce({ participantCount: 1, participants: [participant] });
    api.eventos.respostas.atualizarResposta.mockResolvedValue({ ok: true });
    const updated = {
      ...participant,
      answers: [
        participant.answers[0],
        { ...participant.answers[1], text: '[removido]' }
      ]
    };
    api.eventos.respostas.listar.mockResolvedValueOnce({ participantCount: 1, participants: [updated] });

    const view = mount();
    await view.findByText('ana@exemplo.com');
    await fireEvent.click(view.getByText('ana@exemplo.com'));
    await view.findByText('Pizza');

    await fireEvent.click(view.getByLabelText(/Editar resposta de ana@exemplo\.com para Qual sua comida/));
    const textarea = view.getByDisplayValue('Pizza');
    await fireEvent.input(textarea, { target: { value: '[removido]' } });
    await fireEvent.click(view.getByRole('button', { name: 'Salvar' }));

    await waitFor(() =>
      expect(api.eventos.respostas.atualizarResposta).toHaveBeenCalledWith('42', 'p1', 'q2', {
        text: '[removido]'
      })
    );
    await waitFor(() => expect(get(toast)?.message).toBe('Resposta atualizada!'));
    expect(await view.findByText('[removido]')).toBeInTheDocument();
  });

  it('permite editar uma resposta de múltipla escolha', async () => {
    api.eventos.respostas.listar.mockResolvedValueOnce({ participantCount: 1, participants: [participant] });
    api.eventos.respostas.atualizarResposta.mockResolvedValue({ ok: true });
    const updated = {
      ...participant,
      answers: [
        { ...participant.answers[0], optionId: 'opt-js', optionText: 'JS' },
        participant.answers[1]
      ]
    };
    api.eventos.respostas.listar.mockResolvedValueOnce({ participantCount: 1, participants: [updated] });

    const view = mount();
    await view.findByText('ana@exemplo.com');
    await fireEvent.click(view.getByText('ana@exemplo.com'));
    await view.findByText('Go');

    await fireEvent.click(view.getByLabelText(/Editar resposta de ana@exemplo\.com para Qual sua linguagem/));
    await fireEvent.change(view.getByRole('combobox'), { target: { value: 'opt-js' } });
    await fireEvent.click(view.getByRole('button', { name: 'Salvar' }));

    await waitFor(() =>
      expect(api.eventos.respostas.atualizarResposta).toHaveBeenCalledWith('42', 'p1', 'q1', {
        optionId: 'opt-js'
      })
    );
    expect(await view.findByText('JS')).toBeInTheDocument();
  });

  it('mostra o link de edição do participante ao expandir', async () => {
    api.eventos.respostas.listar.mockResolvedValue({ participantCount: 1, participants: [participant] });
    const view = mount();
    await view.findByText('ana@exemplo.com');

    await fireEvent.click(view.getByText('ana@exemplo.com'));

    expect(await view.findByText('Link de edição:')).toBeInTheDocument();
    expect(view.getByLabelText('Copiar link de edição')).toBeInTheDocument();
  });

  it('salvar a foto sem alterar não chama a API e fecha o editor', async () => {
    // A lista devolve a foto como URL de arquivo; o cropper só emite data URL
    // quando o usuário troca a foto. Salvar sem alterar não deve mandar a URL
    // de volta para a API (quebraria a validação de data:image).
    api.eventos.respostas.listar.mockResolvedValue({ participantCount: 1, participants: [participant] });

    const view = mount();
    await view.findByText('ana@exemplo.com');

    await fireEvent.click(view.getByLabelText('Editar foto de ana@exemplo.com'));

    expect(await view.findByText('Salvar foto')).toBeInTheDocument();

    await fireEvent.click(view.getByRole('button', { name: 'Salvar foto' }));

    await waitFor(() => expect(api.eventos.respostas.atualizarFoto).not.toHaveBeenCalled());
    expect(view.queryByText('Salvar foto')).not.toBeInTheDocument();
  });

  it('mostra erro ao falhar atualização', async () => {
    api.eventos.respostas.listar.mockResolvedValue({ participantCount: 1, participants: [participant] });
    api.eventos.respostas.atualizarResposta.mockRejectedValue(new Error('Resposta muito longa.'));

    const view = mount();
    await view.findByText('ana@exemplo.com');
    await fireEvent.click(view.getByText('ana@exemplo.com'));
    await view.findByText('Pizza');

    await fireEvent.click(view.getByLabelText(/Editar resposta de ana@exemplo\.com para Qual sua comida/));
    await fireEvent.click(view.getByRole('button', { name: 'Salvar' }));

    expect(await view.findByText('Resposta muito longa.')).toBeInTheDocument();
  });
});
