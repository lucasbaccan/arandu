<script>
  import { navigate } from '../lib/router.js';

  // Menu flutuante da tela de apresentação (/stage/{id}/present*): escolhe o
  // nº de colunas do placar. 'Auto' (/present) deixa fitDensity decidir a
  // combinação escala × colunas; 1–4 (/present1..4) forçam, pra comparar.
  // Só aparece nessas telas, que já são exclusivas do organizador
  // (StagePresentation).
  export let id = '';
  export let active = 0; // 0 = auto, 1..4 = colunas forçadas

  const OPTIONS = [
    { cols: 0, label: 'Auto', suffix: '' },
    { cols: 1, label: '1 coluna', suffix: '1' },
    { cols: 2, label: '2 colunas', suffix: '2' },
    { cols: 3, label: '3 colunas', suffix: '3' },
    { cols: 4, label: '4 colunas', suffix: '4' },
  ];

  let open = false;
</script>

<div class="variant-menu" class:open>
  <button
    type="button"
    class="variant-toggle"
    on:click={() => (open = !open)}
    aria-expanded={open}
    aria-label="Escolher o número de colunas da apresentação"
  >
    ⚙
  </button>
  {#if open}
    <div class="variant-options">
      {#each OPTIONS as opt (opt.cols)}
        <button
          type="button"
          class="variant-option"
          class:active={active === opt.cols}
          on:click={() => navigate(`/stage/${id}/present${opt.suffix}`)}
        >
          {opt.label}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .variant-menu {
    position: fixed;
    right: 18px;
    bottom: 18px;
    z-index: 20;
    display: flex;
    flex-direction: column-reverse;
    align-items: flex-end;
    gap: 10px;
  }

  .variant-toggle {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: var(--bg-elev);
    color: var(--text-muted);
    font-size: 1.125rem;
    cursor: pointer;
    box-shadow: var(--shadow);
    opacity: 0.55;
    transition: opacity 0.15s ease, transform 0.15s ease;
  }

  .variant-toggle:hover,
  .variant-menu.open .variant-toggle {
    opacity: 1;
  }

  .variant-toggle:hover {
    transform: scale(1.05);
  }

  .variant-options {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    box-shadow: var(--shadow);
  }

  .variant-option {
    padding: 7px 14px;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--text);
    font-size: 0.875rem;
    font-weight: 600;
    text-align: right;
    cursor: pointer;
    white-space: nowrap;
  }

  .variant-option:hover {
    background: var(--surface-muted);
  }

  .variant-option.active {
    background: var(--accent-soft);
    border-color: var(--accent);
    color: var(--accent);
  }
</style>
