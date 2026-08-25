import { notifyUnauthorized } from './sessionExpired.js';

export class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

// Base URL da API — configurável via VITE_BACKEND_URL (ex: https://api.oracle.lucasbaccan.com.br)
// quando o frontend roda separado do backend (ex: Vercel). Sem a env, usa paths
// relativos (backend serve o frontend no mesmo domínio).
const API_BASE = (import.meta.env.VITE_BACKEND_URL || '').replace(/\/+$/, '');
const apiUrl = (path) => `${API_BASE}${path}`;

/*
 * Sessão expirada/derrubada no meio do uso: sem isto, um 401 vira texto de erro
 * dentro da tela ("Sessão inválida.") e a pessoa fica olhando um dashboard vazio
 * sem saber que precisa entrar de novo. Quem reage é o authStore, via
 * sessionExpired.js.
 * Não vale para /api/conta/* (o 401 ali é a resposta esperada de quem não está
 * logado) nem para /api/publico/* (token de plateia, que não é a sessão do dono).
 */
function isSessionRoute(path) {
  return path.startsWith('/api/') && !path.startsWith('/api/conta/') && !path.startsWith('/api/publico/');
}

async function request(path, options = {}) {
  const res = await fetch(apiUrl(path), {
    credentials: 'include',
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    },
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined
  });

  if (res.status === 204) {
    return null;
  }

  let data = {};
  try {
    data = await res.json();
  } catch {
    // corpo não-JSON
  }

  if (!res.ok) {
    if (res.status === 401 && isSessionRoute(path)) {
      notifyUnauthorized();
    }
    throw new ApiError(res.status, data.error || 'Erro inesperado. Tente novamente.');
  }
  return data;
}

export const api = {
  configuracao: () => request('/api/conta/configuracao'),
  eu: () => request('/api/conta/eu'),
  criarConta: (body) => request('/api/conta/criar-conta', { method: 'POST', body }),
  entrar: (body) => request('/api/conta/entrar', { method: 'POST', body }),
  sair: () => request('/api/conta/sair', { method: 'POST' }),
  eventos: {
    criar: (body) => request('/api/eventos', { method: 'POST', body }),
    listar: () => request('/api/eventos'),
    buscar: (id) => request(`/api/eventos/${id}`),
    atualizar: (id, body) => request(`/api/eventos/${id}`, { method: 'PATCH', body }),
    perguntas: {
      listar: (id) => request(`/api/eventos/${id}/perguntas`),
      criar: (id, body) => request(`/api/eventos/${id}/perguntas`, { method: 'POST', body }),
      atualizar: (id, questionId, body) =>
        request(`/api/eventos/${id}/perguntas/${questionId}`, { method: 'PATCH', body }),
      reordenar: (id, questionIds) =>
        request(`/api/eventos/${id}/perguntas/ordem`, {
          method: 'PUT',
          body: { questionIds }
        }),
      remover: (id, questionId) =>
        request(`/api/eventos/${id}/perguntas/${questionId}`, { method: 'DELETE' })
    },
    respostas: {
      listar: (id) => request(`/api/eventos/${id}/respostas`),
      atualizarResposta: (id, participantId, questionId, body) =>
        request(`/api/eventos/${id}/respostas/${participantId}/resposta/${questionId}`, {
          method: 'PATCH',
          body
        }),
      atualizarFoto: (id, participantId, body) =>
        request(`/api/eventos/${id}/respostas/${participantId}/foto`, { method: 'PATCH', body })
    },
    aoVivo: {
      definirPergunta: (id, questionId) =>
        request(`/api/eventos/${id}/ao-vivo/pergunta`, { method: 'POST', body: { questionId } }),
      revelar: (id, questionId, participantId) =>
        request(`/api/eventos/${id}/ao-vivo/revelar`, { method: 'POST', body: { questionId, participantId } }),
      ocultar: (id, questionId, participantId) =>
        request(`/api/eventos/${id}/ao-vivo/ocultar`, { method: 'POST', body: { questionId, participantId } }),
      revelarTodos: (id, questionId) =>
        request(`/api/eventos/${id}/ao-vivo/revelar-todos`, { method: 'POST', body: { questionId } }),
      reiniciar: (id, questionId) =>
        request(`/api/eventos/${id}/ao-vivo/reiniciar`, { method: 'POST', body: { questionId } }),
      reiniciarTudo: (id) => request(`/api/eventos/${id}/ao-vivo/reiniciar-tudo`, { method: 'POST' }),
      definirEmBranco: (id, blanked) =>
        request(`/api/eventos/${id}/ao-vivo/em-branco`, { method: 'POST', body: { blanked } }),
      ocultarRespostas: (id, hidden) =>
        request(`/api/eventos/${id}/ao-vivo/ocultar-respostas`, { method: 'POST', body: { hidden } }),
      ocultarNomes: (id, hidden) =>
        request(`/api/eventos/${id}/ao-vivo/ocultar-nomes`, { method: 'POST', body: { hidden } }),
      definirModoDensidade: (id, mode) =>
        request(`/api/eventos/${id}/ao-vivo/modo-densidade`, { method: 'POST', body: { mode } }),
      definirMensagem: (id, message) =>
        request(`/api/eventos/${id}/ao-vivo/mensagem`, { method: 'POST', body: { message } }),
      definirInteracoes: (id, enabled) =>
        request(`/api/eventos/${id}/ao-vivo/interacoes`, { method: 'POST', body: { enabled } }),
      estadoAdmin: (id) => request(`/api/eventos/${id}/ao-vivo/estado`),
      urlFluxoAdmin: (id) => apiUrl(`/api/eventos/${id}/ao-vivo/fluxo`),
      dispensarPergunta: (id, messageId) =>
        request(`/api/eventos/${id}/ao-vivo/perguntas/${messageId}/dispensar`, { method: 'POST' }),
      estadoApresentacao: (id) => request(`/api/eventos/${id}/ao-vivo/apresentacao/estado`),
      urlFluxoApresentacao: (id) => apiUrl(`/api/eventos/${id}/ao-vivo/apresentacao/fluxo`)
    }
  },
  publico: {
    eventos: {
      resolverPin: (pin) => request(`/api/publico/eventos/por-pin?pin=${encodeURIComponent(pin)}`),
      buscar: (id) => request(`/api/publico/eventos/${id}`),
      buscarParticipante: (id, token) =>
        request(`/api/publico/eventos/${id}/participante?token=${encodeURIComponent(token)}`),
      enviar: (id, body) => request(`/api/publico/eventos/${id}/enviar`, { method: 'POST', body }),
      aoVivo: {
        entrar: (id, body) => request(`/api/publico/eventos/${id}/ao-vivo/entrar`, { method: 'POST', body }),
        estado: (id, token) =>
          request(`/api/publico/eventos/${id}/ao-vivo/estado?token=${encodeURIComponent(token)}`),
        urlFluxo: (id, token) =>
          apiUrl(`/api/publico/eventos/${id}/ao-vivo/fluxo?token=${encodeURIComponent(token)}`),
        reagir: (id, token, emoji) =>
          request(`/api/publico/eventos/${id}/ao-vivo/reagir?token=${encodeURIComponent(token)}`, {
            method: 'POST',
            body: { emoji }
          }),
        enviarPergunta: (id, token, text, clientId) =>
          request(`/api/publico/eventos/${id}/ao-vivo/perguntas?token=${encodeURIComponent(token)}`, {
            method: 'POST',
            body: { text, clientId }
          }),
        removerPergunta: (id, token, messageId, clientId) =>
          request(`/api/publico/eventos/${id}/ao-vivo/perguntas/${messageId}?token=${encodeURIComponent(token)}`, {
            method: 'DELETE',
            body: { clientId }
          })
      }
    }
  }
};
