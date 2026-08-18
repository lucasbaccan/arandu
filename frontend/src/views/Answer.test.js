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

function mockLoad({
  answersOpen = true,
  allowEdit = true,
  questions = [groupQuestion, openQuestion]
} = {}) {
  api.public.events.get.mockResolvedValue({
    event: { id: '42', title: 'Dinâmica de Testes', answersOpen, allowEdit },
    questions
  });
}

async function identifyAndContinue(name = 'Ana', email = 'ana@exemplo.com') {
  await userEvent.type(screen.getByLabelText('Como quer aparecer'), name);
  await userEvent.type(screen.getByLabelText('E-mail'), email);
  await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
  await fireEvent.click(screen.getByRole('button', { name: 'Continuar mesmo assim' }));
}

describe('Tela de respostas do participante', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('mostra a tela de identificação com o título do evento', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });

    expect(await screen.findByText('Dinâmica de Testes')).toBeInTheDocument();
    expect(screen.getByLabelText('Como quer aparecer')).toBeInTheDocument();
    expect(screen.getByLabelText('E-mail')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Começar' })).toBeInTheDocument();
  });

  it('exige nome e e-mail válidos antes de iniciar', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    expect(await screen.findByText('Informe seu nome.')).toBeInTheDocument();
    expect(screen.getByText('Informe seu e-mail.')).toBeInTheDocument();

    await userEvent.type(screen.getByLabelText('Como quer aparecer'), 'Ana');
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

    await userEvent.type(screen.getByLabelText('Como quer aparecer'), 'Ana');
    await userEvent.type(screen.getByLabelText('E-mail'), 'ana@my_domain.com');
    await fireEvent.click(screen.getByRole('button', { name: 'Começar' }));

    expect(await screen.findByText('Informe um e-mail válido.')).toBeInTheDocument();
    expect(screen.queryByText('Qual sua linguagem favorita?')).not.toBeInTheDocument();
  });

  it('mostra e fecha a tela de privacidade a partir da identificação', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await fireEvent.click(screen.getByRole('button', { name: 'Privacidade e uso dos dados' }));
    expect(
      await screen.findByText('Seus dados servem para uma coisa só: a dinâmica')
    ).toBeInTheDocument();

    await fireEvent.click(screen.getByRole('button', { name: 'Entendi, continuar' }));
    expect(await screen.findByLabelText('E-mail')).toBeInTheDocument();
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

  it('permite avançar sem responder, mas só libera o envio com tudo respondido', async () => {
    mockLoad();
    api.public.events.submit.mockResolvedValue({ ok: true });
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await identifyAndContinue('Ana', 'Participante@Exemplo.com');
    expect(
      await screen.findByText('Qual sua linguagem favorita?', { selector: 'h1' })
    ).toBeInTheDocument();

    // avança mesmo sem selecionar uma opção
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    expect(
      await screen.findByText('Qual sua comida favorita?', { selector: 'h1' })
    ).toBeInTheDocument();

    // avança para a revisão mesmo com a última pergunta em branco
    await fireEvent.click(screen.getByRole('button', { name: 'Revisar' }));
    expect(await screen.findByText('Confira antes de enviar')).toBeInTheDocument();

    // sem todas as respostas, o envio fica bloqueado
    expect(screen.getByRole('button', { name: 'Enviar respostas' })).toBeDisabled();
    expect(api.public.events.submit).not.toHaveBeenCalled();

    // volta pela lista de revisão e responde o que faltou
    await fireEvent.click(screen.getByText('Qual sua linguagem favorita?'));
    await screen.findByText('Qual sua linguagem favorita?', { selector: 'h1' });
    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    await screen.findByText('Qual sua comida favorita?', { selector: 'h1' });

    // texto só com espaços não deve contar como resposta (trim)
    await userEvent.type(screen.getByPlaceholderText('Escreva sua resposta'), '   ');
    await fireEvent.click(screen.getByRole('button', { name: 'Revisar' }));
    expect(screen.getByRole('button', { name: 'Enviar respostas' })).toBeDisabled();

    await fireEvent.click(screen.getByText('Qual sua comida favorita?'));
    await screen.findByText('Qual sua comida favorita?', { selector: 'h1' });
    await userEvent.clear(screen.getByPlaceholderText('Escreva sua resposta'));
    await userEvent.type(screen.getByPlaceholderText('Escreva sua resposta'), '  Pizza  ');
    await fireEvent.click(screen.getByRole('button', { name: 'Revisar' }));

    expect(await screen.findByText('Confira antes de enviar')).toBeInTheDocument();
    expect(screen.getByText('Go')).toBeInTheDocument();
    expect(screen.getByText('Pizza')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Enviar respostas' })).not.toBeDisabled();

    await fireEvent.click(screen.getByRole('button', { name: 'Enviar respostas' }));

    await waitFor(() =>
      expect(api.public.events.submit).toHaveBeenCalledWith('42', {
        email: 'Participante@Exemplo.com',
        name: 'Ana',
        photo: '',
        editToken: '',
        answers: [
          { questionId: 'q1', optionId: 'opt-go', text: '' },
          { questionId: 'q2', optionId: '', text: 'Pizza' }
        ]
      })
    );

    expect(await screen.findByText('Prontinho, Ana!')).toBeInTheDocument();
  });

  it('mantém a resposta ao voltar para a pergunta anterior', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await identifyAndContinue();
    await screen.findByText('Qual sua linguagem favorita?', { selector: 'h1' });

    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    await screen.findByText('Qual sua comida favorita?', { selector: 'h1' });

    await fireEvent.click(screen.getByRole('button', { name: 'Voltar' }));
    await screen.findByText('Qual sua linguagem favorita?', { selector: 'h1' });

    expect(screen.getByLabelText('Go').checked).toBe(true);
  });

  it('permite editar uma resposta a partir da tela de revisão', async () => {
    mockLoad();
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await identifyAndContinue();
    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    await userEvent.type(screen.getByPlaceholderText('Escreva sua resposta'), 'Pizza');
    await fireEvent.click(screen.getByRole('button', { name: 'Revisar' }));
    await screen.findByText('Confira antes de enviar');

    await fireEvent.click(screen.getByText('Qual sua linguagem favorita?'));
    expect(await screen.findByText('Qual sua linguagem favorita?', { selector: 'h1' })).toBeInTheDocument();
    expect(screen.getByLabelText('Go').checked).toBe(true);

    await fireEvent.click(screen.getByLabelText('JS'));
    // pergunta 1 de 2: ainda não é a última, o botão continua "Próxima"
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Revisar' }));
    await screen.findByText('Confira antes de enviar');
    expect(screen.getByText('JS')).toBeInTheDocument();
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

    await identifyAndContinue();
    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Revisar' }));
    await screen.findByText('Confira antes de enviar');
    await fireEvent.click(screen.getByRole('button', { name: 'Enviar respostas' }));

    expect(await screen.findByText('Informe um e-mail válido.')).toBeInTheDocument();
    expect(screen.getByLabelText('E-mail')).toBeInTheDocument();
    expect(screen.queryByText('Confira antes de enviar')).not.toBeInTheDocument();
  });

  it('mostra o link de edição na tela de agradecimento', async () => {
    mockLoad();
    api.public.events.submit.mockResolvedValue({ ok: true, editToken: 'tok123' });
    render(Answer, { props: { id: '42' } });
    await screen.findByText('Dinâmica de Testes');

    await identifyAndContinue();
    await fireEvent.click(screen.getByLabelText('Go'));
    await fireEvent.click(screen.getByRole('button', { name: 'Próxima' }));
    await userEvent.type(screen.getByPlaceholderText('Escreva sua resposta'), 'Pizza');
    await fireEvent.click(screen.getByRole('button', { name: 'Revisar' }));
    await screen.findByText('Confira antes de enviar');
    await fireEvent.click(screen.getByRole('button', { name: 'Enviar respostas' }));

    expect(await screen.findByText('Prontinho, Ana!')).toBeInTheDocument();
    const linkInput = screen.getByDisplayValue(/\/answer\/42\?edit=tok123$/);
    expect(linkInput).toBeInTheDocument();
  });

  it('pré-preenche as respostas ao acessar com um link de edição válido', async () => {
    window.history.pushState({}, '', '/answer/42?edit=tok123');
    mockLoad();
    api.public.events.getParticipant.mockResolvedValue({
      participant: {
        email: 'ana@exemplo.com',
        name: 'Ana',
        photo: '',
        answers: [
          { questionId: 'q1', optionId: 'opt-go', text: '' },
          { questionId: 'q2', optionId: '', text: 'Pizza' }
        ]
      }
    });
    render(Answer, { props: { id: '42' } });

    expect(
      await screen.findByText('Qual sua linguagem favorita?', { selector: 'h1' })
    ).toBeInTheDocument();
    expect(api.public.events.getParticipant).toHaveBeenCalledWith('42', 'tok123');
    expect(screen.getByText('Editando suas respostas anteriores')).toBeInTheDocument();
    expect(screen.getByLabelText('Go').checked).toBe(true);

    window.history.pushState({}, '', '/answer/42');
  });

  it('avisa que a edição está desabilitada sem carregar o formulário', async () => {
    window.history.pushState({}, '', '/answer/42?edit=algum-token');
    mockLoad({ allowEdit: false });
    render(Answer, { props: { id: '42' } });

    expect(
      await screen.findByText(/edição de respostas está desabilitada/i)
    ).toBeInTheDocument();
    expect(api.public.events.getParticipant).not.toHaveBeenCalled();

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
