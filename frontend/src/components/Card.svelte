<script>
  import { createEventDispatcher } from 'svelte';

  export let title = '';
  export let subtitle = '';
  export let clickable = false;
  export let wide = false;

  const dispatch = createEventDispatcher();

  function handleKeydown(e) {
    if (!clickable) return;
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      dispatch('click', e);
    }
  }
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<div
  class="card"
  class:clickable
  class:wide
  role={clickable ? 'button' : undefined}
  tabindex={clickable ? 0 : undefined}
  on:click
  on:keydown={handleKeydown}
>
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
