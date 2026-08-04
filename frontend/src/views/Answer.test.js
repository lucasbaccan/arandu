import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import Answer from './Answer.svelte';

vi.mock('../lib/api.js', () => ({
  api: {
    public: {
      events: {
        get: vi.fn(),
        submit: vi.fn(),
        getParticipant: vi.fn()
      }
    }
  }
}));

import { api } from '../lib/api.js';

const groupQuestion = {
  id: 'q1',
  title: 'Qual sua linguagem favorita?',
  type: 'GROUP',
  options: [
    { id: 'opt-go', text: 'Go' },
    { id: 'opt-js', text: 'JS' }
  ]
};

const openQuestion = {
  id: 'q2',
  title: 'Qual sua comida favorita?',
  type: 'OPEN_TEXT',
  options: []
};

function mockLoad({ answersOpen = true, questions = [groupQuestion, openQuestion] } = {}) {
  api.public.events.get.mockResolvedValue({
    event: { id: '42', title: 'Dinâmica de Testes', answersOpen },
    questions
  });
}

describe('Tela de respostas do participante', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('mostra a tela de identificação com o título do evento', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });

    expect(await screen.findByText('Dinâmica de Testes')).toBeInTheDocument();
    expect(screen.getByLabelText('E-mail')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Começar' })).toBeInTheDocument();
  });

  it('exige e-mail válido antes de iniciar', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    expect(await screen.findByText('Informe seu e-mail.')).toBeInTheDocument();

    await userEvent.type(screen.getByLabelText('E-mail'), 'não-é-email');
    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    expect(await screen.findByText('Informe um e-mail válido.')).toBeInTheDocument();
  });

  it('rejeita na identificação um e-mail que o backend também rejeitaria', async () => {
    // Regra idêntica à do backend (isValidEmail em server.go): domínio não
    // aceita "_", só letras/dígitos/ponto/hífen. Antes da correção, esse
    // e-mail passava aqui e só falhava ao clicar em Finalizar, depois de
    // responder tudo.
    mockLoad();
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@my_domain.com');
    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));

    expect(await screen.findByText('Informe um e-mail válido.')).toBeInTheDocument();
    expect(screen.queryByText('Qual sua linguagem favorita?')).not.toBeInTheDocument();
  });

  it('mostra tela de respostas fechadas quando o evento não está aberto', async () => {
    mockLoad({ answersOpen: false });
    render(Answer, { props: { id: '42' } });

    expect(
      await screen.findByText('As respostas não estão abertas para este evento no momento.')
    ).toBeInTheDocument();
    expect(screen.queryByLabelText('E-mail')).not.toBeInTheDocument();
  });

  it('mostra evento não encontrado', async () => {
    const err = new Error('Evento não encontrado.');
    err.status = 404;
    api.public.events.get.mockRejectedValue(err);
    render(Answer, { props: { id: '999' } });

    expect(await screen.findByText('Evento não encontrado')).toBeInTheDocument();
  });

  it('valida cada pergunta, envia as respostas e mostra a tela de agradecimento', async () => {
    mockLoad();
    api.public.events.submit.mockResolvedValue({ ok: true });
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await userEvent.type(screen.getByLabelText('E-mail'), 'Participante@Exemplo.com');
    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    expect(
      await screen.findByText('A dinâmica não será a mesma sem sua foto.')
    ).toBeInTheDocument();
    await fireEvent.click(screen.getByRole('button', { name: 'Continuar mesmo assim' }));

    expect(await screen.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();

    // não deixa avançar sem selecionar uma opção
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    expect(await screen.findByText('Selecione uma opção para continuar.')).toBeInTheDocument();

    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));

    expect(await screen.findByText('Qual sua comida favorita?')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Finalizar' })).toBeInTheDocument();

    // texto só com espaços não deve ser aceito (trim)
    await userEvent.type(screen.getByPlaceholderText('Escreva sua resposta'), '   ');
    await fireEvent.click(screen.getByRole('button', { name: 'Finalizar' }));
    expect(await screen.findByText('Escreva uma resposta para continuar.')).toBeInTheDocument();
    expect(api.public.events.submit).not.toHaveBeenCalled();

    await userEvent.clear(screen.getByPlaceholderText('Escreva sua resposta'));
    await userEvent.type(screen.getByPlaceholderText('Escreva sua resposta'), '  Pizza  ');
    await fireEvent.click(screen.getByRole('button', { name: 'Finalizar' }));

    await waitFor(() =>
      expect(api.public.events.submit).toHaveBeenCalledWith('42', {
        email: 'Participante@Exemplo.com',
        photo: '',
        editToken: '',
        answers: [
          { questionId: 'q1', optionId: 'opt-go', text: '' },
          { questionId: 'q2', optionId: '', text: 'Pizza' }
        ]
      })
    );

    expect(await screen.findByText('Respostas enviadas!')).toBeInTheDocument();
  });

  it('mantém a resposta ao voltar para a pergunta anterior', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@exemplo.com');
    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Continuar mesmo assim' }));
    await screen.findByText('Qual sua linguagem favorita?');

    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    await screen.findByText('Qual sua comida favorita?');

    await fireEvent.click(screen.getByRole('button', { name: 'Voltar' }));
    await screen.findByText('Qual sua linguagem favorita?');

    expect(screen.getByLabelText('Go').checked).toBe(true);
  });

  it('remove do tab as perguntas fora da tela, evitando o bug de foco/scroll no carrossel', async () => {
    // Todas as perguntas ficam no DOM ao mesmo tempo (só a atual é deslocada
    // para a área visível via transform); sem tabindex="-1" nas ocultas, um
    // Tab movia o foco pra pergunta seguinte e o navegador rolava o
    // carrossel pra mostrá-la, sem que o estado (Pergunta X de Y) mudasse.
    mockLoad();
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@exemplo.com');
    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Continuar mesmo assim' }));
    await screen.findByText('Qual sua linguagem favorita?');

    expect(screen.getByLabelText('Go')).toHaveAttribute('tabindex', '0');
    expect(screen.getByLabelText('JS')).toHaveAttribute('tabindex', '0');
    expect(screen.getByPlaceholderText('Escreva sua resposta')).toHaveAttribute('tabindex', '-1');

    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    await screen.findByText('Qual sua comida favorita?');

    expect(screen.getByPlaceholderText('Escreva sua resposta')).toHaveAttribute('tabindex', '0');
    expect(screen.getByLabelText('Go')).toHaveAttribute('tabindex', '-1');
  });

  it('permite responder de novo (mesmo dispositivo), pois a unicidade é por e-mail no servidor', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });

    expect(await screen.findByText('Dinâmica de Testes')).toBeInTheDocument();
    expect(api.public.events.get).toHaveBeenCalledTimes(1);
  });

  it('volta para a identificação se o servidor ainda assim rejeitar o e-mail ao finalizar', async () => {
    mockLoad({ questions: [groupQuestion] });
    const err = new Error('Informe um e-mail válido.');
    err.status = 400;
    api.public.events.submit.mockRejectedValue(err);
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@exemplo.com');
    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Continuar mesmo assim' }));
    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Finalizar' }));

    expect(await screen.findByText('Informe um e-mail válido.')).toBeInTheDocument();
    expect(screen.getByLabelText('E-mail')).toBeInTheDocument();
    expect(screen.queryByText('Respostas enviadas!')).not.toBeInTheDocument();
  });

  it('mostra o link de edição na tela de agradecimento', async () => {
    mockLoad();
    api.public.events.submit.mockResolvedValue({ ok: true, editToken: 'tok123' });
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@exemplo.com');
    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Continuar mesmo assim' }));

    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    await userEvent.type(screen.getByPlaceholderText('Escreva sua resposta'), 'Pizza');
    await fireEvent.click(screen.getByRole('button', { name: 'Finalizar' }));

    expect(await screen.findByText('Respostas enviadas!')).toBeInTheDocument();
    const linkInput = screen.getByDisplayValue(/\/answer\/42\?edit=tok123$/);
    expect(linkInput).toBeInTheDocument();
    expect(
      screen.getByText(/solicitá-lo ao responsável pelo evento/)
    ).toBeInTheDocument();
  });

  it('pré-preenche as respostas ao acessar com um link de edição válido', async () => {
    window.history.pushState({}, '', '/answer/42?edit=tok123');
    mockLoad();
    api.public.events.getParticipant.mockResolvedValue({
      participant: {
        email: 'ana@exemplo.com',
        photo: '',
        answers: [
          { questionId: 'q1', optionId: 'opt-go', text: '' },
          { questionId: 'q2', optionId: '', text: 'Pizza' }
        ]
      }
    });
    render(Answer, { props: { id: '42' } });

    expect(await screen.findByText('Qual sua linguagem favorita?')).toBeInTheDocument();
    expect(api.public.events.getParticipant).toHaveBeenCalledWith('42', 'tok123');
    expect(screen.getByText('Editando suas respostas anteriores')).toBeInTheDocument();
    expect(screen.getByLabelText('Go').checked).toBe(true);

    window.history.pushState({}, '', '/answer/42');
  });

  it('mostra aviso e segue o fluxo normal quando o link de edição é inválido', async () => {
    window.history.pushState({}, '', '/answer/42?edit=invalido');
    mockLoad();
    api.public.events.getParticipant.mockRejectedValue(new Error('Link de edição inválido ou expirado.'));
    render(Answer, { props: { id: '42' } });

    expect(
      await screen.findByText(/Link de edição inválido ou expirado/)
    ).toBeInTheDocument();
    expect(screen.getByLabelText('E-mail')).toBeInTheDocument();

    window.history.pushState({}, '', '/answer/42');
  });
});
