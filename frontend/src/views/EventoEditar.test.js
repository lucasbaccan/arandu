import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, within, waitFor } from '@testing-library/dom';
import userEvent from '@testing-library/user-event';
import { get } from 'svelte/store';
import EventoEditar from './EventoEditar.svelte';
import { toast } from '../lib/toastStore.js';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    eventos: {
      buscar: vi.fn(),
      atualizar: vi.fn(),
      deletar: vi.fn(),
      perguntas: {
        listar: vi.fn(),
        criar: vi.fn(),
        atualizar: vi.fn(),
        reordenar: vi.fn(),
        remover: vi.fn()
      },
      respostas: {
        listar: vi.fn(),
        atualizarResposta: vi.fn()
      }
    }
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

const event = {
  id: '42',
  title: 'Conecta DevOps',
  pinCode: '123456',
  status: 'PREPARATION',
  createdAt: '2026-08-02T00:00:00Z'
};

function question(id, title) {
  return {
    id,
    eventId: '42',
    title,
    type: 'GROUP',
    orderIndex: 0,
    options: [
      { id: `${id}-1`, text: 'A' },
      { id: `${id}-2`, text: 'B' }
    ]
  };
}

function mount(questions = [], participantCount = 0) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  api.eventos.perguntas.listar.mockResolvedValue({ questions });
  api.eventos.respostas.listar.mockResolvedValue({ participantCount, participants: [] });
  new EventoEditar({ target, props: { id: '42' } });
  return within(target);
}

describe('Editar evento', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    toast.set(null);
    vi.clearAllMocks();
    // A tela grava aba/participante selecionado na URL via history.replaceState;
    // no jsdom o window.location é compartilhado entre os testes do arquivo,
    // então sem resetar aqui um teste que muda de aba/participante vaza pro
    // próximo (que herda uma aba/participante que não pediu).
    window.history.replaceState({}, '', '/');
  });

  it('carrega o evento e preenche o formulário', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();

    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    expect(view.getByText('123456')).toBeInTheDocument();
    expect(view.getByText('Em preparação')).toBeInTheDocument();
  });

  it('mostra a quantidade de pessoas que responderam', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount([], 3);

    await waitFor(() => {
      // O contador da aba "Respostas" também mostra 3; aqui interessa o card.
      expect(view.getByText('3', { selector: '.stat-num' })).toBeInTheDocument();
      expect(view.getByText('responderam')).toBeInTheDocument();
    });
  });

  it('exige título ao salvar', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.input(view.getByLabelText('Título'), { target: { value: '' } });
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    expect(await view.findByText('Informe o título do evento.')).toBeInTheDocument();
    expect(api.eventos.atualizar).not.toHaveBeenCalled();
  });

  it('salva mantendo o PIN quando ele não é alterado', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.atualizar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.input(view.getByLabelText('Título'), { target: { value: 'Título novo' } });
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    await waitFor(() =>
      expect(api.eventos.atualizar).toHaveBeenCalledWith('42', {
        title: 'Título novo',
        pinCode: '123456'
      })
    );
    await waitFor(() => expect(get(toast)?.message).toBe('Alterações salvas!'));
  });

  it('salva com novo PIN ao editar o campo diretamente', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.atualizar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    const pinInput = view.getByLabelText('PIN');
    await userEvent.clear(pinInput);
    await userEvent.type(pinInput, 'dev-team');
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    await waitFor(() =>
      expect(api.eventos.atualizar).toHaveBeenCalledWith('42', {
        title: 'Conecta DevOps',
        pinCode: 'dev-team'
      })
    );
    expect(view.getByText('DEV-TEAM')).toBeInTheDocument();
  });

  it('rejeita PIN inválido', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    const pinInput = view.getByLabelText('PIN');
    await userEvent.clear(pinInput);
    await userEvent.type(pinInput, 'com espaço');
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    expect(await view.findByText(/O PIN deve ter 1 a 25 caracteres/)).toBeInTheDocument();
    expect(api.eventos.atualizar).not.toHaveBeenCalled();
  });

  it('mostra erro do servidor', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.atualizar.mockRejectedValue(new Error('Este PIN já está em uso. Escolha outro.'));
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.input(view.getByLabelText('Título'), { target: { value: 'Título novo' } });
    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    expect(
      await view.findByText('Este PIN já está em uso. Escolha outro.')
    ).toBeInTheDocument();
  });

  it('não mostra "Salvar alterações" sem mudar título/PIN, e some de novo após salvar', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.atualizar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    expect(view.queryByRole('button', { name: 'Salvar alterações' })).not.toBeInTheDocument();

    await fireEvent.input(view.getByLabelText('Título'), { target: { value: 'Título novo' } });
    expect(view.getByRole('button', { name: 'Salvar alterações' })).toBeInTheDocument();

    await fireEvent.click(view.getByRole('button', { name: 'Salvar alterações' }));

    await waitFor(() =>
      expect(view.queryByRole('button', { name: 'Salvar alterações' })).not.toBeInTheDocument()
    );
  });

  it('salva "Permitir editar depois" sozinho, sem precisar do botão de salvar', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.atualizar.mockResolvedValue({ event: { ...event, allowEdit: true } });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('switch', { name: 'Permitir editar depois' }));

    await waitFor(() =>
      expect(api.eventos.atualizar).toHaveBeenCalledWith('42', {
        title: 'Conecta DevOps',
        pinCode: '',
        allowEdit: true
      })
    );
    await waitFor(() => expect(get(toast)?.message).toBe('Salvo'));
    expect(view.queryByRole('button', { name: 'Salvar alterações' })).not.toBeInTheDocument();
  });

  it('exclui o evento após confirmação e volta pro painel', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.deletar.mockResolvedValue({ ok: true });
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: 'Excluir evento' }));

    expect(confirmSpy).toHaveBeenCalledWith(expect.stringContaining('Conecta DevOps'));
    await waitFor(() => expect(api.eventos.deletar).toHaveBeenCalledWith('42'));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/painel'));
    confirmSpy.mockRestore();
  });

  it('não exclui o evento quando o usuário cancela a confirmação', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: 'Excluir evento' }));

    expect(confirmSpy).toHaveBeenCalled();
    expect(api.eventos.deletar).not.toHaveBeenCalled();
    expect(navigate).not.toHaveBeenCalled();
    confirmSpy.mockRestore();
  });

  it('mostra tela de não encontrado quando o evento não existe', async () => {
    const err = new Error('Evento não encontrado.');
    err.status = 404;
    api.eventos.buscar.mockRejectedValue(err);
    const view = mount();

    expect(await view.findByText('Evento não encontrado')).toBeInTheDocument();
  });

  it('carrega e exibe as perguntas do evento', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const questions = [
      {
        id: '1',
        eventId: '42',
        title: 'Qual é o seu prato favorito?',
        type: 'GROUP',
        orderIndex: 0,
        options: [
          { id: '10', text: 'Pizza' },
          { id: '11', text: 'Sushi' }
        ]
      },
      {
        id: '2',
        eventId: '42',
        title: 'Frase motivacional',
        type: 'OPEN_TEXT',
        orderIndex: 1,
        options: []
      }
    ];
    const view = mount(questions);

    expect(await view.findByText('Qual é o seu prato favorito?')).toBeInTheDocument();
    expect(view.getByText('Pizza')).toBeInTheDocument();
    expect(view.getByText('Sushi')).toBeInTheDocument();
    expect(view.getByText('Frase motivacional')).toBeInTheDocument();
    expect(view.getByText('Resposta aberta')).toBeInTheDocument();
  });

  it('alterna entre as abas Perguntas e Respostas', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    api.eventos.respostas.listar.mockResolvedValue({
      participantCount: 1,
      participants: [
        {
          id: 'p1',
          email: 'ana@exemplo.com',
          photo: '',
          createdAt: '2026-08-03T00:00:00Z',
          answers: []
        }
      ]
    });

    // painel de configurações (barra lateral) deve ficar sempre visível
    expect(view.getByLabelText('Título')).toBeInTheDocument();
    expect(view.getByText('Nenhuma pergunta ainda.')).toBeInTheDocument();
    expect(view.queryByText('ana@exemplo.com')).not.toBeInTheDocument();

    await fireEvent.click(view.getByRole('tab', { name: 'Respostas' }));

    expect((await view.findAllByText('ana@exemplo.com')).length).toBeGreaterThan(0);
    expect(view.getByLabelText('Título')).toBeInTheDocument();
    expect(view.queryByText('Nenhuma pergunta ainda.')).not.toBeInTheDocument();

    await fireEvent.click(view.getByRole('tab', { name: 'Perguntas' }));

    expect(view.getByLabelText('Título')).toBeInTheDocument();
    expect(view.getByText('Nenhuma pergunta ainda.')).toBeInTheDocument();
    expect(view.queryByText('ana@exemplo.com')).not.toBeInTheDocument();
  });

  function questionWithOther() {
    return {
      id: 'q1',
      eventId: '42',
      title: 'Qual sua linguagem favorita?',
      type: 'GROUP',
      orderIndex: 0,
      options: [
        { id: 'o-go', text: 'Go' },
        { id: 'o-js', text: 'JS' },
        { id: 'o-other', text: 'Outro', isOther: true }
      ]
    };
  }

  it('mostra "Escolha única" (não "Individual") como o tipo da pergunta nas respostas', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const q = questionWithOther();
    const view = mount([q]);
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    api.eventos.respostas.listar.mockResolvedValue({
      participantCount: 1,
      participants: [
        {
          id: 'p1',
          email: 'ana@exemplo.com',
          photo: '',
          createdAt: '2026-08-03T00:00:00Z',
          answers: [
            { questionId: 'q1', questionTitle: q.title, questionType: 'GROUP', optionId: 'o-go', optionText: 'Go', text: '' }
          ]
        }
      ]
    });

    await fireEvent.click(view.getByRole('tab', { name: 'Respostas' }));
    await view.findByLabelText('Copiar link de edição');

    expect(await view.findByText('Escolha única')).toBeInTheDocument();
    expect(view.queryByText('Individual')).not.toBeInTheDocument();
  });

  it('organizador escolhendo "Outro" precisa descrever a resposta antes de salvar', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const q = questionWithOther();
    const view = mount([q]);
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    api.eventos.respostas.listar.mockResolvedValue({
      participantCount: 1,
      participants: [
        {
          id: 'p1',
          email: 'ana@exemplo.com',
          photo: '',
          createdAt: '2026-08-03T00:00:00Z',
          answers: [
            { questionId: 'q1', questionTitle: q.title, questionType: 'GROUP', optionId: 'o-go', optionText: 'Go', text: '' }
          ]
        }
      ]
    });

    await fireEvent.click(view.getByRole('tab', { name: 'Respostas' }));
    await view.findByLabelText('Copiar link de edição');
    await view.findByRole('button', { name: 'Go' });

    await fireEvent.click(view.getByRole('button', { name: 'Outro' }));
    const input = await view.findByPlaceholderText('Descreva a resposta em "Outro"');

    // tenta salvar vazio -> erro, sem chamar a API
    await fireEvent.change(input, { target: { value: '   ' } });
    await fireEvent.blur(input);
    await waitFor(() => expect(get(toast)?.message).toBe('Descreva a resposta em "Outro".'));
    expect(api.eventos.respostas.atualizarResposta).not.toHaveBeenCalled();

    api.eventos.respostas.atualizarResposta.mockResolvedValue({ ok: true });
    const updated = {
      id: 'p1',
      email: 'ana@exemplo.com',
      photo: '',
      createdAt: '2026-08-03T00:00:00Z',
      answers: [
        { questionId: 'q1', questionTitle: q.title, questionType: 'GROUP', optionId: 'o-other', optionText: 'Outro', text: 'Rust' }
      ]
    };
    api.eventos.respostas.listar.mockResolvedValue({ participantCount: 1, participants: [updated] });

    await fireEvent.change(input, { target: { value: 'Rust' } });
    await fireEvent.blur(input);

    await waitFor(() =>
      expect(api.eventos.respostas.atualizarResposta).toHaveBeenCalledWith('42', 'p1', 'q1', {
        optionId: 'o-other',
        text: 'Rust'
      })
    );
    expect(await view.findByDisplayValue('Rust')).toBeInTheDocument();
  });

  it('resposta já salva em "Outro" mostra o texto pré-preenchido sem precisar clicar', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const q = questionWithOther();
    const view = mount([q]);
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    api.eventos.respostas.listar.mockResolvedValue({
      participantCount: 1,
      participants: [
        {
          id: 'p1',
          email: 'ana@exemplo.com',
          photo: '',
          createdAt: '2026-08-03T00:00:00Z',
          answers: [
            { questionId: 'q1', questionTitle: q.title, questionType: 'GROUP', optionId: 'o-other', optionText: 'Outro', text: 'Kotlin' }
          ]
        }
      ]
    });

    await fireEvent.click(view.getByRole('tab', { name: 'Respostas' }));
    await view.findByLabelText('Copiar link de edição');

    expect(await view.findByDisplayValue('Kotlin')).toBeInTheDocument();
    expect(view.getByRole('button', { name: 'Outro' })).toHaveClass('selected');
  });

  it('adiciona uma pergunta em grupo', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();

    const created = {
      id: '9',
      eventId: '42',
      title: 'Qual é o seu café preferido?',
      type: 'GROUP',
      orderIndex: 0,
      options: [
        { id: '1', text: 'Espresso' },
        { id: '2', text: 'Cappuccino' }
      ]
    };
    api.eventos.perguntas.criar.mockResolvedValue({ question: created });

    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar pergunta' }));
    await userEvent.type(view.getByLabelText('Pergunta'), 'Qual é o seu café preferido?');
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Espresso');
    await userEvent.type(view.getByPlaceholderText('Opção 2'), 'Cappuccino');
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar' }));

    await waitFor(() =>
      expect(api.eventos.perguntas.criar).toHaveBeenCalledWith('42', {
        title: 'Qual é o seu café preferido?',
        type: 'GROUP',
        options: ['Espresso', 'Cappuccino'],
        allowOther: false
      })
    );
    expect(await view.findByText('Qual é o seu café preferido?')).toBeInTheDocument();
    await waitFor(() => expect(get(toast)?.message).toBe('Pergunta adicionada!'));
  });

  it('adiciona uma pergunta em grupo com a opção "Outro"', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();

    const created = {
      id: '9',
      eventId: '42',
      title: 'Qual é o seu café preferido?',
      type: 'GROUP',
      orderIndex: 0,
      options: [
        { id: '1', text: 'Espresso' },
        { id: '2', text: 'Cappuccino' },
        { id: '3', text: 'Outro', isOther: true }
      ]
    };
    api.eventos.perguntas.criar.mockResolvedValue({ question: created });

    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar pergunta' }));
    await userEvent.type(view.getByLabelText('Pergunta'), 'Qual é o seu café preferido?');
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Espresso');
    await userEvent.type(view.getByPlaceholderText('Opção 2'), 'Cappuccino');
    await fireEvent.click(view.getByRole('switch', { name: 'Adicionar opção Outro com resposta livre' }));
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar' }));

    await waitFor(() =>
      expect(api.eventos.perguntas.criar).toHaveBeenCalledWith('42', {
        title: 'Qual é o seu café preferido?',
        type: 'GROUP',
        options: ['Espresso', 'Cappuccino'],
        allowOther: true
      })
    );
    expect(await view.findByText('Qual é o seu café preferido?')).toBeInTheDocument();
  });

  it('cria uma pergunta de resposta aberta sem exigir opções', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();

    const created = {
      id: '11',
      eventId: '42',
      title: 'Qual sua comida favorita?',
      type: 'OPEN_TEXT',
      orderIndex: 0,
      options: []
    };
    api.eventos.perguntas.criar.mockResolvedValue({ question: created });

    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar pergunta' }));
    await fireEvent.click(view.getByLabelText('Resposta aberta'));
    await userEvent.type(view.getByLabelText('Pergunta'), 'Qual sua comida favorita?');

    expect(view.queryByPlaceholderText('Opção 1')).not.toBeInTheDocument();

    await fireEvent.click(view.getByRole('button', { name: 'Adicionar' }));

    await waitFor(() =>
      expect(api.eventos.perguntas.criar).toHaveBeenCalledWith('42', {
        title: 'Qual sua comida favorita?',
        type: 'OPEN_TEXT',
        options: [],
        allowOther: false
      })
    );
    expect(await view.findByText('Qual sua comida favorita?')).toBeInTheDocument();
    expect(view.getByText('Resposta aberta')).toBeInTheDocument();
  });

  it('insere uma pergunta entre duas existentes', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const questions = [question('1', 'Primeira'), question('2', 'Segunda')];
    const view = mount(questions);
    await view.findByText('Primeira');

    const created = {
      id: '9',
      eventId: '42',
      title: 'Nova pergunta',
      type: 'GROUP',
      orderIndex: 2,
      options: [
        { id: '90', text: 'Sim' },
        { id: '91', text: 'Não' }
      ]
    };
    api.eventos.perguntas.criar.mockResolvedValue({ question: created });
    api.eventos.perguntas.reordenar.mockResolvedValue(null);

    await fireEvent.click(view.getByLabelText('Adicionar pergunta após "Primeira"'));
    await userEvent.type(view.getByLabelText('Pergunta'), 'Nova pergunta');
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Sim');
    await userEvent.type(view.getByPlaceholderText('Opção 2'), 'Não');
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar' }));

    await waitFor(() =>
      expect(api.eventos.perguntas.criar).toHaveBeenCalledWith('42', {
        title: 'Nova pergunta',
        type: 'GROUP',
        options: ['Sim', 'Não'],
        allowOther: false
      })
    );
    await waitFor(() =>
      expect(api.eventos.perguntas.reordenar).toHaveBeenCalledWith('42', ['1', '9', '2'])
    );
    const titles = view
      .getAllByText(/Primeira|Segunda|Nova pergunta/)
      .map((el) => el.textContent);
    expect(titles).toEqual(['Primeira', 'Nova pergunta', 'Segunda']);
  });

  it('exige texto na pergunta', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar pergunta' }));
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar' }));

    expect(await view.findByText('Informe o texto da pergunta.')).toBeInTheDocument();
    expect(api.eventos.perguntas.criar).not.toHaveBeenCalled();
  });

  it('exige pelo menos 2 opções em perguntas de grupo', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar pergunta' }));
    await userEvent.type(view.getByLabelText('Pergunta'), 'Qual sua linguagem?');
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Go');
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar' }));

    expect(await view.findByText('Informe pelo menos 2 opções.')).toBeInTheDocument();
    expect(api.eventos.perguntas.criar).not.toHaveBeenCalled();
  });

  it('permite adicionar e remover campos de opção', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar pergunta' }));
    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar opção' }));
    expect(await view.findByPlaceholderText('Opção 3')).toBeInTheDocument();

    await fireEvent.click(view.getByLabelText('Remover opção 3'));
    await waitFor(() => expect(view.queryByPlaceholderText('Opção 3')).not.toBeInTheDocument());
  });

  it('reordena opções com as setas sem precisar apagar e adicionar de novo', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar pergunta' }));
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Go');
    await userEvent.type(view.getByPlaceholderText('Opção 2'), 'Python');

    expect(view.getByLabelText('Mover opção 1 para cima')).toBeDisabled();
    expect(view.getByLabelText('Mover opção 2 para baixo')).toBeDisabled();

    await fireEvent.click(view.getByLabelText('Mover opção 1 para baixo'));

    expect(view.getByPlaceholderText('Opção 1').value).toBe('Python');
    expect(view.getByPlaceholderText('Opção 2').value).toBe('Go');
  });

  it('move pergunta para baixo com a seta e persiste a ordem', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.perguntas.reordenar.mockResolvedValue(null);
    const questions = [
      question('1', 'Primeira'),
      question('2', 'Segunda'),
      question('3', 'Terceira')
    ];
    const view = mount(questions);
    await view.findByText('Primeira');

    await fireEvent.click(view.getAllByLabelText('Mover pergunta para baixo')[0]);

    await waitFor(() =>
      expect(api.eventos.perguntas.reordenar).toHaveBeenCalledWith('42', ['2', '1', '3'])
    );
    const titles = view.getAllByText(/Primeira|Segunda|Terceira/).map((el) => el.textContent);
    expect(titles).toEqual(['Segunda', 'Primeira', 'Terceira']);
  });

  it('move pergunta para cima com a seta', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.perguntas.reordenar.mockResolvedValue(null);
    const questions = [question('1', 'Primeira'), question('2', 'Segunda')];
    const view = mount(questions);
    await view.findByText('Primeira');

    await fireEvent.click(view.getAllByLabelText('Mover pergunta para cima')[1]);

    await waitFor(() =>
      expect(api.eventos.perguntas.reordenar).toHaveBeenCalledWith('42', ['2', '1'])
    );
  });

  it('desabilita setas nos extremos da lista', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const questions = [question('1', 'Primeira'), question('2', 'Segunda')];
    const view = mount(questions);
    await view.findByText('Primeira');

    expect(view.getAllByLabelText('Mover pergunta para cima')[0]).toBeDisabled();
    expect(view.getAllByLabelText('Mover pergunta para baixo')[1]).toBeDisabled();
  });

  // jsdom não faz layout de verdade: getBoundingClientRect() sempre retorna
  // zeros. Simula posições realistas para os itens da lista de perguntas,
  // empilhados verticalmente com a altura informada.
  function mockQuestionSlotRects(list, heights) {
    const slots = list.querySelectorAll(':scope > li.question-slot');
    let top = 0;
    slots.forEach((el, i) => {
      const height = heights[i] ?? 60;
      const itemTop = top; // valor próprio por item; evita closure compartilhada no loop
      el.getBoundingClientRect = () => ({
        top: itemTop,
        bottom: itemTop + height,
        height,
        left: 0,
        right: 300,
        width: 300
      });
      top += height;
    });
    return slots;
  }

  // fireEvent.dragOver/drop (@testing-library/dom) não repassam clientY para
  // o evento disparado neste ambiente jsdom; MouseEvent nativo repassa.
  function dispatchDragEventAt(el, type, clientY, dataTransfer) {
    const ev = new MouseEvent(type, { bubbles: true, cancelable: true, clientY });
    ev.dataTransfer = dataTransfer;
    el.dispatchEvent(ev);
  }

  it('troca a pergunta adjacente ao cruzar o meio dela ao arrastar para baixo', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.perguntas.reordenar.mockResolvedValue(null);
    const questions = [question('1', 'Primeira'), question('2', 'Segunda'), question('3', 'Terceira')];
    const view = mount(questions);
    await view.findByText('Primeira');

    const first = view.getAllByLabelText('Mover pergunta para cima')[0].closest('.question-item');
    const list = first.closest('ul');
    mockQuestionSlotRects(list, [60, 60, 60]); // Primeira: 0-60, Segunda: 60-120, Terceira: 120-180
    const dt = { setData: vi.fn(), effectAllowed: '', dropEffect: '' };

    await fireEvent.dragStart(first, { dataTransfer: dt });
    // cruza o meio de "Segunda" (60 + 30 = 90) -> deve trocar imediatamente,
    // sem precisar arrastar até "Terceira"
    await dispatchDragEventAt(list, 'dragover', 95, dt);
    await dispatchDragEventAt(list, 'drop', 95, dt);

    await waitFor(() =>
      expect(api.eventos.perguntas.reordenar).toHaveBeenCalledWith('42', ['2', '1', '3'])
    );
    expect(dt.setData).toHaveBeenCalled();
  });

  it('move o item além do vizinho imediato quando o arrasto cruza mais de um meio', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.perguntas.reordenar.mockResolvedValue(null);
    const questions = [question('1', 'Primeira'), question('2', 'Segunda'), question('3', 'Terceira')];
    const view = mount(questions);
    await view.findByText('Primeira');

    const first = view.getAllByLabelText('Mover pergunta para cima')[0].closest('.question-item');
    const list = first.closest('ul');
    mockQuestionSlotRects(list, [60, 60, 60]);
    const dt = { setData: vi.fn(), effectAllowed: '', dropEffect: '' };

    await fireEvent.dragStart(first, { dataTransfer: dt });
    // cruza o meio de "Segunda" (90) e de "Terceira" (150)
    await dispatchDragEventAt(list, 'dragover', 160, dt);
    await dispatchDragEventAt(list, 'drop', 160, dt);

    await waitFor(() =>
      expect(api.eventos.perguntas.reordenar).toHaveBeenCalledWith('42', ['2', '3', '1'])
    );
  });

  it('troca a pergunta adjacente ao cruzar o meio dela ao arrastar para cima', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.perguntas.reordenar.mockResolvedValue(null);
    const questions = [question('1', 'Primeira'), question('2', 'Segunda'), question('3', 'Terceira')];
    const view = mount(questions);
    await view.findByText('Primeira');

    const third = view.getAllByLabelText('Mover pergunta para cima')[2].closest('.question-item');
    const list = third.closest('ul');
    mockQuestionSlotRects(list, [60, 60, 60]);
    const dt = { setData: vi.fn(), effectAllowed: '', dropEffect: '' };

    await fireEvent.dragStart(third, { dataTransfer: dt });
    // "Terceira" (índice 2) cruza o meio de "Segunda" (60 + 30 = 90) arrastando para cima
    await dispatchDragEventAt(list, 'dragover', 85, dt);
    await dispatchDragEventAt(list, 'drop', 85, dt);

    await waitFor(() =>
      expect(api.eventos.perguntas.reordenar).toHaveBeenCalledWith('42', ['1', '3', '2'])
    );
  });

  it('recarrega as perguntas quando a reordenação falha', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.perguntas.reordenar.mockRejectedValueOnce(new Error('Falha ao reordenar.'));
    api.eventos.perguntas.listar.mockResolvedValue({
      questions: [question('1', 'Primeira'), question('2', 'Segunda')]
    });
    const questions = [question('2', 'Segunda'), question('1', 'Primeira')];
    const view = mount(questions);
    await view.findByText('Segunda');

    await fireEvent.click(view.getAllByLabelText('Mover pergunta para cima')[1]);

    await waitFor(() =>
      expect(view.getAllByText(/Primeira|Segunda/)[0].textContent).toBe('Primeira')
    );
    expect(get(toast)?.message).toBe('Falha ao reordenar.');
  });

  it('remove uma pergunta existente', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const questions = [
      {
        id: '5',
        eventId: '42',
        title: 'Pergunta a remover',
        type: 'GROUP',
        orderIndex: 0,
        options: [{ id: '1', text: 'A' }]
      }
    ];
    api.eventos.perguntas.remover.mockResolvedValue(null);
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    const view = mount(questions);
    await view.findByText('Pergunta a remover');

    await fireEvent.click(view.getByLabelText('Remover pergunta Pergunta a remover'));

    expect(confirmSpy).toHaveBeenCalled();
    await waitFor(() => expect(api.eventos.perguntas.remover).toHaveBeenCalledWith('42', '5'));
    await waitFor(() => expect(view.queryByText('Pergunta a remover')).not.toBeInTheDocument());
    await waitFor(() => expect(get(toast)?.message).toBe('Pergunta removida.'));
  });

  it('nao remove a pergunta se a confirmacao for cancelada', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const questions = [question('5', 'Pergunta a manter')];
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    const view = mount(questions);
    await view.findByText('Pergunta a manter');

    await fireEvent.click(view.getByLabelText('Remover pergunta Pergunta a manter'));

    expect(api.eventos.perguntas.remover).not.toHaveBeenCalled();
    expect(view.getByText('Pergunta a manter')).toBeInTheDocument();
  });

  it('edita uma pergunta existente', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const questions = [question('1', 'Primeira')];
    const updated = question('1', 'Primeira editada');
    updated.options = [
      { id: '50', text: 'X' },
      { id: '51', text: 'B' }
    ];
    api.eventos.perguntas.atualizar.mockResolvedValue({ question: updated });
    const view = mount(questions);
    await view.findByText('Primeira');

    await fireEvent.click(view.getByLabelText('Editar pergunta Primeira'));

    const titleInput = view.getByLabelText('Pergunta');
    expect(titleInput.value).toBe('Primeira');
    expect(view.getByPlaceholderText('Opção 1').value).toBe('A');
    expect(view.getByPlaceholderText('Opção 2').value).toBe('B');
    expect(view.getByText('Editar pergunta')).toBeInTheDocument();

    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, 'Primeira editada');
    await userEvent.clear(view.getByPlaceholderText('Opção 1'));
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'X');
    await fireEvent.click(view.getByRole('button', { name: 'Salvar pergunta' }));

    await waitFor(() =>
      expect(api.eventos.perguntas.atualizar).toHaveBeenCalledWith('42', '1', {
        title: 'Primeira editada',
        options: ['X', 'B'],
        allowOther: false
      })
    );
    expect(await view.findByText('Primeira editada')).toBeInTheDocument();
    await waitFor(() => expect(get(toast)?.message).toBe('Pergunta atualizada!'));
    await waitFor(() =>
      expect(view.getByLabelText('Editar pergunta Primeira editada')).toBeInTheDocument()
    );
  });

  it('edita pergunta com "Outro": a opção sintética some da lista e o switch vem ligado', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const q = question('1', 'Primeira');
    q.options = [
      { id: '10', text: 'A' },
      { id: '11', text: 'B' },
      { id: '12', text: 'Outro', isOther: true }
    ];
    const view = mount([q]);
    await view.findByText('Primeira');

    await fireEvent.click(view.getByLabelText('Editar pergunta Primeira'));

    expect(view.getByPlaceholderText('Opção 1').value).toBe('A');
    expect(view.getByPlaceholderText('Opção 2').value).toBe('B');
    expect(view.queryByPlaceholderText('Opção 3')).not.toBeInTheDocument();
    expect(
      view.getByRole('switch', { name: 'Adicionar opção Outro com resposta livre' })
    ).toHaveAttribute('aria-checked', 'true');
  });

  it('cancela a edição de uma pergunta', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    const view = mount([question('1', 'Primeira')]);
    await view.findByText('Primeira');

    await fireEvent.click(view.getByLabelText('Editar pergunta Primeira'));
    expect(view.getByText('Editar pergunta')).toBeInTheDocument();
    await fireEvent.click(view.getByRole('button', { name: 'Cancelar' }));

    expect(view.getByLabelText('Editar pergunta Primeira')).toBeInTheDocument();
    expect(api.eventos.perguntas.atualizar).not.toHaveBeenCalled();
  });

  it('mostra erro do servidor ao adicionar pergunta', async () => {
    api.eventos.buscar.mockResolvedValue({ event });
    api.eventos.perguntas.criar.mockRejectedValue(new Error('As opções devem ser diferentes entre si.'));
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar pergunta' }));
    await userEvent.type(view.getByLabelText('Pergunta'), 'Duplicada');
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Go');
    await userEvent.type(view.getByPlaceholderText('Opção 2'), 'Python');
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar' }));

    expect(
      await view.findByText('As opções devem ser diferentes entre si.')
    ).toBeInTheDocument();
  });
});
