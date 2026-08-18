<script>
  /*
   * Faixa de 44px logo abaixo da topbar: carrega o contexto (o caminho
   * "Eventos › Conecta DevOps › Editar") e as ações daquela tela. É o único
   * lugar onde a ação primária da tela mora no shell do organizador.
   */
  import { navigate } from '../lib/router.js';

  // crumbs: [{ label, href? }] — o último item é o atual (peso 700, sem link).
  export let crumbs = [];

  function go(e, href) {
    e.preventDefault();
    navigate(href);
  }
</script>

<div class="crumbbar">
  <nav class="crumbs" aria-label="Caminho">
    {#each crumbs as crumb, i (crumb.label)}
      {#if i > 0}
        <span class="crumb-sep" aria-hidden="true">›</span>
      {/if}
      {#if crumb.href}
        <a class="crumb-link" href={crumb.href} on:click={(e) => go(e, crumb.href)}>{crumb.label}</a>
      {:else}
        <span class="crumb-current" aria-current="page">{crumb.label}</span>
      {/if}
    {/each}
  </nav>
  <slot name="status" />
  <span class="crumb-spacer"></span>
  <slot name="actions" />
</div>

<style>
  .crumbbar {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: var(--crumbbar-h);
    padding: 5px 20px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
    font-size: 0.8125rem;
  }

  .crumbs {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .crumb-link {
    color: var(--text-muted);
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .crumb-link:hover {
    color: var(--accent);
  }

  .crumb-current {
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .crumb-sep {
    color: var(--border-strong);
    flex-shrink: 0;
  }

  .crumb-spacer {
    flex: 1;
  }
</style>
