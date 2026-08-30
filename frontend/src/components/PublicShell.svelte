<script>
  /*
   * Shell das telas públicas de entrada (Home, Login, Cadastro, Responder —
   * identificação). Marca centralizada a 168px, coluna de 440px e os ícones
   * fixos no canto (voltar à esquerda, tema e ajuda à direita) que substituem
   * a topbar nas telas sem barra.
   */
  import { navigate } from '../lib/router.js';
  import ThemeToggle from './ThemeToggle.svelte';
  import HelpButton from './HelpButton.svelte';

  // Rótulo do botão de voltar no canto superior esquerdo. Vazio = sem botão.
  export let backLabel = '';
  export let backHref = '/';
  export let width = 'var(--entry-col)';
  // Logo completa acima da coluna — todas as telas de entrada usam 168px.
  export let showLogo = true;

  function goBack(e) {
    e.preventDefault();
    navigate(backHref);
  }
</script>

<div class="public-shell">
  {#if backLabel}
    <a class="corner-back" href={backHref} on:click={goBack}>‹ {backLabel}</a>
  {/if}
  <div class="corner-tools">
    <ThemeToggle shape="round" />
    <HelpButton shape="round" />
  </div>

  <div class="public-center">
    {#if showLogo}
      <button
        type="button"
        class="public-logo-btn"
        title="Ir para o início"
        aria-label="Ir para o início"
        on:click={() => navigate('/')}
      >
        <img class="public-logo" src="/img/arandu-completo.png" alt="Arandu" />
      </button>
    {/if}
    <div class="public-col" style="width: {width}">
      <slot />
    </div>
  </div>
</div>

<style>
  .public-shell {
    position: relative;
    flex: 1;
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  .corner-back {
    position: absolute;
    z-index: 2;
    top: 16px;
    left: 20px;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 7px 14px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    background: color-mix(in srgb, var(--bg-elev) 70%, transparent);
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 700;
    text-decoration: none;
  }

  .corner-back:hover {
    color: var(--accent);
    border-color: var(--accent);
    text-decoration: none;
  }

  .corner-tools {
    position: absolute;
    z-index: 2;
    top: 16px;
    right: 20px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .public-center {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    /* safe center: com overflow (telas baixas/landscape) o topo fica
       acessível em vez de cortado pelo centering. */
    justify-content: safe center;
    gap: 24px;
    padding: 72px 24px 40px;
  }

  .public-logo-btn {
    padding: 0;
    border: none;
    background: none;
    cursor: pointer;
  }

  .public-logo {
    width: 168px;
    max-width: 60vw;
    height: auto;
  }

  .public-col {
    max-width: 100%;
    display: flex;
    flex-direction: column;
  }

  /* Alvo de toque mínimo em telas de toque. */
  @media (pointer: coarse) {
    .corner-back {
      min-height: 44px;
      padding: 10px 16px;
    }
  }

  /* Telas baixas (ex.: iPhone SE em retrato): menos rolagem, logo menor.
     60px (não 56) para o logo não roçar nos controles de canto em 320px. */
  @media (max-height: 700px) {
    .public-center {
      padding-top: 60px;
      gap: 16px;
    }

    .public-logo {
      width: 120px;
    }
  }
</style>
