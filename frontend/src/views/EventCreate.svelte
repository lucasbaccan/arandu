<script>
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { showToast } from '../lib/toastStore.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Input from '../components/Input.svelte';

  let title = '';
  let customPin = false;
  let pinCode = '';
  let error = '';
  let submitting = false;

  function validate() {
    if (!title.trim()) return 'Informe o título do evento.';
    if (customPin && !/^[a-zA-Z0-9_-]{1,25}$/.test(pinCode.trim())) {
      return 'O PIN deve ter 1 a 25 caracteres: letras, números, _ ou -.';
    }
    return '';
  }

  async function handleSubmit() {
    error = validate();
    if (error) return;
    submitting = true;
    try {
      await api.events.create({
        title: title.trim(),
        pinCode: customPin ? pinCode.trim() : ''
      });
      showToast('Evento criado com sucesso!');
      navigate('/dashboard');
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<main class="page">
  <Card title="Novo evento" subtitle="Configure o acesso dos participantes.">
    <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
      <Input
        label="Título"
        bind:value={title}
        placeholder="Ex: Conecta DevOps 2026"
        autocomplete="off"
        required
      />
      <label class="field check">
        <input type="checkbox" bind:checked={customPin} />
        <span>Definir PIN personalizado</span>
      </label>
      {#if customPin}
        <Input
          label="PIN"
          bind:value={pinCode}
          placeholder="Ex: dev-team"
          hint="1 a 25 caracteres: letras, números, _ ou -"
          uppercase
        />
      {:else}
        <p class="text-muted pin-note">
          Será gerado automaticamente um PIN numérico de 6 dígitos, compartilhável por link ou QR Code.
        </p>
      {/if}
      {#if error}
        <p class="form-error">{error}</p>
      {/if}
      <Button type="submit" variant="accent-invert" block disabled={submitting}>
        {submitting ? 'Criando…' : 'Criar evento'}
      </Button>
    </form>
    <p class="switch">
      <a href="/dashboard" on:click|preventDefault={() => navigate('/dashboard')}>Voltar</a>
    </p>
  </Card>
</main>
