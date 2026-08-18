<script>
  import { login } from '../lib/authStore.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import PublicShell from '../components/PublicShell.svelte';

  let email = '';
  let password = '';
  let error = '';
  let submitting = false;

  function validate() {
    if (!email.trim()) return 'Informe seu e-mail.';
    if (!password) return 'Informe sua senha.';
    return '';
  }

  async function handleSubmit() {
    error = validate();
    if (error) return;
    submitting = true;
    try {
      await login(email.trim(), password);
      navigate('/dashboard');
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<main class="login">
  <PublicShell backLabel="Início" backHref="/">
    <div class="card entry-card">
      <div class="card-head">
        <h1>Entrar na sua conta</h1>
        <p class="card-sub">Para criar e conduzir eventos.</p>
      </div>
      <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
        <Input
          label="E-mail"
          type="email"
          bind:value={email}
          placeholder="seu@melhor.email"
          autocomplete="email"
          required
        />
        <Input
          label="Senha"
          type="password"
          bind:value={password}
          placeholder="••••••••"
          autocomplete="current-password"
          required
        />
        {#if error}
          <p class="form-error">{error}</p>
        {/if}
        <Button type="submit" block disabled={submitting}>
          {submitting ? 'Entrando…' : 'Entrar'}
        </Button>
      </form>
      <p class="switch">
        Não tem conta?
        <a href="/register" on:click|preventDefault={() => navigate('/register')}>Criar conta</a>
      </p>
    </div>
  </PublicShell>
</main>

<style>
  .login {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .entry-card {
    max-width: none;
  }

  .card-head {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .card-sub {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.875rem;
  }

  .switch a {
    font-weight: 700;
  }
</style>
