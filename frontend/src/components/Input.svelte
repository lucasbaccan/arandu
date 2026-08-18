<script>
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
</style>
