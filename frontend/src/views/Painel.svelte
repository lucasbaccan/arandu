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
      const data = await api.eventos.listar();
      eventos = data.events || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
  loadEvents();

  function abrirNovoEvento() {
    navigate('/evento/novo');
  }

  function abrirEvento(id) {
    navigate(`/evento/${id}`);
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
                <span class="evento-title-text" title={ev.title}>{ev.title}</span>
                <div class="evento-counts">
                  <span class="text-muted evento-question-count">
                    {ev.questionCount} pergunta{ev.questionCount === 1 ? '' : 's'}
                  </span>
                  <span class="evento-responses">
                    {ev.participantCount} respost{ev.participantCount === 1 ? 'a' : 'as'}
                  </span>
                </div>
              </div>
              <span class="col-status">
                <Chip
                  dot
                  label={statusInfo(ev.status).label}
                  tint={statusInfo(ev.status).tint}
                  color={statusInfo(ev.status).color}
                />
              </span>
              <span class="meta-group">
                <span class="col-pin">
                  <PinChip pin={ev.pinCode} />
                </span>
              </span>
              <span class="col-created text-muted">{formatDate(ev.createdAt)}</span>
              <span class="col-chevron evento-chevron">❯</span>
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
    gap: 8px;
  }

  .filter {
    min-height: 40px;
    padding: 9px 12px;
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
    grid-template-columns: minmax(160px, 1fr) 170px 140px 110px 24px;
    gap: 16px;
    padding: 0 18px;
    color: var(--text-muted);
    font-size: 0.6875rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }

  .eventos-table-body {
    display: flex;
    flex-direction: column;
    gap: 8px;
    /* Sem overflow próprio: a lista cresce com o conteúdo e a página rola
       como um todo — scrollbar do navegador na borda, não uma barra flutuando
       no meio da tela. flex: none impede o flex-shrink do pai (shell-body)
       de comprimir a lista quando o conteúdo passa da altura da janela. */
    flex: none;
  }

  .eventos-empty-search {
    margin: 0;
  }

  /* A cor da situação vira a borda esquerda da linha — o mesmo sinal do chip,
     legível de relance na lista inteira. */
  .evento-row {
    display: grid;
    grid-template-columns: minmax(160px, 1fr) 170px 140px 110px 24px;
    gap: 16px;
    align-items: center;
    min-height: 64px;
    /* flex-shrink: 0 — a linha é item de .eventos-table-body (coluna flex).
       O card nunca comprime abaixo da altura do conteúdo. */
    flex-shrink: 0;
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

  .evento-counts {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }

  .evento-counts .evento-responses {
    flex-shrink: 0;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .col-created {
    font-size: 0.875rem;
  }

  .evento-chevron {
    color: var(--text-muted);
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
    .col-created {
      display: none;
    }
  }

  @media (max-width: 640px) {
    /* A busca volta no mobile (o CrumbBar quebra em 2 linhas nessa faixa):
       linha própria ao lado do botão "Novo evento". */
    .search {
      display: flex;
      flex: 1;
      min-width: 0;
    }

    /* 16px evita o zoom automático do iOS ao focar (a busca não é .field). */
    .search input {
      font-size: 1rem;
    }
  }

  /* Chip de status longo ("Coletando respostas") cabe na coluna de 150px da
     tabela 3 colunas (521-900px): trunca em vez de encostar no chevron. */
  .col-status {
    min-width: 0;
  }

  .col-status :global(.chip) {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Filtros com alvo de toque de 44px em telas de toque. */
  @media (pointer: coarse) {
    .filter {
      min-height: 44px;
    }
  }

  /* Telefone: o grid de 3 colunas estoura (mínimo intrínseco 362px vs
     296-350px disponíveis em 360-414px). A linha vira um card empilhado com
     PIN de volta — dado essencial para o organizador ao vivo. Perguntas e
     respostas ficam juntas na segunda linha do título.
     Vai até 640px: em 521-640px o grid de 3 colunas deixava o título estreito
     e o chevron em posição errada. */
  @media (max-width: 640px) {
    .eventos-table-head {
      display: none;
    }

    .evento-row {
      position: relative;
      display: flex;
      flex-flow: row wrap;
      align-items: center;
      gap: 6px 10px;
      min-height: 0;
      padding: 12px 40px 12px 16px;
    }

    .evento-row .col-title {
      flex: 1 1 100%;
      min-width: 0;
      padding-right: 30px;
    }

    .evento-title-text {
      white-space: normal;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
    }

    .evento-chevron {
      position: absolute;
      top: 50%;
      right: 14px;
      transform: translateY(-50%);
      font-size: 1.125rem;
      color: var(--text-muted);
    }

    /* Com PIN longo, o valor trunca com ellipsis sem quebrar a linha. */
    .evento-row .meta-group {
      display: flex;
      align-items: center;
      flex-wrap: nowrap;
      gap: 0;
      min-width: 0;
      max-width: 100%;
    }

    .evento-row .meta-group .col-pin {
      flex: 1 1 auto;
      min-width: 0;
      display: inline-flex;
      align-items: center;
    }

    .evento-row .col-pin::before {
      content: 'PIN';
      font-size: 0.6875rem;
      font-weight: 800;
      letter-spacing: 0.1em;
      color: var(--text-muted);
      margin-right: 6px;
    }
  }
</style>
