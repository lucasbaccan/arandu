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
    list: () => request('/api/events')
  }
};
