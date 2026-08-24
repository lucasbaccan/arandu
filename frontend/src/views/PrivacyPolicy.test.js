import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/svelte';
import PrivacyPolicy from './PrivacyPolicy.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

import { navigate } from '../lib/router.js';

describe('Política de privacidade', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('mostra os itens de privacidade', () => {
    render(PrivacyPolicy);
    expect(
      screen.getByRole('heading', { name: 'Seus dados servem para uma coisa só: a dinâmica' })
    ).toBeInTheDocument();
    expect(screen.getByText('O que pedimos')).toBeInTheDocument();
    expect(screen.getByText('Quanto tempo fica guardado')).toBeInTheDocument();
  });

  it('navega de volta para "Como funciona" pelo rodapé', async () => {
    render(PrivacyPolicy);
    await fireEvent.click(screen.getByRole('link', { name: 'Como funciona o Arandu' }));
    expect(navigate).toHaveBeenCalledWith('/como-funciona');
  });

  it('navega para o início pelo rodapé', async () => {
    render(PrivacyPolicy);
    await fireEvent.click(screen.getByRole('link', { name: '‹ Voltar ao início' }));
    expect(navigate).toHaveBeenCalledWith('/');
  });
});
