<script>
  // Indicador de perda de conexão com a apresentação ao vivo — as telas que
  // assinam SSE (/palco, /apresentar, /plateia) mostram ele quando o stream
  // para de entregar frames em tempo real (proxy que "silencia" a conexão,
  // rede caindo, etc.), mesmo que o poll de re-sync continue trazendo dados.
  export let text = 'Sem conexão';
</script>

<span class="live-status" role="status" aria-live="polite">
  <span class="live-status-dot" aria-hidden="true"></span>
  {text}
</span>

<style>
  .live-status {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 3px 10px;
    border-radius: 999px;
    background: var(--tint-danger);
    color: var(--danger);
    font-size: 0.75rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .live-status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
    animation: live-status-pulse 1.2s ease-in-out infinite;
  }

  @keyframes live-status-pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.25;
    }
  }
</style>
