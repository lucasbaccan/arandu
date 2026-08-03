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
  let submitting = false;

  function validate() {
    if (!name.trim()) return 'Informe seu nome.';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      return 'Informe um e-mail válido.';
    }
    const min = $authConfig.minPasswordLength;
    if (password.length < min) {
      return `A senha deve ter pelo menos ${min} caracteres.`;
    }
    if (password !== confirm) return 'As senhas não conferem.';
    return '';
  }

  async function handleSubmit() {
    error = validate();
    if (error) return;
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
  <Card title="Criar conta" subtitle="Comece a organizar suas dinâmicas.">
    <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
      <Input label="Nome" bind:value={name} placeholder="Seu nome" autocomplete="name" required />
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
        autocomplete="new-password"
        hint={`Mínimo de ${$authConfig.minPasswordLength} caracteres`}
        required
      />
      <Input
        label="Confirmar senha"
        type="password"
        bind:value={confirm}
        autocomplete="new-password"
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
