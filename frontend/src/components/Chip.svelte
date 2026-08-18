<script>
  /*
   * Chip único do sistema: fundo tonal + texto colorido, tamanho fixo. Status
   * de evento, tipo de pergunta e opção de resposta são variantes desta peça.
   */
  export let label = '';
  export let tint = 'var(--tint-neutral)';
  export let color = 'var(--text-muted)';
  // 'pill' = status (raio 999, com ponto opcional); 'square' = rótulo de tipo.
  export let shape = 'pill';
  export let dot = false;
</script>

<span class="chip" class:chip-square={shape === 'square'} style="background:{tint};color:{color}">
  {#if dot}<span class="chip-dot" aria-hidden="true">●</span>{/if}
  <slot>{label}</slot>
</span>

<style>
  .chip-dot {
    font-size: 0.5rem;
    line-height: 1;
  }
</style>
