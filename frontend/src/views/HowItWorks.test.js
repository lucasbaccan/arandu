import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/svelte';
import HowItWorks from './HowItWorks.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

import { navigate } from '../lib/router.js';

describe('Como funciona o Arandu', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('mostra o tutorial para quem participa e para quem organiza', () => {
    render(HowItWorks);
    expect(screen.getByRole('heading', { name: 'Como funciona o Arandu' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Para quem participa' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Para quem organiza' })).toBeInTheDocument();
  });

  it('navega para a política de privacidade pelo rodapé', async () => {
    render(HowItWorks);
    await fireEvent.click(screen.getByRole('link', { name: 'Política de privacidade' }));
    expect(navigate).toHaveBeenCalledWith('/privacidade');
  });

  it('navega para o início pelo rodapé', async () => {
    render(HowItWorks);
    await fireEvent.click(screen.getByRole('link', { name: '‹ Voltar ao início' }));
    expect(navigate).toHaveBeenCalledWith('/');
  });
});
