<script>
  import { navigate } from '../lib/router.js';
  import { api } from '../lib/api.js';
  import PublicShell from '../components/PublicShell.svelte';

  let code = '';
  let codeError = '';
  let checking = false;
  let focused = false;

  $: codeEmpty = !code.trim();

  async function joinWithCode() {
    codeError = '';
    const trimmed = code.trim();
    if (!trimmed) {
      codeError = 'Informe o código do evento.';
      return;
    }
    checking = true;
    try {
      const { id } = await api.publico.eventos.resolverPin(trimmed);
      navigate(`/plateia/${id}?pin=${encodeURIComponent(trimmed)}`);
    } catch (e) {
      codeError = e.status === 404 ? 'Código não encontrado.' : e.message;
    } finally {
      checking = false;
    }
  }
</script>

<main class="inicio">
  <PublicShell>
    <h1 class="entry-title">Qual é o código do evento?</h1>
    <p class="entry-sub">Sem conta, sem instalação. O organizador mostra o código na tela.</p>

    <form class="code-form" novalidate on:submit|preventDefault={joinWithCode}>
      <div class="code-pill" class:invalid={!!codeError} class:focused>
        <input
          class="code-input"
          bind:value={code}
          on:focus={() => (focused = true)}
          on:blur={() => (focused = false)}
          on:input={() => (codeError = '')}
          placeholder="DEV-TEAM"
          aria-label="Código do evento"
          autocomplete="off"
        />
        <button type="submit" class="btn btn-primary" disabled={checking || codeEmpty}>
          {checking ? '…' : 'Entrar'}
        </button>
      </div>
      <p class="code-note" class:error={!!codeError}>
        {codeError || 'Recebeu um link? Ele já leva você direto ao evento.'}
      </p>
    </form>

    <div class="divider"><span>organizador</span></div>

    <div class="organizer-actions">
      <button type="button" class="btn btn-secondary" on:click={() => navigate('/entrar')}>
        Entrar na conta
      </button>
      <button type="button" class="btn btn-ghost" on:click={() => navigate('/criar-conta')}>
        Criar conta
      </button>
    </div>
  </PublicShell>
</main>

<style>
  .inicio {
    flex: 1;
    display: flex;
    flex-direction: column;
    text-align: center;
  }

  .entry-title {
    margin: 0;
    font-size: 2rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.15;
  }

  .entry-sub {
    margin: 8px 0 0;
    color: var(--text-muted);
    font-size: 0.9375rem;
    line-height: 1.5;
  }

  .code-form {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: 24px;
  }

  .code-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 6px 6px 18px;
    background: var(--bg-elev);
    border: 1.5px solid var(--border-strong);
    border-radius: 12px;
    box-shadow: var(--shadow);
    transition: border-color 0.15s ease;
  }

  .code-pill.focused {
    border-color: var(--accent);
  }

  .code-pill.invalid {
    border-color: var(--danger);
  }

  .code-input {
    flex: 1;
    min-width: 0;
    padding: 10px 0;
    border: none;
    background: transparent;
    font-family: var(--font-ui);
    font-size: 1.25rem;
    font-weight: 700;
    letter-spacing: 0.06em;
    color: var(--text);
    text-transform: uppercase;
    outline: none;
  }

  .code-input::placeholder {
    color: var(--text-subtle);
  }

  .code-pill .btn {
    padding: 12px 22px;
  }

  .code-note {
    margin: 0;
    font-size: 0.8125rem;
    color: var(--text-muted);
  }

  .code-note.error {
    color: var(--danger);
    font-weight: 700;
  }

  .divider {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 32px;
    color: var(--text-muted);
    font-size: 0.6875rem;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  .divider::before,
  .divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--border);
  }

  .organizer-actions {
    display: flex;
    gap: 10px;
    margin-top: 16px;
  }

  .organizer-actions .btn {
    flex: 1;
  }

  @media (max-width: 520px) {
    .organizer-actions {
      flex-direction: column;
    }
  }
</style>
