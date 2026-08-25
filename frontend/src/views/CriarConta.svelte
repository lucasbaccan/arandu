<script>
  import { register, authConfig } from '../lib/authStore.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import PublicShell from '../components/PublicShell.svelte';

  let name = '';
  let email = '';
  let password = '';
  let confirm = '';
  let error = '';
  let fieldErrors = {};
  let submitting = false;

  $: minLength = $authConfig.minPasswordLength;

  // Barra de força: só um retorno visual do quanto a senha passou do mínimo.
  // Não é regra de validação — o backend continua exigindo apenas o mínimo.
  $: strength = (() => {
    if (!password) return { pct: 0, color: 'var(--border)' };
    if (password.length < minLength) return { pct: 25, color: 'var(--danger)' };
    if (password.length < minLength + 4) return { pct: 55, color: 'var(--orange)' };
    return { pct: 100, color: 'var(--success)' };
  })();

  function validate() {
    const errors = {};
    if (!name.trim()) errors.name = 'Informe seu nome.';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      errors.email = 'Informe um e-mail válido.';
    }
    if (password.length < minLength) {
      errors.password = `A senha deve ter pelo menos ${minLength} caracteres.`;
    }
    if (confirm !== password) {
      errors.confirm = 'As senhas não conferem.';
    }
    return errors;
  }

  async function handleSubmit() {
    error = '';
    fieldErrors = validate();
    if (Object.keys(fieldErrors).length) return;
    submitting = true;
    try {
      await register(name.trim(), email.trim(), password);
      navigate('/painel');
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<main class="criar-conta">
  <PublicShell backLabel="Início" backHref="/">
    <div class="card entry-card">
      <div class="card-head">
        <h1>Criar conta de organizador</h1>
        <p class="card-sub">Participantes não precisam de conta — só do código.</p>
      </div>
      <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
        <Input
          label="Nome completo"
          bind:value={name}
          placeholder="Seu nome"
          autocomplete="name"
          error={fieldErrors.name}
          required
        />
        <Input
          label="E-mail"
          type="email"
          bind:value={email}
          placeholder="seu@melhor.email"
          autocomplete="email"
          error={fieldErrors.email}
          required
        />
        <Input
          label="Senha"
          type="password"
          bind:value={password}
          placeholder="••••••••"
          autocomplete="new-password"
          error={fieldErrors.password}
          required
        >
          <span slot="below" class="strength" class:hidden={!!fieldErrors.password}>
            <span class="strength-rail">
              <span class="strength-fill" style="width:{strength.pct}%;background:{strength.color}"></span>
            </span>
            Mínimo de {minLength} caracteres
          </span>
        </Input>
        <Input
          label="Confirmar senha"
          type="password"
          bind:value={confirm}
          placeholder="••••••••"
          autocomplete="new-password"
          error={fieldErrors.confirm}
          required
        />
        {#if error}
          <p class="form-error">{error}</p>
        {/if}
        <Button type="submit" block disabled={submitting}>
          {submitting ? 'Criando conta…' : 'Criar conta'}
        </Button>
      </form>
      <p class="switch">
        Já tem conta?
        <a href="/entrar" on:click|preventDefault={() => navigate('/entrar')}>Entrar</a>
      </p>
    </div>
  </PublicShell>
</main>

<style>
  .criar-conta {
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

  .strength {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .strength.hidden {
    display: none;
  }

  .strength-rail {
    width: 48px;
    height: 4px;
    flex-shrink: 0;
    border-radius: 999px;
    background: var(--border);
    overflow: hidden;
  }

  .strength-fill {
    display: block;
    height: 100%;
    border-radius: 999px;
    transition: width 0.2s ease, background 0.2s ease;
  }

  .switch a {
    font-weight: 700;
  }
</style>
