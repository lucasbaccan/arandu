<script>
  import { theme, toggleTheme } from '../lib/themeStore.js';

  // 'square' = topbar do organizador (raio 10), 'round' = canto das telas
  // públicas (pílula sobre o gradiente).
  export let shape = 'square';
</script>

<button
  type="button"
  class="theme-toggle"
  class:round={shape === 'round'}
  aria-label="Alternar tema"
  title={$theme === 'dark' ? 'Usar tema claro' : 'Usar tema escuro'}
  on:click={toggleTheme}
>
  {$theme === 'dark' ? '☀' : '☾'}
</button>

<style>
  .theme-toggle {
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.9375rem;
    line-height: 1;
    cursor: pointer;
    transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
  }

  .theme-toggle:hover {
    color: var(--text);
    border-color: var(--accent);
    background: var(--surface-muted);
  }

  .theme-toggle.round {
    border-radius: 999px;
    border-color: var(--border-strong);
    background: color-mix(in srgb, var(--bg-elev) 70%, transparent);
  }

  /* Alvo de toque mínimo em telas de toque (34px -> 44px). */
  @media (pointer: coarse) {
    .theme-toggle {
      width: 44px;
      height: 44px;
    }
  }
</style>
