<script>
  import { navigate } from '../lib/router.js';
  import { api } from '../lib/api.js';
  import Button from '../components/Button.svelte';

  let code = '';
  let codeError = '';
  let checking = false;

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
      const { id } = await api.public.events.resolvePin(trimmed);
      navigate(`/audience/${id}?pin=${encodeURIComponent(trimmed)}`);
    } catch (e) {
      codeError = e.status === 404 ? 'Código não encontrado.' : e.message;
    } finally {
      checking = false;
    }
  }
</script>

<main class="home">
  <div class="home-content">
    <img class="home-logo" src="/img/arandu-completo.png" alt="Arandu" />
    <h1 class="home-heading">Qual é o código do evento?</h1>
    <p class="home-tagline">Sem conta, sem instalação. O organizador mostra o código na tela.</p>

    <form class="code-pill-form" novalidate on:submit|preventDefault={joinWithCode}>
      <div class="code-pill" class:invalid={!!codeError}>
        <input
          class="code-pill-input"
          bind:value={code}
          placeholder="EX: DEV-TEAM"
          autocomplete="off"
        />
        <Button type="submit" size="lg" disabled={checking || codeEmpty}>
          {checking ? '…' : 'Entrar'}
        </Button>
      </div>
      {#if codeError}<p class="form-error home-code-error">{codeError}</p>{/if}
    </form>

    <p class="home-hint">Recebeu um link? Ele já leva você direto ao evento.</p>

    <div class="home-organizer">
      <div class="home-divider">
        <span>organizador</span>
      </div>
      <div class="home-actions">
        <Button variant="outline" block on:click={() => navigate('/login')}>Entrar na conta</Button>
        <Button variant="secondary" block on:click={() => navigate('/register')}>
          Criar conta
        </Button>
      </div>
    </div>
  </div>
</main>

<style>
  .home {
    flex: 1;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .home-content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    width: min(560px, 92vw);
    text-align: center;
  }

  .home-logo {
    width: min(300px, 70vw);
    height: auto;
    margin-bottom: 8px;
  }

  .home-heading {
    margin: 0;
    font-size: clamp(1.7rem, 4vw, 2.6rem);
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .home-tagline {
    margin: 0;
    color: var(--text-muted);
    font-size: 1.05rem;
  }

  .code-pill-form {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
    margin-top: 14px;
  }

  .code-pill {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 8px 8px 8px 20px;
    background: var(--bg-elev);
    border: 2px solid var(--accent);
    border-radius: 12px;
    box-shadow: var(--shadow);
  }

  .code-pill.invalid {
    border-color: var(--danger);
  }

  .code-pill-input {
    flex: 1;
    min-width: 0;
    padding: 12px 0;
    border: none;
    background: transparent;
    font-family: var(--font-ui);
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text);
    text-transform: uppercase;
    outline: none;
  }

  .home-code-error {
    margin: 0;
    text-align: center;
  }

  .home-hint {
    margin: 2px 0 0;
    font-size: 0.85rem;
    color: var(--text-muted);
  }

  .home-organizer {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: 100%;
    margin-top: 20px;
  }

  .home-divider {
    display: flex;
    align-items: center;
    gap: 12px;
    color: var(--text-muted);
    font-size: 0.8rem;
  }

  .home-divider::before,
  .home-divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--border-strong);
  }

  .home-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .home-actions :global(.btn) {
    flex: 1;
    min-width: 160px;
  }

  .code-pill-form :global(.btn),
  .home-actions :global(.btn) {
    transition:
      background 0.15s ease,
      color 0.15s ease,
      border-color 0.15s ease,
      transform 0.15s ease,
      box-shadow 0.15s ease;
  }

  .code-pill-form :global(.btn:hover:not(:disabled)),
  .home-actions :global(.btn:hover:not(:disabled)) {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(43, 0, 187, 0.25);
  }
</style>
