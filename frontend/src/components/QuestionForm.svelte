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
  export let initialType = 'GROUP';
  export let typeEditable = true;
  export let heading = 'Adicionar pergunta';
  export let submitLabel = 'Adicionar';
  export let error = '';
  export let submitting = false;
  export let showCancel = false;

  let title = initialTitle;
  let options = [...initialOptions];
  let type = initialType;

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
    dispatch('submit', { title, options: type === 'OPEN_TEXT' ? [] : options, type });
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
    <span class="label">Tipo de resposta</span>
    <div class="type-toggle">
      <label class="type-option type-group" class:selected={type === 'GROUP'}>
        <input type="radio" class="sr-only" bind:group={type} value="GROUP" disabled={!typeEditable} />
        <span class="type-dot" aria-hidden="true"></span>
        <span class="type-text">Múltipla escolha</span>
      </label>
      <label class="type-option type-open" class:selected={type === 'OPEN_TEXT'}>
        <input type="radio" class="sr-only" bind:group={type} value="OPEN_TEXT" disabled={!typeEditable} />
        <span class="type-dot" aria-hidden="true"></span>
        <span class="type-text">Resposta aberta</span>
      </label>
    </div>
  </div>

  {#if type === 'OPEN_TEXT'}
    <p class="text-muted open-text-note">
      Os participantes vão digitar uma resposta livre em texto, sem opções fixas.
    </p>
  {:else}
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
  {/if}

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
    width: 100%;
    box-sizing: border-box;
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

  .type-toggle {
    display: flex;
    gap: 8px;
  }

  .type-option {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border: 1px solid var(--border);
    border-radius: 8px;
    font-size: 0.85rem;
    line-height: 1.2;
    cursor: pointer;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .type-option:has(input:disabled) {
    cursor: default;
  }

  .type-option:has(input:focus-visible) {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .type-text {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .type-dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .type-group .type-dot {
    background: var(--accent);
  }

  .type-open .type-dot {
    background: var(--open-text-color, #f5a623);
  }

  .type-group.selected {
    border-color: var(--accent);
    background: rgba(79, 140, 255, 0.12);
  }

  .type-open.selected {
    border-color: var(--open-text-color, #f5a623);
    background: rgba(245, 166, 35, 0.12);
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .open-text-note {
    margin: -6px 0 0;
    font-size: 0.85rem;
  }
</style>
