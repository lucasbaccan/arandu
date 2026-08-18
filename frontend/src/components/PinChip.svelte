<script>
  /*
   * PIN em monoespaçada com botão copiar — a mesma peça no Dashboard, no
   * Editar evento, no Palco e no telão da Plateia.
   */
  import CopyButton from './CopyButton.svelte';

  export let pin = '';
  // 'inline' = dentro da linha da tabela; 'boxed' = pílula na faixa de
  // breadcrumb; 'stage' = versão grande do telão.
  export let variant = 'inline';
  export let copyable = true;

  $: value = (pin || '').toUpperCase();
</script>

<span class="pin-chip pin-{variant}">
  {#if variant === 'stage'}
    <span class="pin-caption">Entre com</span>
  {/if}
  <span class="mono-pin pin-value">{value}</span>
  {#if copyable}
    <CopyButton text={value} label="Copiar PIN" />
  {/if}
</span>

<style>
  .pin-chip {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .pin-boxed {
    padding: 5px 8px 5px 10px;
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: var(--surface-muted);
  }

  .pin-value {
    color: var(--text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pin-stage {
    gap: 10px;
    padding: 8px 16px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--surface-muted);
  }

  .pin-caption {
    font-size: 0.75rem;
    font-weight: 800;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--text-subtle);
  }

  .pin-stage .pin-value {
    font-size: 1.25rem;
    letter-spacing: 0.12em;
  }
</style>
