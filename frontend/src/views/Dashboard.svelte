<script>
  import { user, logout } from '../lib/authStore.js';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { formatDate } from '../lib/formatDate.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import CopyButton from '../components/CopyButton.svelte';

  let events = [];
  let loading = true;
  let error = '';

  const statusLabels = {
    PREPARATION: 'Em preparação',
    OPEN_FOR_ANSWERS: 'Coletando respostas',
    CLOSED_FOR_ANSWERS: 'Respostas encerradas',
    PRESENTING: 'Ao vivo',
    FINISHED: 'Finalizado'
  };

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

  function statusLabel(status) {
    return statusLabels[status] || status;
  }
</script>

<main class="page page-wide">
  <div class="dash-head">
    <div class="dash-brand">
      <img class="dash-logo" src="/img/arandu-logo.png" alt="Arandu" />
      <div>
        <h1 class="dash-title">Meus eventos</h1>
        <p class="dash-user">Olá, {($user && $user.name) || '…'}</p>
      </div>
    </div>
    <div class="dash-actions">
      <Button on:click={goCreate}>Novo evento</Button>
      <Button variant="secondary" on:click={handleLogout}>Sair</Button>
    </div>
  </div>

  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if error}
    <p class="form-error">{error}</p>
  {:else if events.length === 0}
    <Card>
      <h2>Nenhum evento ainda</h2>
      <p class="subtitle">Crie seu primeiro evento para começar uma dinâmica.</p>
      <Button block on:click={goCreate}>Criar evento</Button>
    </Card>
  {:else}
    <div class="event-list">
      {#each events as ev (ev.id)}
        <Card clickable on:click={() => openEvent(ev.id)}>
          <div class="event-row">
            <div class="event-info">
              <h2 class="event-title">{ev.title}</h2>
              <p class="subtitle">
                <span class="badge badge-{ev.status.toLowerCase()}">{statusLabel(ev.status)}</span>
                <span class="text-muted">· criado em {formatDate(ev.createdAt)}</span>
              </p>
            </div>
            <div class="pin-chip" title="Código de acesso">
              <span class="pin-chip-label">PIN</span>
              <div class="pin-chip-value">
                <strong>{ev.pinCode.toUpperCase()}</strong>
                <CopyButton text={ev.pinCode.toUpperCase()} label="Copiar PIN" />
              </div>
            </div>
          </div>
        </Card>
      {/each}
    </div>
  {/if}
</main>
