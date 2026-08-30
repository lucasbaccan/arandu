<script>
  import { onMount } from 'svelte';
  import { route, navigate } from '../lib/router.js';
  import { user } from '../lib/authStore.js';
  import { theme } from '../lib/themeStore.js';

  function go(path) {
    return (e) => {
      e.preventDefault();
      navigate(path);
    };
  }

  // Atalhos internos: o mapa do site (todas as URLs) e o guia de componentes
  // (design). Estes dois são o acesso rápido desta tela — ver /mapa-do-site para o
  // mapa completo.
  const shortcuts = [
    {
      path: '/mapa-do-site',
      title: 'Mapa do site',
      desc: 'Todas as URLs da aplicação, agrupadas por fluxo.'
    },
    {
      path: '/tela',
      title: 'Guia de componentes',
      desc: 'Cores, tipografia e componentes prontos — a referência de design.'
    }
  ];

  let pathname = '';
  let search = '';
  let size = '';

  onMount(() => {
    pathname = window.location.pathname;
    search = window.location.search;
    const updateSize = () => (size = `${window.innerWidth}×${window.innerHeight}`);
    updateSize();
    window.addEventListener('resize', updateSize);
    return () => window.removeEventListener('resize', updateSize);
  });

  $: userLine = $user ? `${$user.name} <${$user.email}>` : 'deslogado';
</script>

<main class="page page-wide debug">
  <div class="dbg-head">
    <h1>Debug — Arandu</h1>
    <p class="text-muted">
      Tela interna de desenvolvimento, fora do fluxo do produto. Serve de porta de
      entrada para as páginas de referência internas: o mapa do site (todos os
      URLs) e o guia de componentes (design).
    </p>
  </div>

  <section class="dbg-section">
    <h2>Atalhos internos</h2>
    <div class="dbg-shortcuts">
      {#each shortcuts as s (s.path)}
        <a class="dbg-card" href={s.path} on:click={go(s.path)}>
          <strong>{s.title}</strong>
          <span class="text-muted">{s.desc}</span>
          <code>{s.path}</code>
        </a>
      {/each}
    </div>
  </section>

  <section class="dbg-section">
    <h2>Estado da aplicação</h2>
    <dl class="dbg-state">
      <div class="dbg-state-row">
        <dt>Pathname</dt>
        <dd><code>{pathname || '—'}</code></dd>
      </div>
      <div class="dbg-state-row">
        <dt>Query string</dt>
        <dd><code>{search || '(vazia)'}</code></dd>
      </div>
      <div class="dbg-state-row">
        <dt>Rota (store)</dt>
        <dd><code>{$route}</code></dd>
      </div>
      <div class="dbg-state-row">
        <dt>Tema</dt>
        <dd><code>{$theme}</code></dd>
      </div>
      <div class="dbg-state-row">
        <dt>Usuário</dt>
        <dd><code>{userLine}</code></dd>
      </div>
      <div class="dbg-state-row">
        <dt>Janela</dt>
        <dd><code>{size || '—'}</code></dd>
      </div>
    </dl>
  </section>

  <footer class="dbg-footer">
    <button
      type="button"
      class="dbg-back"
      on:click={() => (history.length > 1 ? history.back() : navigate('/mapa-do-site'))}
    >‹ Voltar</button>
    <span class="dbg-footer-sep" aria-hidden="true">·</span>
    <a href="/mapa-do-site" on:click={go('/mapa-do-site')}>Ver mapa completo do site ›</a>
  </footer>
</main>

<style>
  .debug {
    align-items: stretch;
    gap: 24px;
    padding-bottom: 64px;
  }

  .dbg-head h1 {
    margin: 0 0 8px;
    font-size: 2rem;
  }

  .dbg-head p {
    max-width: 680px;
    margin: 0;
  }

  .dbg-section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    padding: 20px 24px;
  }

  .dbg-section h2 {
    margin: 0 0 14px;
    font-size: 1.15rem;
  }

  .dbg-shortcuts {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(240px, 100%), 1fr));
    gap: 14px;
  }

  .dbg-card {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 14px 16px;
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    background: var(--bg-input);
    color: var(--text);
    text-decoration: none;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .dbg-card:hover {
    border-color: var(--accent);
    background: var(--accent-soft);
    text-decoration: none;
  }

  .dbg-card strong {
    font-size: 0.9375rem;
  }

  .dbg-card span {
    font-size: 0.8125rem;
  }

  .dbg-card code {
    margin-top: 4px;
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  .dbg-state {
    margin: 0;
    display: flex;
    flex-direction: column;
  }

  .dbg-state-row {
    display: grid;
    grid-template-columns: 140px 1fr;
    gap: 12px;
    padding: 9px 0;
    border-top: 1px solid var(--border);
    align-items: baseline;
  }

  .dbg-state-row:first-child {
    border-top: none;
  }

  .dbg-state-row dt {
    font-size: 0.6875rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--text-muted);
  }

  .dbg-state-row dd {
    margin: 0;
    min-width: 0;
  }

  .dbg-state-row code {
    font-size: 0.85rem;
    overflow-wrap: break-word;
  }

  .dbg-footer {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    row-gap: 4px;
    gap: 10px;
    font-size: 0.8125rem;
    font-weight: 700;
  }

  .dbg-footer a,
  .dbg-back {
    display: inline-flex;
    align-items: center;
    min-height: 44px;
    padding: 0 12px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 700;
    cursor: pointer;
  }

  .dbg-footer a:hover,
  .dbg-back:hover {
    color: var(--accent);
  }

  .dbg-back:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .dbg-footer-sep {
    color: var(--border-strong);
  }

  @media (max-width: 520px) {
    .dbg-state-row {
      grid-template-columns: 1fr;
      gap: 2px;
    }
  }
</style>
