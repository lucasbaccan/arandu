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
    gap: 8px;
  }

  .reaction-btn {
    width: 44px;
    height: 44px;
    flex-shrink: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: var(--bg-input);
    font-size: 1.3rem;
    line-height: 1;
    cursor: pointer;
    transition: border-color 0.15s ease, transform 0.1s ease;
  }

  .reaction-btn:not(:disabled):hover {
    border-color: var(--accent);
    transform: scale(1.08);
  }

  .reaction-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
