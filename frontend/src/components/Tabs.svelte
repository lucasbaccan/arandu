<script>
  /*
   * Abas sublinhadas com contador — o mesmo padrão no Editar evento
   * (Perguntas · Respostas) e no trilho do Palco (Perguntas · Q&A).
   */
  import { createEventDispatcher } from 'svelte';

  // tabs: [{ value, label, count? }]
  export let tabs = [];
  export let value = '';
  export let compact = false;

  const dispatch = createEventDispatcher();

  function select(v) {
    if (v === value) return;
    value = v;
    dispatch('change', v);
  }
</script>

<div class="tabs" class:compact role="tablist">
  {#each tabs as tab (tab.value)}
    <button
      type="button"
      role="tab"
      class="tab"
      class:active={value === tab.value}
      aria-selected={value === tab.value}
      on:click={() => select(tab.value)}
    >
      {tab.label}
      {#if tab.count !== undefined && tab.count !== null}
        <!-- fora do nome acessível: o contador é reforço visual, o rótulo já
             identifica a aba para leitores de tela -->
        <span class="tab-count" aria-hidden="true">{tab.count}</span>
      {/if}
    </button>
  {/each}
  <span class="tabs-spacer"></span>
  <slot name="actions" />
</div>

<style>
  .tabs {
    display: flex;
    align-items: center;
    gap: 24px;
    border-bottom: 1px solid var(--border);
  }

  .tabs.compact {
    gap: 20px;
    padding: 0 16px;
  }

  .tab {
    padding: 0 2px 10px;
    border: none;
    border-bottom: 2px solid transparent;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.875rem;
    font-weight: 700;
    cursor: pointer;
    transition: color 0.15s ease, border-color 0.15s ease;
  }

  .tabs.compact .tab {
    padding: 14px 2px 11px;
  }

  /* Alvo de toque das abas (~30px -> ~42px) em telas de toque. */
  @media (pointer: coarse) {
    .tab {
      padding: 12px 2px 14px;
    }
  }

  .tab:hover {
    color: var(--text);
  }

  .tab.active {
    color: var(--text);
    border-bottom-color: var(--accent);
  }

  .tab-count {
    color: var(--text-subtle);
    font-weight: 800;
  }

  .tabs-spacer {
    flex: 1;
  }

  /* Mobile: abas + ação em uma linha estouram; deixa quebrar. */
  @media (max-width: 640px) {
    .tabs {
      flex-wrap: wrap;
      gap: 8px 16px;
    }
  }
</style>
