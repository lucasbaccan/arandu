<script context="module">
  export const MIN_OPTIONS = 2;
  export const MAX_OPTIONS = 10;
</script>

<script>
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';
  import Input from './Input.svelte';

  export let initialTitle = '';
  export let initialOptions = ['', ''];
  export let heading = 'Adicionar pergunta';
  export let submitLabel = 'Adicionar';
  export let error = '';
  export let submitting = false;
  export let showCancel = false;

  let title = initialTitle;
  let options = [...initialOptions];

  const dispatch = createEventDispatcher();

  function addOption() {
    if (options.length >= MAX_OPTIONS) return;
    options = [...options, ''];
  }

  function removeOption(index) {
    if (options.length <= MIN_OPTIONS) return;
    options = options.filter((_, i) => i !== index);
  }

  function handleSubmit() {
    dispatch('submit', { title, options });
  }

  function handleCancel() {
    dispatch('cancel');
  }
</script>

<form class="form question-form" novalidate on:submit|preventDefault={handleSubmit}>
  <h3>{heading}</h3>
  <Input
    label="Pergunta"
    bind:value={title}
    placeholder="Ex: Qual é o seu prato favorito?"
    hint="Máximo de 300 caracteres"
  />

  <div class="field">
    <span class="label">Opções</span>
    {#each options as _, i (i)}
      <div class="option-row">
        <Input bind:value={options[i]} placeholder={`Opção ${i + 1}`} autocomplete="off" />
        {#if options.length > MIN_OPTIONS}
          <button
            type="button"
            class="icon-btn"
            title="Remover opção"
            aria-label={`Remover opção ${i + 1}`}
            on:click={() => removeOption(i)}
          >×</button>
        {/if}
      </div>
    {/each}
    {#if options.length < MAX_OPTIONS}
      <Button variant="secondary" type="button" on:click={addOption}>+ Adicionar opção</Button>
    {/if}
  </div>

  {#if error}
    <p class="form-error">{error}</p>
  {/if}

  <div class="form-actions">
    <Button type="submit" disabled={submitting}>
      {submitting ? 'Salvando…' : submitLabel}
    </Button>
    {#if showCancel}
      <Button variant="secondary" type="button" on:click={handleCancel} disabled={submitting}>
        Cancelar
      </Button>
    {/if}
  </div>
</form>

<style>
  .question-form {
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 14px;
  }

  .question-form h3 {
    margin: 0;
    font-size: 1rem;
  }

  .option-row {
    display: flex;
    align-items: flex-start;
    gap: 8px;
  }

  .option-row :global(.field) {
    flex: 1;
  }

  .form-actions {
    display: flex;
    gap: 10px;
  }
</style>
