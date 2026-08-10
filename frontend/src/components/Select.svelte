<script>
  import { tick, createEventDispatcher } from 'svelte';

  export let label = '';
  export let value = '';
  export let options = []; // [{ value, label }]
  export let placeholder = 'Selecione';
  export let error = '';
  export let hint = '';

  const dispatch = createEventDispatcher();

  let open = false;
  let triggerEl;
  let listEl;
  let activeIndex = -1;

  $: selected = options.find((o) => o.value === value);

  function toggle() {
    if (open) close();
    else openList();
  }

  async function openList() {
    open = true;
    activeIndex = options.findIndex((o) => o.value === value);
    await tick();
    listEl?.focus();
  }

  function close() {
    open = false;
    triggerEl?.focus();
  }

  function choose(opt) {
    value = opt.value;
    dispatch('change', opt.value);
    close();
  }

  function onTriggerKeydown(e) {
    if (e.key === 'ArrowDown' || e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      openList();
    }
  }

  function onListKeydown(e) {
    if (e.key === 'Escape') {
      e.preventDefault();
      close();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      activeIndex = Math.min(activeIndex + 1, options.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      activeIndex = Math.max(activeIndex - 1, 0);
    } else if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      if (options[activeIndex]) choose(options[activeIndex]);
    } else if (e.key === 'Tab') {
      close();
    }
  }

  function handleWindowMousedown(e) {
    if (!open) return;
    if (triggerEl?.contains(e.target) || listEl?.contains(e.target)) return;
    close();
  }
</script>

<svelte:window on:mousedown={handleWindowMousedown} />

<div class="field">
  <label>
    {#if label}
      <span class="label">{label}</span>
    {/if}
    <div class="dropdown" class:open>
      <button
        type="button"
        class="dropdown-trigger"
        class:invalid={!!error}
        bind:this={triggerEl}
        aria-haspopup="listbox"
        aria-expanded={open}
        on:click={toggle}
        on:keydown={onTriggerKeydown}
      >
        <span class="dropdown-value" class:placeholder={!selected}>
          {selected ? selected.label : placeholder}
        </span>
        <span class="dropdown-caret" aria-hidden="true">▾</span>
      </button>
      {#if open}
        <ul
          class="dropdown-list"
          role="listbox"
          tabindex="-1"
          bind:this={listEl}
          on:keydown={onListKeydown}
        >
          {#each options as opt, i (opt.value)}
            <!-- svelte-ignore a11y-click-events-have-key-events -->
            <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
            <li
              role="option"
              aria-selected={opt.value === value}
              class="dropdown-option"
              class:active={i === activeIndex}
              class:selected={opt.value === value}
              on:mouseenter={() => (activeIndex = i)}
              on:click={() => choose(opt)}
            >
              {opt.label}
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </label>
  {#if error}
    <span class="error">{error}</span>
  {:else if hint}
    <span class="hint">{hint}</span>
  {/if}
</div>

<style>
  .dropdown {
    position: relative;
  }

  .dropdown-trigger {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    background: var(--bg-input);
    border: 1px solid var(--border-strong);
    border-radius: 8px;
    color: var(--text);
    padding: 10px 12px;
    font-size: 1rem;
    font-family: inherit;
    cursor: pointer;
    outline: none;
    transition: border-color 0.15s ease;
  }

  .dropdown-trigger:focus-visible,
  .dropdown.open .dropdown-trigger {
    border-color: var(--accent);
  }

  .dropdown-trigger.invalid {
    border-color: var(--danger);
  }

  .dropdown-value {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .dropdown-value.placeholder {
    color: var(--text-muted);
  }

  .dropdown-caret {
    flex-shrink: 0;
    color: var(--text-muted);
    font-size: 0.7rem;
    transition: transform 0.15s ease;
  }

  .dropdown.open .dropdown-caret {
    transform: rotate(180deg);
  }

  .dropdown-list {
    position: absolute;
    z-index: 20;
    top: calc(100% + 6px);
    left: 0;
    right: 0;
    margin: 0;
    padding: 6px;
    list-style: none;
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
    border-radius: 10px;
    box-shadow: var(--shadow);
    max-height: 240px;
    overflow-y: auto;
    outline: none;
  }

  .dropdown-option {
    padding: 9px 10px;
    border-radius: 6px;
    font-size: 0.95rem;
    color: var(--text);
    cursor: pointer;
  }

  .dropdown-option.active {
    background: var(--bg-input);
  }

  .dropdown-option.selected {
    color: var(--accent);
    font-weight: 700;
  }
</style>
