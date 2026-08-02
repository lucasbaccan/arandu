<script>
  import { login } from '../lib/authStore.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Input from '../components/Input.svelte';

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

<main class="page">
  <Card title="Entrar" subtitle="Acesse sua conta para gerenciar eventos.">
    <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
      <Input
        label="E-mail"
        type="email"
        bind:value={email}
        placeholder="voce@empresa.com"
        autocomplete="email"
        required
      />
      <Input
        label="Senha"
        type="password"
        bind:value={password}
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
      <a href="/register" on:click|preventDefault={() => navigate('/register')}>
        Criar conta
      </a>
    </p>
  </Card>
</main>
