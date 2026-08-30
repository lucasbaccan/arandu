<script>
  import { showToast } from '../lib/toastStore.js';

  export let text = '';
  export let label = 'Copiar';

  function legacyCopy(value) {
    const textarea = document.createElement('textarea');
    textarea.value = value;
    textarea.style.position = 'fixed';
    textarea.style.top = '-1000px';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();
    const ok = document.execCommand('copy');
    document.body.removeChild(textarea);
    if (!ok) throw new Error('execCommand copy falhou');
  }

  async function copy() {
    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
      } else {
        legacyCopy(text);
      }
      showToast('Copiado!');
    } catch {
      try {
        legacyCopy(text);
        showToast('Copiado!');
      } catch {
        showToast('Não foi possível copiar.', 'error');
      }
    }
  }
</script>

<button
  type="button"
  class="copy-btn"
  title={label}
  aria-label={label}
  on:click|stopPropagation={copy}
>
  <svg viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <rect x="7" y="7" width="10" height="10" rx="2" stroke="currentColor" stroke-width="1.5" />
    <path
      d="M13 7V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h2"
      stroke="currentColor"
      stroke-width="1.5"
    />
  </svg>
</button>

<style>
  .copy-btn {
    position: relative;
    width: 26px;
    height: 26px;
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 7px;
    border: 1px solid var(--accent);
    background: rgba(43, 0, 187, 0.1);
    color: var(--accent);
    cursor: pointer;
    transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
  }

  .copy-btn:hover {
    color: #fff;
    border-color: var(--accent-hover);
    background: var(--accent-hover);
  }

  .copy-btn svg {
    width: 14px;
    height: 14px;
  }

  /* Alvo de toque de ~40px sem mudar o visual (26px reais + 7px de folga). */
  .copy-btn::after {
    content: '';
    position: absolute;
    inset: -7px;
  }
</style>
