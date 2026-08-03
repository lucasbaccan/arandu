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
</script>

<div class="field">
  <label>
    {#if label}
      <span class="label">{label}</span>
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
</div>
