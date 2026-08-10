<script>
  import { navigate } from '../lib/router.js';
  import { api } from '../lib/api.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';

  let code = '';
  let codeError = '';
  let checking = false;

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
    <p class="home-tagline">Dinâmicas de grupo ao vivo</p>

    <form class="code-form" novalidate on:submit|preventDefault={joinWithCode}>
      <Input
        label="Código do evento"
        bind:value={code}
        error={codeError}
        placeholder="Ex: DEV-TEAM"
        uppercase
        required
      />
      <Button type="submit" block disabled={checking}>
        {checking ? 'Verificando…' : 'Entrar na apresentação'}
      </Button>
    </form>

    <div class="home-divider">
      <span>ou</span>
    </div>

    <div class="home-actions">
      <Button variant="outline" on:click={() => navigate('/login')}>Entrar</Button>
      <Button variant="outline" on:click={() => navigate('/register')}>
        Criar conta
      </Button>
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
    gap: 12px;
  }

  .home-logo {
    width: min(460px, 85vw);
    height: auto;
  }

  .code-form {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: min(280px, 85vw);
    margin-top: 8px;
  }

  .home-divider {
    display: flex;
    align-items: center;
    gap: 10px;
    width: min(240px, 70vw);
    margin: 4px 0;
    color: var(--text-muted);
    font-size: 0.8rem;
  }

  .home-divider::before,
  .home-divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--border);
  }

  .home-actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .home-actions :global(.btn) {
    min-width: 240px;
  }

  .code-form :global(.btn),
  .home-actions :global(.btn) {
    transition:
      background 0.15s ease,
      color 0.15s ease,
      border-color 0.15s ease,
      transform 0.15s ease,
      box-shadow 0.15s ease;
  }

  .code-form :global(.btn:hover:not(:disabled)),
  .home-actions :global(.btn:hover:not(:disabled)) {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(43, 0, 187, 0.25);
  }
</style>
