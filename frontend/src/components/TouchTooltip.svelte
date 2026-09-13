<script>
  /*
   * Faz o `title="..."` (usado em ícone-botões e nomes truncados por toda a
   * app pra dar a dica que no desktop aparece no hover) também aparecer no
   * toque: pressionar e segurar mostra a dica; soltar antes do delay ainda
   * funciona como um tap normal (clica). Soltar depois de já ter "espiado" a
   * dica não dispara o clique — evita ações destrutivas (remover pergunta,
   * por ex.) por engano só por ter segurado o dedo um pouco mais.
   * Montado uma única vez em App.svelte: nenhum botão precisa mudar.
   */
  import { onMount, onDestroy } from 'svelte';

  const DELAY = 450;
  const MOVE_TOLERANCE = 10;

  let bubble = null;
  let timer = null;
  let startX = 0;
  let startY = 0;
  let target = null;
  let peeked = false;

  function clearTimer() {
    if (timer) {
      clearTimeout(timer);
      timer = null;
    }
  }

  function hide() {
    bubble = null;
  }

  function showFor(el) {
    const text = el.getAttribute('title');
    if (!text) return;
    const rect = el.getBoundingClientRect();
    const above = rect.top > 60;
    bubble = {
      text,
      x: rect.left + rect.width / 2,
      y: above ? rect.top - 8 : rect.bottom + 8,
      above
    };
    peeked = true;
  }

  function onTouchStart(e) {
    const el = e.target.closest('[title]');
    hide();
    clearTimer();
    peeked = false;
    if (!el) {
      target = null;
      return;
    }
    target = el;
    const t = e.touches[0];
    startX = t.clientX;
    startY = t.clientY;
    timer = setTimeout(() => showFor(el), DELAY);
  }

  function onTouchMove(e) {
    if (!timer && !bubble) return;
    const t = e.touches[0];
    if (Math.abs(t.clientX - startX) > MOVE_TOLERANCE || Math.abs(t.clientY - startY) > MOVE_TOLERANCE) {
      clearTimer();
      hide();
      target = null;
    }
  }

  function onTouchEnd(e) {
    clearTimer();
    if (peeked && target) {
      e.preventDefault();
    }
    hide();
    target = null;
    peeked = false;
  }

  function onTouchCancel() {
    clearTimer();
    hide();
    target = null;
    peeked = false;
  }

  onMount(() => {
    document.addEventListener('touchstart', onTouchStart, { passive: true });
    document.addEventListener('touchmove', onTouchMove, { passive: true });
    document.addEventListener('touchend', onTouchEnd);
    document.addEventListener('touchcancel', onTouchCancel);
  });

  onDestroy(() => {
    clearTimer();
    document.removeEventListener('touchstart', onTouchStart);
    document.removeEventListener('touchmove', onTouchMove);
    document.removeEventListener('touchend', onTouchEnd);
    document.removeEventListener('touchcancel', onTouchCancel);
  });
</script>

{#if bubble}
  <div class="touch-tooltip" style="left: {bubble.x}px; top: {bubble.y}px; transform: translate(-50%, {bubble.above ? '-100%' : '0'});">
    {bubble.text}
  </div>
{/if}

<style>
  .touch-tooltip {
    position: fixed;
    z-index: 1000;
    max-width: min(240px, 80vw);
    padding: 6px 10px;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: var(--bg-elev);
    color: var(--text);
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 600;
    line-height: 1.3;
    text-align: center;
    box-shadow: var(--shadow-hover);
    pointer-events: none;
    white-space: normal;
  }
</style>
