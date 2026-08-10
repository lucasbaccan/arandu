<script>
  import { register, authConfig } from '../lib/authStore.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Input from '../components/Input.svelte';

  let name = '';
  let email = '';
  let password = '';
  let confirm = '';
  let error = '';
  let fieldErrors = {};
  let submitting = false;

  function validate() {
    const errors = {};
    if (!name.trim()) errors.name = 'Informe seu nome.';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      errors.email = 'Informe um e-mail válido.';
    }
    const min = $authConfig.minPasswordLength;
    if (password.length < min) {
      errors.password = `A senha deve ter pelo menos ${min} caracteres.`;
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
      navigate('/dashboard');
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<main class="page">
  <Card>
    <a
      class="back-logo"
      href="/"
      aria-label="Voltar para o início"
      on:click|preventDefault={() => navigate('/')}
    >
      <img src="/img/arandu-completo.png" alt="Arandu" />
    </a>
    <h1>Criar conta</h1>
    <p class="subtitle">Comece a organizar suas dinâmicas.</p>
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
        autocomplete="new-password"
        hint={`Mínimo de ${$authConfig.minPasswordLength} caracteres`}
        error={fieldErrors.password}
        required
      />
      <Input
        label="Confirmar senha"
        type="password"
        bind:value={confirm}
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
      <a href="/login" on:click|preventDefault={() => navigate('/login')}>Entrar</a>
    </p>
  </Card>
</main>

<style>
  .back-logo {
    display: block;
    text-align: center;
  }

  .back-logo img {
    width: min(180px, 55vw);
    height: auto;
  }
</style>
