<script>
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { formatDate } from '../lib/formatDate.js';
  import { statusInfo } from '../lib/eventStatus.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Chip from '../components/Chip.svelte';
  import CrumbBar from '../components/CrumbBar.svelte';
  import PinChip from '../components/PinChip.svelte';
  import TopBar from '../components/TopBar.svelte';

  const FILTERS = [
    { value: 'all', label: 'Todos' },
    { value: 'live', label: 'Ao vivo' },
    { value: 'done', label: 'Finalizados' }
  ];

  let eventos = [];
  let loading = true;
  let error = '';
  let searchValue = '';
  let filter = 'all';

  $: eventosFiltrados = eventos.filter((ev) => {
    const matchesFilter =
      filter === 'all' ||
      (filter === 'live' && ev.status === 'PRESENTING') ||
      (filter === 'done' && ev.status === 'FINISHED');
    const q = searchValue.trim().toLowerCase();
    const matchesSearch =
      !q || ev.title.toLowerCase().includes(q) || ev.pinCode.toLowerCase().includes(q);
    return matchesFilter && matchesSearch;
  });

  $: liveCount = eventosFiltrados.filter((ev) => ev.status === 'PRESENTING').length;

  async function loadEvents() {
    try {
      const data = await api.events.list();
      eventos = data.events || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
  loadEvents();

  function abrirNovoEvento() {
    navigate('/eventos/novo');
  }

  function abrirEvento(id) {
    navigate(`/eventos/${id}`);
  }
</script>

<main class="shell">
  <TopBar area="Organizador" />

  <CrumbBar crumbs={[{ label: 'Eventos' }]}>
    <svelte:fragment slot="actions">
      <span class="search">
        <span class="search-icon" aria-hidden="true">⌕</span>
        <input bind:value={searchValue} placeholder="Buscar por título ou PIN" aria-label="Buscar por título ou PIN" />
      </span>
      <Button size="sm" on:click={abrirNovoEvento}>Novo evento</Button>
    </svelte:fragment>
  </CrumbBar>

  <div class="shell-body">
    {#if loading}
      <p class="text-muted">Carregando…</p>
    {:else if error}
      <p class="form-error">{error}</p>
    {:else if eventos.length === 0}
      <div class="empty-wrap">
        <Card>
          <h2>Nenhum evento ainda</h2>
          <p class="subtitle">Crie seu primeiro evento para começar uma dinâmica.</p>
          <Button block on:click={abrirNovoEvento}>Criar evento</Button>
        </Card>
      </div>
    {:else}
      <div class="eventos-head">
        <div class="eventos-head-text">
          <h1 class="section-title">Eventos</h1>
          <span class="section-sub">
            {eventosFiltrados.length} evento{eventosFiltrados.length === 1 ? '' : 's'} · {liveCount} ao vivo
          </span>
        </div>
        <div class="filters">
          {#each FILTERS as f (f.value)}
            <button
              type="button"
              class="filter"
              class:active={filter === f.value}
              aria-pressed={filter === f.value}
              on:click={() => (filter = f.value)}
            >
              {f.label}
            </button>
          {/each}
        </div>
      </div>

      <div class="eventos-table-head">
        <span class="col-title">Evento</span>
        <span class="col-status">Situação</span>
        <span class="col-pin">PIN</span>
        <span class="col-responses">Respostas</span>
        <span class="col-created">Criado</span>
        <span class="col-chevron"></span>
      </div>

      {#if eventosFiltrados.length === 0}
        <p class="text-muted eventos-empty-search">Nenhum evento encontrado.</p>
      {:else}
        <div class="eventos-table-body">
          {#each eventosFiltrados as ev (ev.id)}
            <div
              class="evento-row"
              style="--row-accent:{statusInfo(ev.status).color}"
              role="button"
              tabindex="0"
              on:click={() => abrirEvento(ev.id)}
              on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && abrirEvento(ev.id)}
            >
              <div class="col-title evento-title-cell">
                <span class="evento-title-text">{ev.title}</span>
                <span class="text-muted evento-question-count">
                  {ev.questionCount} pergunta{ev.questionCount === 1 ? '' : 's'}
                </span>
              </div>
              <span class="col-status">
                <Chip
                  dot
                  label={statusInfo(ev.status).label}
                  tint={statusInfo(ev.status).tint}
                  color={statusInfo(ev.status).color}
                />
              </span>
              <span class="col-pin">
                <PinChip pin={ev.pinCode} />
              </span>
              <span class="col-responses evento-people">{ev.participantCount}</span>
              <span class="col-created text-muted">{formatDate(ev.createdAt)}</span>
              <span class="col-chevron evento-chevron">›</span>
            </div>
          {/each}
        </div>
      {/if}
    {/if}
  </div>
</main>

<style>
  .search {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 260px;
    padding: 7px 12px;
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: var(--surface-muted);
    color: var(--text-subtle);
    font-size: 0.8125rem;
    transition: border-color 0.15s ease;
  }

  .search:focus-within {
    border-color: var(--accent);
  }

  .search input {
    flex: 1;
    min-width: 0;
    border: none;
    outline: none;
    background: transparent;
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    color: var(--text);
  }

  .search-icon {
    flex-shrink: 0;
  }

  .shell {
    flex: none;
    height: 100vh;
    height: 100dvh;
  }

  .empty-wrap {
    display: flex;
    justify-content: center;
    padding-top: 24px;
  }

  .eventos-head {
    display: flex;
    align-items: flex-end;
    gap: 16px;
    flex-wrap: wrap;
  }

  .eventos-head-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1;
  }

  .filters {
    display: flex;
    gap: 6px;
  }

  .filter {
    padding: 6px 12px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--bg-elev);
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.75rem;
    font-weight: 700;
    cursor: pointer;
    transition: background 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  }

  .filter:hover {
    border-color: var(--accent);
  }

  .filter.active {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
  }

  .eventos-table-head {
    display: grid;
    grid-template-columns: minmax(160px, 1fr) 170px 140px 100px 110px 24px;
    gap: 16px;
    padding: 0 18px;
    color: var(--text-subtle);
    font-size: 0.6875rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }

  .eventos-table-body {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  .eventos-empty-search {
    margin: 0;
  }

  /* A cor da situação vira a borda esquerda da linha — o mesmo sinal do chip,
     legível de relance na lista inteira. */
  .evento-row {
    display: grid;
    grid-template-columns: minmax(160px, 1fr) 170px 140px 100px 110px 24px;
    gap: 16px;
    align-items: center;
    min-height: 64px;
    padding: 12px 18px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-left: 3px solid var(--row-accent);
    border-radius: var(--radius-row);
    cursor: pointer;
    transition: border-color 0.15s ease, box-shadow 0.15s ease, transform 0.1s ease;
  }

  .evento-row:hover,
  .evento-row:focus-visible {
    border-color: var(--accent);
    border-left-color: var(--row-accent);
    box-shadow: var(--shadow-hover);
    outline: none;
  }

  .evento-row:active {
    transform: scale(0.997);
  }

  .evento-title-cell {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }

  .evento-title-text {
    font-size: 0.9375rem;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .evento-question-count {
    font-size: 0.75rem;
  }

  .evento-people {
    font-size: 0.9375rem;
    font-weight: 700;
  }

  .col-created {
    font-size: 0.875rem;
  }

  .evento-chevron {
    color: var(--text-subtle);
    font-size: 1.125rem;
    text-align: right;
  }

  @media (max-width: 900px) {
    .search {
      display: none;
    }

    .eventos-table-head,
    .evento-row {
      grid-template-columns: minmax(120px, 1fr) 150px 24px;
    }

    .col-pin,
    .col-responses,
    .col-created {
      display: none;
    }
  }
</style>
