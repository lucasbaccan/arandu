<script>
  import { navigate } from '../lib/router.js';

  export let shape = 'square';

  let open = false;
  let wrapEl;

  function toggle() {
    open = !open;
  }

  function goToHowItWorks(e) {
    e.preventDefault();
    open = false;
    navigate('/como-funciona');
  }

  function goToSiteMap(e) {
    e.preventDefault();
    open = false;
    navigate('/mapa-do-site');
  }

  function handleWindowMousedown(e) {
    if (!open) return;
    if (wrapEl && wrapEl.contains(e.target)) return;
    open = false;
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') open = false;
  }
</script>

<svelte:window on:mousedown={handleWindowMousedown} on:keydown={handleKeydown} />

<div class="help-wrap" bind:this={wrapEl}>
  <button
    type="button"
    class="help-btn"
    class:round={shape === 'round'}
    aria-label="Ajuda"
    aria-haspopup="dialog"
    aria-expanded={open}
    on:click={toggle}
  >
    ?
  </button>
  {#if open}
    <div class="help-pop" role="dialog" aria-label="Ajuda">
      <p class="help-title">Como funciona o Arandu</p>
      <ul>
        <li>Participante entra com o <strong>código do evento</strong> — sem conta, sem instalação.</li>
        <li>Quem organiza cria o evento, monta as perguntas e conduz a apresentação.</li>
        <li>As respostas só aparecem no telão quando o organizador revelar.</li>
      </ul>
      <a href="/como-funciona" class="help-more" on:click={goToHowItWorks}>Saiba mais →</a>
      <a href="/mapa-do-site" class="help-more" on:click={goToSiteMap}>Mapa do site →</a>
      <p class="help-foot text-muted">Dúvidas sobre um evento: fale com quem organiza.</p>
    </div>
  {/if}
</div>

<style>
  .help-wrap {
    position: relative;
    display: flex;
  }

  .help-btn {
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.9375rem;
    font-weight: 700;
    line-height: 1;
    cursor: pointer;
    transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
  }

  .help-btn:hover {
    color: var(--text);
    border-color: var(--accent);
    background: var(--surface-muted);
  }

  .help-btn.round {
    border-radius: 999px;
    border-color: var(--border-strong);
    background: color-mix(in srgb, var(--bg-elev) 70%, transparent);
  }

  .help-pop {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    z-index: 30;
    width: 280px;
    padding: 14px 16px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    box-shadow: var(--shadow);
    text-align: left;
  }

  .help-title {
    margin: 0 0 8px;
    font-size: 0.875rem;
    font-weight: 800;
  }

  .help-pop ul {
    margin: 0;
    padding-left: 16px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 0.8125rem;
    line-height: 1.45;
    color: var(--text-muted);
  }

  .help-more {
    display: inline-block;
    margin-top: 10px;
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--accent);
  }

  .help-foot {
    margin: 10px 0 0;
    font-size: 0.75rem;
  }
</style>
