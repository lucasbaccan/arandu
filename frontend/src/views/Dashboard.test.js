import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import Dashboard from './Dashboard.svelte';
import { user } from '../lib/authStore.js';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

vi.mock('../lib/api.js', () => ({
  api: {
    logout: vi.fn()
  }
}));

import { navigate } from '../lib/router.js';
import { api } from '../lib/api.js';

describe('Dashboard', () => {
  beforeEach(() => {
    user.set(null);
    vi.clearAllMocks();
  });

  it('mostra o nome do usuário logado', () => {
    user.set({ id: '1', email: 'ana@x.com', name: 'Ana' });
    render(Dashboard);
    expect(screen.getByText('Olá, Ana')).toBeInTheDocument();
  });

  it('não mostra "undefined" quando o usuário ainda não carregou', () => {
    render(Dashboard);
    expect(screen.queryByText(/undefined/)).not.toBeInTheDocument();
  });

  it('sai da conta e volta para a home', async () => {
    user.set({ id: '1', email: 'ana@x.com', name: 'Ana' });
    render(Dashboard);

    await fireEvent.click(screen.getByRole('button', { name: 'Sair' }));

    await waitFor(() => expect(api.logout).toHaveBeenCalled());
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/'));
  });
});
