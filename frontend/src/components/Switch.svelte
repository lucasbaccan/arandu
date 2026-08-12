<script>
  import { createEventDispatcher } from 'svelte';

  export let checked = false;
  export let disabled = false;

  const dispatch = createEventDispatcher();

  function toggle() {
    if (disabled) return;
    dispatch('change');
  }
</script>

<button
  type="button"
  role="switch"
  aria-checked={checked}
  class="switch"
  class:on={checked}
  disabled={disabled}
  on:click={toggle}
  {...$$restProps}
>
  <span class="knob"></span>
</button>

<style>
  .switch {
    width: 44px;
    height: 24px;
    flex-shrink: 0;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--bg-input);
    position: relative;
    cursor: pointer;
    padding: 0;
    transition: background 0.15s ease, border-color 0.15s ease;
  }

  .switch.on {
    background: var(--accent);
    border-color: var(--accent);
  }

  .switch:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .knob {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #fff;
    transition: transform 0.15s ease;
  }

  .switch.on .knob {
    transform: translateX(20px);
  }
</style>
