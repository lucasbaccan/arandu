<script>
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { showToast } from '../lib/toastStore.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import CrumbBar from '../components/CrumbBar.svelte';
  import TopBar from '../components/TopBar.svelte';

  let title = '';
  // 'auto' = PIN numérico gerado na criação; 'custom' = código escolhido aqui.
  let pinMode = 'auto';
  let pinCode = '';
  let error = '';
  let submitting = false;

  $: pinHint =
    pinMode === 'auto'
      ? 'Seis dígitos gerados na criação — dá para trocar depois.'
      : '1 a 25 caracteres: letras, números, _ ou -';

  function validate() {
    if (!title.trim()) return 'Informe o título do evento.';
    if (pinMode === 'custom' && !/^[a-zA-Z0-9_-]{1,25}$/.test(pinCode.trim())) {
      return 'O PIN deve ter 1 a 25 caracteres: letras, números, _ ou -.';
    }
    return '';
  }

  async function handleSubmit() {
    error = validate();
    if (error) return;
    submitting = true;
    try {
      await api.events.create({
        title: title.trim(),
        pinCode: pinMode === 'custom' ? pinCode.trim() : ''
      });
      showToast('Evento criado com sucesso!');
      navigate('/painel');
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<main class="shell">
  <TopBar area="Organizador" />
  <CrumbBar crumbs={[{ label: 'Eventos', href: '/painel' }, { label: 'Novo evento' }]} />

  <div class="shell-body shell-body-center">
    <form class="form-col" novalidate on:submit|preventDefault={handleSubmit}>
      <div class="form-head">
        <h1 class="section-title">Novo evento</h1>
        <p class="section-sub">Dá para mudar tudo depois — inclusive o PIN.</p>
      </div>

      <div class="card form-card">
        <Input
          label="Título do evento"
          bind:value={title}
          placeholder="Ex: Conecta DevOps 2026"
          autocomplete="off"
          required
        />

        <div class="pin-field">
          <span class="label">Código de acesso</span>
          <div class="pin-choices">
            <button
              type="button"
              class="pin-choice"
              class:active={pinMode === 'auto'}
              aria-pressed={pinMode === 'auto'}
              on:click={() => (pinMode = 'auto')}
            >
              <span class="pin-choice-title">Gerar automático</span>
              <!-- Forma, não valor: o PIN só existe depois de criar o evento,
                   então mostrar um número aqui seria promessa falsa. -->
              <span class="pin-choice-value mono-pin">••• •••</span>
            </button>
            <button
              type="button"
              class="pin-choice"
              class:active={pinMode === 'custom'}
              aria-pressed={pinMode === 'custom'}
              on:click={() => (pinMode = 'custom')}
            >
              <span class="pin-choice-title">Personalizado</span>
              <span class="pin-choice-value mono-pin">{pinCode.toUpperCase() || 'DEV-TEAM'}</span>
            </button>
          </div>
          {#if pinMode === 'custom'}
            <Input
              label="PIN personalizado"
              bind:value={pinCode}
              placeholder="Ex: dev-team"
              uppercase
            />
          {/if}
          <span class="hint">{pinHint}</span>
        </div>
      </div>

      {#if error}
        <p class="form-error">{error}</p>
      {/if}

      <div class="form-actions">
        <Button type="submit" disabled={submitting}>
          {submitting ? 'Criando…' : 'Criar evento'}
        </Button>
        <Button
          type="button"
          variant="secondary"
          on:click={() => navigate('/painel')}
          disabled={submitting}
        >
          Cancelar
        </Button>
      </div>
    </form>
  </div>
</main>

<style>
  .form-col {
    width: var(--form-col);
    max-width: 100%;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .form-head {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .form-card {
    max-width: none;
    gap: 18px;
    padding: 24px;
  }

  .pin-field {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .label {
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--text-muted);
  }

  .hint {
    color: var(--text-muted);
    font-size: 0.75rem;
  }

  /*
   * O PIN aparece antes de existir: a escolha vira duas opções lado a lado, em
   * vez de checkbox + parágrafo explicando o que vai acontecer.
   */
  .pin-choices {
    display: flex;
    gap: 10px;
  }

  .pin-choice {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    padding: 12px 14px;
    border: 1.5px solid var(--border);
    border-radius: var(--radius-control);
    background: var(--bg-elev);
    font-family: var(--font-ui);
    text-align: left;
    cursor: pointer;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .pin-choice:hover {
    border-color: var(--accent);
  }

  .pin-choice.active {
    border-color: var(--accent);
    background: var(--tint-purple);
  }

  .pin-choice-title {
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--text-muted);
  }

  .pin-choice.active .pin-choice-title {
    color: var(--text);
  }

  .pin-choice-value {
    font-size: 1.25rem;
    letter-spacing: 0.1em;
    color: var(--text-subtle);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
  }

  .pin-choice.active .pin-choice-value {
    color: var(--accent);
  }

  .form-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  @media (max-width: 560px) {
    .pin-choices {
      flex-direction: column;
    }
  }
</style>
