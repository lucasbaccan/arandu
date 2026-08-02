import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/svelte';
import Home from './Home.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

import { navigate } from '../lib/router.js';

describe('Tela inicial', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('navega para o login ao clicar em Entrar', async () => {
    render(Home);
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    expect(navigate).toHaveBeenCalledWith('/login');
  });

  it('navega para o registro ao clicar em Criar conta', async () => {
    render(Home);
    await fireEvent.click(screen.getByRole('button', { name: 'Criar conta' }));
    expect(navigate).toHaveBeenCalledWith('/register');
  });
});
