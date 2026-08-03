<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { showToast } from '../lib/toastStore.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Input from '../components/Input.svelte';

  const statusLabels = {
    PREPARATION: 'Em preparação',
    OPEN_FOR_ANSWERS: 'Coletando respostas',
    CLOSED_FOR_ANSWERS: 'Respostas encerradas',
    PRESENTING: 'Ao vivo',
    FINISHED: 'Finalizado'
  };

  let loading = true;
  let error = '';
  let notFound = false;

  let title = '';
  let pinCode = '';
  let changePin = false;
  let showRanking = false;
  let status = '';
  let submitting = false;

  async function load() {
    try {
      const { event } = await api.events.get(id);
      title = event.title;
      pinCode = event.pinCode;
      showRanking = event.configShowRanking;
      status = event.status;
    } catch (e) {
      if (e.status === 404) {
        notFound = true;
      } else {
        error = e.message;
      }
    } finally {
      loading = false;
    }
  }
  load();

  function validate() {
    if (!title.trim()) return 'Informe o título do evento.';
    if (changePin && !/^[a-zA-Z0-9_-]{1,25}$/.test(pinCode.trim())) {
      return 'O PIN deve ter 1 a 25 caracteres: letras, números, _ ou -.';
    }
    return '';
  }

  async function handleSubmit() {
    error = validate();
    if (error) return;
    submitting = true;
    try {
      await api.events.update(id, {
        title: title.trim(),
        pinCode: changePin ? pinCode.trim() : '',
        configShowRanking: showRanking
      });
      changePin = false;
      showToast('Alterações salvas!');
    } catch (e) {
      showToast(e.message, 'error');
      error = e.message;
    } finally {
      submitting = false;
    }
  }

  function back() {
    navigate('/dashboard');
  }

  function statusLabel(s) {
    return statusLabels[s] || s;
  }
</script>

<main class="page">
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if notFound}
    <Card title="Evento não encontrado">
      <p class="subtitle">Ele pode ter sido removido ou você não tem acesso.</p>
      <Button block variant="secondary" on:click={back}>Voltar</Button>
    </Card>
  {:else}
    <Card title="Editar evento">
      <p class="subtitle">
        <span class="badge badge-{status.toLowerCase()}">{statusLabel(status)}</span>
        <span class="text-muted">· PIN atual: <strong>{pinCode}</strong></span>
      </p>

      <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
        <Input
          label="Título"
          bind:value={title}
          placeholder="Ex: Conecta DevOps 2026"
          autocomplete="off"
          required
        />

        <label class="field check">
          <input type="checkbox" bind:checked={changePin} />
          <span>Definir novo PIN</span>
        </label>
        {#if changePin}
          <Input
            label="Novo PIN"
            bind:value={pinCode}
            placeholder="Ex: dev-team"
            hint="1 a 25 caracteres: letras, números, _ ou -"
          />
        {/if}

        <label class="field check">
          <input type="checkbox" bind:checked={showRanking} />
          <span>Exibir ranking de pontos</span>
        </label>

        {#if error}
          <p class="form-error">{error}</p>
        {/if}

        <div class="form-actions">
          <Button type="submit" disabled={submitting}>
            {submitting ? 'Salvando…' : 'Salvar alterações'}
          </Button>
          <Button variant="secondary" on:click={back} disabled={submitting}>Voltar</Button>
        </div>
      </form>
    </Card>
  {/if}
</main>

<style>
  .form-actions {
    display: flex;
    gap: 10px;
  }
</style>
