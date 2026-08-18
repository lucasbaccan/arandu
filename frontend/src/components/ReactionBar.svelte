<script>
  import { createEventDispatcher } from 'svelte';

  // Mesmo conjunto do backend (allowedReactionEmojis em internal/api/live.go)
  // — mudar um lado exige mudar o outro.
  export let emojis = ['👍', '❤️', '😂', '🎉', '👏'];
  export let disabled = false;

  const dispatch = createEventDispatcher();

  function react(emoji) {
    if (disabled) return;
    dispatch('react', emoji);
  }
</script>

<div class="reaction-bar">
  {#each emojis as emoji (emoji)}
    <button
      type="button"
      class="reaction-btn"
      {disabled}
      aria-label={`Reagir com ${emoji}`}
      on:click={() => react(emoji)}
    >
      {emoji}
    </button>
  {/each}
</div>

<style>
  .reaction-bar {
    display: flex;
    justify-content: space-between;
    gap: 8px;
  }

  /* Alvo de 56px: é o que o polegar acerta com o celular na mão. */
  .reaction-btn {
    width: 56px;
    height: 56px;
    flex-shrink: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: var(--surface-muted);
    font-size: 1.625rem;
    line-height: 1;
    cursor: pointer;
    transition: border-color 0.15s ease, background 0.15s ease, transform 0.15s ease;
  }

  .reaction-btn:not(:disabled):hover {
    border-color: var(--accent);
    background: var(--accent-soft);
  }

  .reaction-btn:not(:disabled):active {
    transform: scale(0.9);
  }

  .reaction-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
