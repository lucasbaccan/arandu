<script>
  export let title = '';
  export let subtitle = '';
  export let clickable = false;
  export let wide = false;
</script>

<div class="card" class:clickable class:wide on:click>
  {#if title}
    <h1>{title}</h1>
  {/if}
  {#if subtitle}
    <p class="subtitle">{subtitle}</p>
  {/if}
  <slot />
</div>

<style>
  .clickable {
    cursor: pointer;
    transition: border-color 0.15s ease, transform 0.1s ease;
  }

  .clickable:hover {
    border-color: var(--accent);
  }

  .clickable:active {
    transform: scale(0.995);
  }
</style>
