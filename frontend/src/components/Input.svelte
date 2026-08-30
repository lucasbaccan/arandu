<script>
  import { onMount } from 'svelte';

  export let label = '';
  export let type = 'text';
  export let value = '';
  export let error = '';
  export let placeholder = '';
  export let autocomplete = undefined;
  export let required = false;
  export let hint = '';
  export let uppercase = false;
  // Conteúdo à direita do rótulo (ex: "Esqueci a senha" no Login).
  export let labelAside = '';
  export let labelAsideHref = '';
  // Renderiza um <textarea> auto-expansível (cresce até 4 linhas, depois rola).
  export let textarea = false;

  let taEl;

  function syncHeight() {
    if (!taEl) return;
    taEl.style.height = 'auto';
    taEl.style.height = taEl.scrollHeight + 'px';
  }

  onMount(syncHeight);
</script>

<div class="field">
  <label>
    {#if label}
      <span class="label label-row">
        <span>{label}</span>
        {#if labelAside}
          <a class="label-aside" href={labelAsideHref} on:click>{labelAside}</a>
        {/if}
      </span>
    {/if}
    {#if textarea}
      <textarea
        bind:this={taEl}
        {value}
        on:input={(e) => {
          value = e.currentTarget.value;
          syncHeight();
        }}
        {placeholder}
        {required}
        class:invalid={!!error}
        class:uppercase
        class="ta-auto"
      ></textarea>
    {:else}
      <input
        {type}
        {value}
        on:input={(e) => (value = e.currentTarget.value)}
        {placeholder}
        {autocomplete}
        {required}
        class:invalid={!!error}
        class:uppercase
      />
    {/if}
  </label>
  {#if error}
    <span class="error">{error}</span>
  {:else if hint}
    <span class="hint">{hint}</span>
  {/if}
  <slot name="below" />
</div>

<style>
  .label-row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 10px;
  }

  .label-aside {
    font-size: 0.75rem;
    font-weight: 600;
  }

  /* Auto-expansível: altura acompanha o conteúdo até 4 linhas (1.45 * 4em +
     24px de padding), depois vira rolagem interna. */
  .ta-auto {
    box-sizing: border-box;
    resize: none;
    overflow-y: auto;
    line-height: 1.45;
    min-height: calc(1.45em + 24px);
    max-height: calc(1.45em * 4 + 24px);
  }
</style>
