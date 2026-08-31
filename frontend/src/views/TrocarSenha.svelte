<script>
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import TopBar from '../components/TopBar.svelte';
  import CrumbBar from '../components/CrumbBar.svelte';

  let senhaAtual = '';
  let novaSenha = '';
  let confirmacao = '';
  let error = '';
  let ok = '';
  let submitting = false;

  function validate() {
    if (!senhaAtual) return 'Informe sua senha atual.';
    if (!novaSenha) return 'Informe a nova senha.';
    if (novaSenha !== confirmacao) return 'A confirmação não confere com a nova senha.';
    return '';
  }

  async function handleSubmit() {
    error = '';
    ok = '';
    const v = validate();
    if (v) {
      error = v;
      return;
    }
    submitting = true;
    try {
      await api.trocarSenha({ senhaAtual, novaSenha });
      ok = 'Senha alterada com sucesso.';
      senhaAtual = novaSenha = confirmacao = '';
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<main class="trocar-senha">
  <TopBar area="Conta" />
  <CrumbBar crumbs={[{ label: 'Conta' }, { label: 'Trocar senha' }]}>
    <svelte:fragment slot="actions">
      <Button variant="ghost" on:click={() => navigate('/painel')}>Voltar</Button>
    </svelte:fragment>
  </CrumbBar>
  <section class="content">
    <div class="card">
      <div class="card-head">
        <h1>Trocar senha</h1>
        <p class="card-sub">Digite sua senha atual e escolha uma nova.</p>
      </div>
      <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
        <Input
          label="Senha atual"
          type="password"
          bind:value={senhaAtual}
          placeholder="••••••••"
          autocomplete="current-password"
          required
        />
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
        {#if ok}
          <p class="form-ok">{ok}</p>
        {/if}
        <Button type="submit" block disabled={submitting}>
          {submitting ? 'Salvando…' : 'Salvar nova senha'}
        </Button>
      </form>
    </div>
  </section>
</main>

<style>
  .trocar-senha {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .content {
    flex: 1;
    display: flex;
    justify-content: center;
    padding: 24px 16px;
  }

  .card {
    width: 100%;
    max-width: 420px;
    align-self: flex-start;
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

  .form-ok {
    margin: 0 0 12px;
    color: var(--success, #16a34a);
    font-weight: 700;
    font-size: 0.875rem;
  }
</style>
