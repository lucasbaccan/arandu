export class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

async function request(path, options = {}) {
  const res = await fetch(path, {
    credentials: 'same-origin',
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
        })
    }
  },
  public: {
    events: {
      get: (id) => request(`/api/public/events/${id}`),
      submit: (id, body) => request(`/api/public/events/${id}/submit`, { method: 'POST', body })
    }
  }
};
