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
    events: {
      get: vi.fn(),
      update: vi.fn(),
      questions: {
        list: vi.fn(),
        create: vi.fn(),
        update: vi.fn(),
        reorder: vi.fn(),
        remove: vi.fn()
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
  configShowRanking: true,
  createdAt: '2026-08-02T00:00:00Z'
};

function question(id, title) {
  return {
    id,
    eventId: '42',
    title,
    type: 'GROUP',
    layoutView: 'TIMELINE',
    orderIndex: 0,
    options: [
      { id: `${id}-1`, text: 'A' },
      { id: `${id}-2`, text: 'B' }
    ]
  };
}

function mount(questions = []) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  api.events.questions.list.mockResolvedValue({ questions });
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

  it('carrega e exibe as perguntas do evento', async () => {
    api.events.get.mockResolvedValue({ event });
    const questions = [
      {
        id: '1',
        eventId: '42',
        title: 'Qual é o seu prato favorito?',
        type: 'GROUP',
        layoutView: 'TIMELINE',
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
        type: 'INDIVIDUAL',
        layoutView: 'TIMELINE',
        orderIndex: 1,
        options: [{ id: '12', text: 'Rascunho' }]
      }
    ];
    const view = mount(questions);

    expect(await view.findByText('Qual é o seu prato favorito?')).toBeInTheDocument();
    expect(view.getByText('Pizza')).toBeInTheDocument();
    expect(view.getByText('Sushi')).toBeInTheDocument();
    expect(view.getByText('Frase motivacional')).toBeInTheDocument();
    expect(view.getByText('Individual')).toBeInTheDocument();
  });

  it('adiciona uma pergunta em grupo', async () => {
    api.events.get.mockResolvedValue({ event });
    const view = mount();

    const created = {
      id: '9',
      eventId: '42',
      title: 'Qual é o seu café preferido?',
      type: 'GROUP',
      layoutView: 'TIMELINE',
      orderIndex: 0,
      options: [
        { id: '1', text: 'Espresso' },
        { id: '2', text: 'Cappuccino' }
      ]
    };
    api.events.questions.create.mockResolvedValue({ question: created });

    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));
    await userEvent.type(view.getByLabelText('Pergunta'), 'Qual é o seu café preferido?');
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Espresso');
    await userEvent.type(view.getByPlaceholderText('Opção 2'), 'Cappuccino');
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar pergunta' }));

    await waitFor(() =>
      expect(api.events.questions.create).toHaveBeenCalledWith('42', {
        title: 'Qual é o seu café preferido?',
        type: 'GROUP',
        options: ['Espresso', 'Cappuccino']
      })
    );
    expect(await view.findByText('Qual é o seu café preferido?')).toBeInTheDocument();
    await waitFor(() => expect(get(toast)?.message).toBe('Pergunta adicionada!'));
  });

  it('exige texto na pergunta', async () => {
    api.events.get.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: 'Adicionar pergunta' }));

    expect(await view.findByText('Informe o texto da pergunta.')).toBeInTheDocument();
    expect(api.events.questions.create).not.toHaveBeenCalled();
  });

  it('exige pelo menos 2 opções em perguntas de grupo', async () => {
    api.events.get.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await userEvent.type(view.getByLabelText('Pergunta'), 'Qual sua linguagem?');
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Go');
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar pergunta' }));

    expect(await view.findByText('Informe pelo menos 2 opções.')).toBeInTheDocument();
    expect(api.events.questions.create).not.toHaveBeenCalled();
  });

  it('permite adicionar e remover campos de opção', async () => {
    api.events.get.mockResolvedValue({ event });
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await fireEvent.click(view.getByRole('button', { name: '+ Adicionar opção' }));
    expect(await view.findByPlaceholderText('Opção 3')).toBeInTheDocument();

    await fireEvent.click(view.getByLabelText('Remover opção 3'));
    await waitFor(() => expect(view.queryByPlaceholderText('Opção 3')).not.toBeInTheDocument());
  });

  it('move pergunta para baixo com a seta e persiste a ordem', async () => {
    api.events.get.mockResolvedValue({ event });
    api.events.questions.reorder.mockResolvedValue(null);
    const questions = [
      question('1', 'Primeira'),
      question('2', 'Segunda'),
      question('3', 'Terceira')
    ];
    const view = mount(questions);
    await view.findByText('Primeira');

    await fireEvent.click(view.getAllByLabelText('Mover pergunta para baixo')[0]);

    await waitFor(() =>
      expect(api.events.questions.reorder).toHaveBeenCalledWith('42', ['2', '1', '3'])
    );
    const titles = view.getAllByText(/Primeira|Segunda|Terceira/).map((el) => el.textContent);
    expect(titles).toEqual(['Segunda', 'Primeira', 'Terceira']);
  });

  it('move pergunta para cima com a seta', async () => {
    api.events.get.mockResolvedValue({ event });
    api.events.questions.reorder.mockResolvedValue(null);
    const questions = [question('1', 'Primeira'), question('2', 'Segunda')];
    const view = mount(questions);
    await view.findByText('Primeira');

    await fireEvent.click(view.getAllByLabelText('Mover pergunta para cima')[1]);

    await waitFor(() =>
      expect(api.events.questions.reorder).toHaveBeenCalledWith('42', ['2', '1'])
    );
  });

  it('desabilita setas nos extremos da lista', async () => {
    api.events.get.mockResolvedValue({ event });
    const questions = [question('1', 'Primeira'), question('2', 'Segunda')];
    const view = mount(questions);
    await view.findByText('Primeira');

    expect(view.getAllByLabelText('Mover pergunta para cima')[0]).toBeDisabled();
    expect(view.getAllByLabelText('Mover pergunta para baixo')[1]).toBeDisabled();
  });

  it('reordena por arrastar e soltar', async () => {
    api.events.get.mockResolvedValue({ event });
    api.events.questions.reorder.mockResolvedValue(null);
    const questions = [question('1', 'Primeira'), question('2', 'Segunda'), question('3', 'Terceira')];
    const view = mount(questions);
    await view.findByText('Primeira');

    const first = view.getAllByLabelText('Mover pergunta para cima')[0].closest('li');
    const list = first.parentElement;
    const dt = { setData: vi.fn(), effectAllowed: '', dropEffect: '' };

    await fireEvent.dragStart(first, { dataTransfer: dt });
    await fireEvent.dragOver(list, { dataTransfer: dt });
    await fireEvent.drop(list, { dataTransfer: dt });

    // jsdom reporta retângulos zerados: o alvo vira o fim da lista
    await waitFor(() =>
      expect(api.events.questions.reorder).toHaveBeenCalledWith('42', ['2', '3', '1'])
    );
    expect(dt.setData).toHaveBeenCalled();
  });

  it('recarrega as perguntas quando a reordenação falha', async () => {
    api.events.get.mockResolvedValue({ event });
    api.events.questions.reorder.mockRejectedValueOnce(new Error('Falha ao reordenar.'));
    api.events.questions.list.mockResolvedValue({
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
    api.events.get.mockResolvedValue({ event });
    const questions = [
      {
        id: '5',
        eventId: '42',
        title: 'Pergunta a remover',
        type: 'GROUP',
        layoutView: 'TIMELINE',
        orderIndex: 0,
        options: [{ id: '1', text: 'A' }]
      }
    ];
    api.events.questions.remove.mockResolvedValue(null);
    const view = mount(questions);
    await view.findByText('Pergunta a remover');

    await fireEvent.click(view.getByLabelText('Remover pergunta Pergunta a remover'));

    await waitFor(() => expect(api.events.questions.remove).toHaveBeenCalledWith('42', '5'));
    await waitFor(() => expect(view.queryByText('Pergunta a remover')).not.toBeInTheDocument());
    await waitFor(() => expect(get(toast)?.message).toBe('Pergunta removida.'));
  });

  it('edita uma pergunta existente', async () => {
    api.events.get.mockResolvedValue({ event });
    const questions = [question('1', 'Primeira')];
    const updated = question('1', 'Primeira editada');
    updated.options = [
      { id: '50', text: 'X' },
      { id: '51', text: 'B' }
    ];
    api.events.questions.update.mockResolvedValue({ question: updated });
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
      expect(api.events.questions.update).toHaveBeenCalledWith('42', '1', {
        title: 'Primeira editada',
        options: ['X', 'B']
      })
    );
    expect(await view.findByText('Primeira editada')).toBeInTheDocument();
    await waitFor(() => expect(get(toast)?.message).toBe('Pergunta atualizada!'));
    await waitFor(() =>
      expect(view.getByRole('button', { name: 'Adicionar pergunta' })).toBeInTheDocument()
    );
  });

  it('cancela a edição de uma pergunta', async () => {
    api.events.get.mockResolvedValue({ event });
    const view = mount([question('1', 'Primeira')]);
    await view.findByText('Primeira');

    await fireEvent.click(view.getByLabelText('Editar pergunta Primeira'));
    await fireEvent.click(view.getByRole('button', { name: 'Cancelar' }));

    expect(view.getByRole('button', { name: 'Adicionar pergunta' })).toBeInTheDocument();
    expect(view.getByLabelText('Pergunta').value).toBe('');
    expect(api.events.questions.update).not.toHaveBeenCalled();
  });

  it('mostra erro do servidor ao adicionar pergunta', async () => {
    api.events.get.mockResolvedValue({ event });
    api.events.questions.create.mockRejectedValue(new Error('As opções devem ser diferentes entre si.'));
    const view = mount();
    await waitFor(() => expect(view.getByLabelText('Título').value).toBe('Conecta DevOps'));

    await userEvent.type(view.getByLabelText('Pergunta'), 'Duplicada');
    await userEvent.type(view.getByPlaceholderText('Opção 1'), 'Go');
    await userEvent.type(view.getByPlaceholderText('Opção 2'), 'Python');
    await fireEvent.click(view.getByRole('button', { name: 'Adicionar pergunta' }));

    expect(
      await view.findByText('As opções devem ser diferentes entre si.')
    ).toBeInTheDocument();
  });
});
