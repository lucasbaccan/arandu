<script>
  import { user, logout } from '../lib/authStore.js';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { formatDate } from '../lib/formatDate.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';

  let events = [];
  let loading = true;
  let error = '';
  let searchValue = '';

  let userMenuOpen = false;
  let avatarEl;
  let userMenuEl;

  const statusMeta = {
    PREPARATION: { label: 'Em preparação', tint: 'var(--tint-orange)', color: 'var(--orange)' },
    OPEN_FOR_ANSWERS: { label: 'Coletando respostas', tint: 'var(--tint-cyan)', color: 'var(--cyan-hover)' },
    CLOSED_FOR_ANSWERS: { label: 'Respostas encerradas', tint: 'rgba(104, 103, 122, 0.12)', color: 'var(--text-muted)' },
    PRESENTING: { label: 'Ao vivo', tint: 'var(--tint-success)', color: 'var(--success)' },
    FINISHED: { label: 'Finalizado', tint: 'rgba(104, 103, 122, 0.12)', color: 'var(--text-muted)' }
  };

  function statusInfo(status) {
    return statusMeta[status] || { label: status, tint: 'var(--bg-input)', color: 'var(--text-muted)' };
  }

  $: filteredEvents = events.filter((ev) => {
    const q = searchValue.trim().toLowerCase();
    if (!q) return true;
    return ev.title.toLowerCase().includes(q) || ev.pinCode.toLowerCase().includes(q);
  });

  $: liveCount = events.filter((ev) => ev.status === 'PRESENTING').length;

  async function loadEvents() {
    try {
      const data = await api.events.list();
      events = data.events || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
  loadEvents();

  async function handleLogout() {
    await logout();
    navigate('/');
  }

  function goCreate() {
    navigate('/events/new');
  }

  function openEvent(id) {
    navigate(`/events/${id}`);
  }

  function toggleUserMenu() {
    userMenuOpen = !userMenuOpen;
  }

  function closeUserMenu() {
    userMenuOpen = false;
  }

  function handleWindowMousedown(e) {
    if (!userMenuOpen) return;
    if (avatarEl?.contains(e.target) || userMenuEl?.contains(e.target)) return;
    closeUserMenu();
  }

  function logoutFromMenu() {
    closeUserMenu();
    handleLogout();
  }
</script>

<svelte:window on:mousedown={handleWindowMousedown} />

<main class="page page-wide dash-page">
  <div class="dash-topbar">
    <img class="dash-topbar-logo" src="/img/arandu-logo.png" alt="Arandu" />
    <span class="dash-topbar-title">Painel do organizador</span>
    <span class="dash-topbar-spacer"></span>
    <input
      class="dash-search"
      bind:value={searchValue}
      placeholder="Buscar por título ou PIN"
    />
    <Button variant="accent-invert" size="sm" on:click={goCreate}>Novo evento</Button>
    <div class="user-menu-wrap">
      <button
        type="button"
        class="avatar-btn"
        bind:this={avatarEl}
        title={($user && $user.name) || ''}
        aria-haspopup="menu"
        aria-expanded={userMenuOpen}
        on:click={toggleUserMenu}
      >
        {(($user && $user.name) || '?')[0].toUpperCase()}
      </button>
      {#if userMenuOpen}
        <div class="user-menu" role="menu" bind:this={userMenuEl}>
          <p class="user-menu-name">Olá, {($user && $user.name) || '…'}</p>
          <button type="button" class="user-menu-item" role="menuitem" on:click={logoutFromMenu}>
            Sair
          </button>
        </div>
      {/if}
    </div>
  </div>

  <div class="dash-body">
    {#if loading}
      <p class="text-muted">Carregando…</p>
    {:else if error}
      <p class="form-error">{error}</p>
    {:else if events.length === 0}
      <Card>
        <h2>Nenhum evento ainda</h2>
        <p class="subtitle">Crie seu primeiro evento para começar uma dinâmica.</p>
        <Button variant="accent-invert" block on:click={goCreate}>Criar evento</Button>
      </Card>
    {:else}
      <div class="events-summary">
        <h1 class="dash-title">Eventos</h1>
        <span class="text-muted">
          {events.length} evento{events.length === 1 ? '' : 's'} · {liveCount} ao vivo
        </span>
      </div>

      <div class="events-table-head">
        <span class="col-title">Evento</span>
        <span class="col-status">Situação</span>
        <span class="col-pin">PIN</span>
        <span class="col-responses">Respostas</span>
        <span class="col-created">Criado</span>
        <span class="col-chevron"></span>
      </div>

      {#if filteredEvents.length === 0}
        <p class="text-muted events-empty-search">Nenhum evento encontrado para "{searchValue}".</p>
      {:else}
        <div class="events-table-body">
          {#each filteredEvents as ev (ev.id)}
            <div
              class="event-row-grid"
              role="button"
              tabindex="0"
              on:click={() => openEvent(ev.id)}
              on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && openEvent(ev.id)}
            >
              <div class="col-title event-title-cell">
                <span class="event-title-text">{ev.title}</span>
                <span class="text-muted event-question-count">
                  {ev.questionCount} pergunta{ev.questionCount === 1 ? '' : 's'}
                </span>
              </div>
              <span class="col-status">
                <span
                  class="event-status-badge"
                  style="background:{statusInfo(ev.status).tint};color:{statusInfo(ev.status).color}"
                >
                  {statusInfo(ev.status).label}
                </span>
              </span>
              <span class="col-pin event-pin">#{ev.pinCode.toUpperCase()}</span>
              <span class="col-responses event-people">{ev.participantCount}</span>
              <span class="col-created text-muted">{formatDate(ev.createdAt)}</span>
              <span class="col-chevron event-chevron">›</span>
            </div>
          {/each}
        </div>
      {/if}
    {/if}
  </div>
</main>

<style>
  .dash-page {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    padding: 0;
  }

  .dash-topbar {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 14px;
    height: 56px;
    padding: 0 24px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
  }

  .dash-topbar-logo {
    height: 24px;
    width: auto;
  }

  .dash-topbar-title {
    font-size: 0.95rem;
    font-weight: 700;
  }

  .dash-topbar-spacer {
    flex: 1;
  }

  .dash-search {
    width: 220px;
    padding: 8px 12px;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: var(--bg-input);
    font-family: var(--font-ui);
    font-size: 0.85rem;
    color: var(--text);
    outline: none;
  }

  .dash-search:focus {
    border-color: var(--accent);
  }

  .user-menu-wrap {
    position: relative;
  }

  .avatar-btn {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: none;
    background: var(--accent);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.8rem;
    font-weight: 700;
    font-family: var(--font-ui);
    cursor: pointer;
  }

  .user-menu {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    z-index: 20;
    display: flex;
    flex-direction: column;
    min-width: 160px;
    padding: 8px;
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
    border-radius: 10px;
    box-shadow: var(--shadow);
  }

  .user-menu-name {
    margin: 2px 8px 6px;
    font-size: 0.8rem;
    color: var(--text-muted);
    white-space: nowrap;
  }

  .user-menu-item {
    text-align: left;
    padding: 8px 10px;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--danger);
    font-family: var(--font-ui);
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
  }

  .user-menu-item:hover {
    background: var(--tint-danger);
  }

  .dash-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 26px 32px 24px;
  }

  .events-summary {
    display: flex;
    align-items: baseline;
    gap: 12px;
  }

  .dash-title {
    margin: 0;
    font-size: 1.6rem;
  }

  .events-table-head {
    display: grid;
    grid-template-columns: minmax(160px, 1fr) 160px 110px 100px 110px 24px;
    gap: 16px;
    padding: 0 16px;
    color: var(--text-muted);
    font-size: 0.7rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .events-table-body {
    display: flex;
    flex-direction: column;
    gap: 8px;
    overflow-y: auto;
  }

  .events-empty-search {
    margin: 0;
  }

  .event-row-grid {
    display: grid;
    grid-template-columns: minmax(160px, 1fr) 160px 110px 100px 110px 24px;
    gap: 16px;
    align-items: center;
    min-height: 62px;
    padding: 10px 16px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 10px;
    cursor: pointer;
    transition: border-color 0.15s ease;
  }

  .event-row-grid:hover,
  .event-row-grid:focus-visible {
    border-color: var(--accent);
    outline: none;
  }

  .event-title-cell {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .event-title-text {
    font-size: 0.95rem;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .event-question-count {
    font-size: 0.75rem;
  }

  .event-status-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border-radius: 999px;
    font-size: 0.72rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .event-pin {
    font-size: 0.9rem;
    font-weight: 700;
    letter-spacing: 0.1em;
  }

  .event-people {
    font-size: 0.95rem;
    font-weight: 700;
  }

  .event-chevron {
    color: var(--text-muted);
    font-size: 1.1rem;
    text-align: right;
  }

  @media (max-width: 760px) {
    .dash-search {
      display: none;
    }

    .events-table-head {
      grid-template-columns: minmax(120px, 1fr) 100px 24px;
    }

    .event-row-grid {
      grid-template-columns: minmax(120px, 1fr) 100px 24px;
    }

    .col-pin,
    .col-responses,
    .col-created {
      display: none;
    }
  }
</style>
