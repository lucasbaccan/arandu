<script>
  export let title = '';
  export let subtitle = '';
</script>

<div class="card">
  {#if title}
    <h1>{title}</h1>
  {/if}
  {#if subtitle}
    <p class="subtitle">{subtitle}</p>
  {/if}
  <slot />
</div>
