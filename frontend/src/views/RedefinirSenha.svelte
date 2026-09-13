<script>
  import { api } from '../lib/api.js';
  import { authConfig } from '../lib/authStore.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import PublicShell from '../components/PublicShell.svelte';

  const token = new URLSearchParams(window.location.search).get('token') || '';

  let checking = true;
  let valid = false;
  let novaSenha = '';
  let confirmacao = '';
  let error = '';
  let ok = false;
  let submitting = false;

  $: minLength = $authConfig.minPasswordLength;

  async function checkToken() {
    if (!token) {
      checking = false;
      valid = false;
      return;
    }
    try {
      await api.publico.redefinirSenha.validarToken(token);
      valid = true;
    } catch {
      valid = false;
    } finally {
      checking = false;
    }
  }
  checkToken();

  function validate() {
    if (novaSenha.length < minLength) return `A senha deve ter pelo menos ${minLength} caracteres.`;
    if (novaSenha !== confirmacao) return 'A confirmação não confere com a nova senha.';
    return '';
  }

  async function handleSubmit() {
    error = '';
    const v = validate();
    if (v) {
      error = v;
      return;
    }
    submitting = true;
    try {
      await api.publico.redefinirSenha.redefinir(token, novaSenha);
      ok = true;
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<main class="redefinir-senha">
  <PublicShell backLabel="Início" backHref="/">
    <div class="card entry-card">
      {#if checking}
        <p class="text-muted">Verificando link…</p>
      {:else if ok}
        <div class="card-head">
          <h1>Senha redefinida</h1>
          <p class="card-sub">Sua senha foi alterada com sucesso.</p>
        </div>
        <Button block on:click={() => navigate('/entrar')}>Ir para o login</Button>
      {:else if !valid}
        <div class="card-head">
          <h1>Link inválido</h1>
          <p class="card-sub">
            Este link de redefinição de senha é inválido ou já expirou. Peça para gerarem um novo.
          </p>
        </div>
        <Button variant="secondary" block on:click={() => navigate('/entrar')}>Voltar ao login</Button>
      {:else}
        <div class="card-head">
          <h1>Definir nova senha</h1>
          <p class="card-sub">Escolha uma nova senha para a sua conta.</p>
        </div>
        <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
          <Input
            label="Nova senha"
            type="password"
            bind:value={novaSenha}
            placeholder="••••••••"
            autocomplete="new-password"
            required
          />
          <Input
            label="Confirmar nova senha"
            type="password"
            bind:value={confirmacao}
            placeholder="••••••••"
            autocomplete="new-password"
            required
          />
          {#if error}
            <p class="form-error">{error}</p>
          {/if}
          <Button type="submit" block disabled={submitting}>
            {submitting ? 'Salvando…' : 'Salvar nova senha'}
          </Button>
        </form>
      {/if}
    </div>
  </PublicShell>
</main>

<style>
  .redefinir-senha {
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
    margin-bottom: 16px;
  }

  .card-sub {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.875rem;
  }
</style>
