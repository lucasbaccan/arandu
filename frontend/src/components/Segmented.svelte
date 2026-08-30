<script>
  import { createEventDispatcher } from 'svelte';

  export let options = []; // [{ value, label }]
  export let value = '';
  export let disabled = false;

  const dispatch = createEventDispatcher();

  function select(v) {
    if (disabled || v === value) return;
    value = v;
    dispatch('change', v);
  }
</script>

<div class="segmented" role="radiogroup">
  {#each options as opt (opt.value)}
    <button
      type="button"
      role="radio"
      aria-checked={value === opt.value}
      class="segmented-option"
      class:active={value === opt.value}
      {disabled}
      on:click={() => select(opt.value)}
    >
      {opt.label}
    </button>
  {/each}
</div>

<style>
  .segmented {
    display: inline-flex;
    background: var(--bg-input);
    border: 1px solid var(--border-strong);
    border-radius: 10px;
    padding: 3px;
    gap: 2px;
  }

  .segmented-option {
    border: none;
    background: transparent;
    color: var(--text-muted);
    padding: 8px 16px;
    font-size: 0.9rem;
    font-weight: 600;
    font-family: inherit;
    border-radius: 7px;
    cursor: pointer;
    transition: background-color 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;
  }

  .segmented-option.active {
    background: var(--bg-elev);
    color: var(--accent);
    box-shadow: 0 1px 4px rgba(23, 21, 42, 0.12);
  }

  .segmented-option:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Alvo de toque (~36px -> ~44px) em telas de toque. */
  @media (pointer: coarse) {
    .segmented-option {
      padding: 12px 16px;
    }
  }
</style>
