import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/svelte';
import HelpButton from './HelpButton.svelte';

vi.mock('../lib/router.js', () => ({
  navigate: vi.fn()
}));

import { navigate } from '../lib/router.js';

describe('HelpButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('abre o popover de ajuda ao clicar', async () => {
    render(HelpButton);
    expect(screen.queryByText('Como funciona o Arandu')).not.toBeInTheDocument();
    await fireEvent.click(screen.getByRole('button', { name: 'Ajuda' }));
    expect(screen.getByText('Como funciona o Arandu')).toBeInTheDocument();
  });

  it('navega para o tutorial pelo "Saiba mais"', async () => {
    render(HelpButton);
    await fireEvent.click(screen.getByRole('button', { name: 'Ajuda' }));
    await fireEvent.click(screen.getByRole('link', { name: 'Saiba mais →' }));
    expect(navigate).toHaveBeenCalledWith('/como-funciona');
  });

  it('navega para o mapa do site pelo "Mapa do site"', async () => {
    render(HelpButton);
    await fireEvent.click(screen.getByRole('button', { name: 'Ajuda' }));
    await fireEvent.click(screen.getByRole('link', { name: 'Mapa do site →' }));
    expect(navigate).toHaveBeenCalledWith('/mapa-do-site');
  });
});
