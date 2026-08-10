<script>
  import { fly, fade } from 'svelte/transition';
  import { reactions } from '../lib/reactionStore.js';
</script>

<div class="reaction-layer" aria-hidden="true">
  {#each $reactions as r (r.id)}
    <span
      class="reaction-emoji"
      style="left: {10 + Math.random() * 80}%;"
      in:fly={{ y: 40, duration: 300 }}
      out:fade={{ duration: 400 }}
    >
      {r.emoji}
    </span>
  {/each}
</div>

<style>
  .reaction-layer {
    position: fixed;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
    z-index: 90;
  }

  .reaction-emoji {
    position: absolute;
    bottom: 10%;
    font-size: 2rem;
    animation: reaction-rise 2s ease-out forwards;
  }

  @keyframes reaction-rise {
    from {
      transform: translateY(0);
      opacity: 1;
    }
    to {
      transform: translateY(-60vh);
      opacity: 0;
    }
  }
</style>
