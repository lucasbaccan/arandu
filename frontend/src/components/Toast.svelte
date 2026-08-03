<script>
  import { toast } from '../lib/toastStore.js';

  const icons = { success: '✓', error: '✕', info: 'ℹ' };
</script>

{#if $toast}
  <div class="toast toast-{$toast.type}" role="status">
    <span class="toast-icon">{icons[$toast.type] || ''}</span>
    <span>{$toast.message}</span>
  </div>
{/if}

<style>
  .toast {
    position: fixed;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 100;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 12px 20px;
    box-shadow: var(--shadow);
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.95rem;
    max-width: 90vw;
    animation: toast-in 0.25s ease;
  }

  .toast-success {
    border-color: var(--success);
    color: var(--success);
  }

  .toast-error {
    border-color: var(--danger);
    color: var(--danger);
  }

  .toast-info {
    color: var(--text);
  }

  .toast-icon {
    font-weight: 700;
  }

  @keyframes toast-in {
    from {
      opacity: 0;
      transform: translateX(-50%) translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateX(-50%) translateY(0);
    }
  }
</style>
