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
 * Não vale para /api/auth/* (o 401 ali é a resposta esperada de quem não está
 * logado) nem para /api/public/* (token de plateia, que não é a sessão do dono).
 */
function isSessionRoute(path) {
  return path.startsWith('/api/') && !path.startsWith('/api/auth/') && !path.startsWith('/api/public/');
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
  config: () => request('/api/auth/config'),
  me: () => request('/api/auth/me'),
  register: (body) => request('/api/auth/register', { method: 'POST', body }),
  login: (body) => request('/api/auth/login', { method: 'POST', body }),
  logout: () => request('/api/auth/logout', { method: 'POST' }),
  events: {
    create: (body) => request('/api/events', { method: 'POST', body }),
    list: () => request('/api/events'),
    get: (id) => request(`/api/events/${id}`),
    update: (id, body) => request(`/api/events/${id}`, { method: 'PATCH', body }),
    questions: {
      list: (id) => request(`/api/events/${id}/questions`),
      create: (id, body) => request(`/api/events/${id}/questions`, { method: 'POST', body }),
      update: (id, questionId, body) =>
        request(`/api/events/${id}/questions/${questionId}`, { method: 'PATCH', body }),
      reorder: (id, questionIds) =>
        request(`/api/events/${id}/questions/order`, {
          method: 'PUT',
          body: { questionIds }
        }),
      remove: (id, questionId) =>
        request(`/api/events/${id}/questions/${questionId}`, { method: 'DELETE' })
    },
    responses: {
      list: (id) => request(`/api/events/${id}/responses`),
      updateAnswer: (id, participantId, questionId, body) =>
        request(`/api/events/${id}/responses/${participantId}/answers/${questionId}`, {
          method: 'PATCH',
          body
        }),
      updatePhoto: (id, participantId, body) =>
        request(`/api/events/${id}/responses/${participantId}/photo`, { method: 'PATCH', body })
    },
    live: {
      setQuestion: (id, questionId) =>
        request(`/api/events/${id}/live/question`, { method: 'POST', body: { questionId } }),
      reveal: (id, questionId, participantId) =>
        request(`/api/events/${id}/live/reveal`, { method: 'POST', body: { questionId, participantId } }),
      unreveal: (id, questionId, participantId) =>
        request(`/api/events/${id}/live/unreveal`, { method: 'POST', body: { questionId, participantId } }),
      revealAll: (id, questionId) =>
        request(`/api/events/${id}/live/reveal-all`, { method: 'POST', body: { questionId } }),
      reset: (id, questionId) =>
        request(`/api/events/${id}/live/reset`, { method: 'POST', body: { questionId } }),
      resetAll: (id) => request(`/api/events/${id}/live/reset-all`, { method: 'POST' }),
      setBlanked: (id, blanked) =>
        request(`/api/events/${id}/live/blank`, { method: 'POST', body: { blanked } }),
      setAnswersHidden: (id, hidden) =>
        request(`/api/events/${id}/live/hide-answers`, { method: 'POST', body: { hidden } }),
      setNamesHidden: (id, hidden) =>
        request(`/api/events/${id}/live/hide-names`, { method: 'POST', body: { hidden } }),
      setMessage: (id, message) =>
        request(`/api/events/${id}/live/message`, { method: 'POST', body: { message } }),
      setInteractionsEnabled: (id, enabled) =>
        request(`/api/events/${id}/live/interactions`, { method: 'POST', body: { enabled } }),
      adminState: (id) => request(`/api/events/${id}/live/state`),
      adminStreamUrl: (id) => apiUrl(`/api/events/${id}/live/stream`),
      dismissQA: (id, messageId) =>
        request(`/api/events/${id}/live/qa/${messageId}/dismiss`, { method: 'POST' }),
      presentationState: (id) => request(`/api/events/${id}/live/presentation/state`),
      presentationStreamUrl: (id) => apiUrl(`/api/events/${id}/live/presentation/stream`)
    }
  },
  public: {
    events: {
      resolvePin: (pin) => request(`/api/public/events/by-pin?pin=${encodeURIComponent(pin)}`),
      get: (id) => request(`/api/public/events/${id}`),
      getParticipant: (id, token) =>
        request(`/api/public/events/${id}/participant?token=${encodeURIComponent(token)}`),
      submit: (id, body) => request(`/api/public/events/${id}/submit`, { method: 'POST', body }),
      live: {
        join: (id, body) => request(`/api/public/events/${id}/live/join`, { method: 'POST', body }),
        state: (id, token) =>
          request(`/api/public/events/${id}/live/state?token=${encodeURIComponent(token)}`),
        streamUrl: (id, token) =>
          apiUrl(`/api/public/events/${id}/live/stream?token=${encodeURIComponent(token)}`),
        react: (id, token, emoji) =>
          request(`/api/public/events/${id}/live/react?token=${encodeURIComponent(token)}`, {
            method: 'POST',
            body: { emoji }
          }),
        submitQuestion: (id, token, text) =>
          request(`/api/public/events/${id}/live/qa?token=${encodeURIComponent(token)}`, {
            method: 'POST',
            body: { text }
          })
      }
    }
  }
};
