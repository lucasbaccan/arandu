<script>
  import Button from './Button.svelte';
  import Input from './Input.svelte';
  import Switch from './Switch.svelte';

  export let status = '';
  export let statusBusy = false;
  export let questionCount = 0;
  export let participantCount = 0;
  export let title;
  export let pinCode;
  export let showRanking = false;
  export let allowEdit;
  export let error = '';
  export let submitting = false;
  export let toggleAnswersOpen;
  export let handleSubmit;
  // Função opcional de exclusão do evento: sem ela, a zona de perigo não renderiza.
  export let onDeleteEvent = null;
  export let deletingEvent = false;
</script>

<div
  class="rail-card answers-switch"
  style="--rail-accent:{status === 'OPEN_FOR_ANSWERS' ? 'var(--success)' : 'var(--border-strong)'}"
>
  <div>
    <strong>Respostas {status === 'OPEN_FOR_ANSWERS' ? 'abertas' : 'fechadas'}</strong>
    <p class="text-muted">Só entram respostas enquanto estiver ligado.</p>
  </div>
  <Switch
    checked={status === 'OPEN_FOR_ANSWERS'}
    disabled={statusBusy}
    on:change={toggleAnswersOpen}
  />
</div>

<div class="stat-row">
  <div class="rail-card stat-box">
    <span class="stat-num">{questionCount}</span>
    <span class="stat-label">pergunta{questionCount === 1 ? '' : 's'}</span>
  </div>
  <div class="rail-card stat-box">
    <span class="stat-num">{participantCount}</span>
    <span class="stat-label">{participantCount === 1 ? 'respondeu' : 'responderam'}</span>
  </div>
</div>

<div class="rail-card rail-section">
  <span class="overline">Configurações</span>
  <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
    <Input
      label="Título"
      bind:value={title}
      placeholder="Ex: Conecta DevOps 2026"
      autocomplete="off"
      required
    />

    <Input
      label="PIN"
      bind:value={pinCode}
      placeholder="Ex: dev-team"
      hint="1 a 25 caracteres: letras, números, _ ou -"
      uppercase
    />

    <div class="config-row">
      <span>Exibir ranking de pontos</span>
      <Switch
        aria-label="Exibir ranking de pontos"
        checked={showRanking}
        on:change={() => (showRanking = !showRanking)}
      />
    </div>

    <div class="config-row config-row-hint">
      <span>
        Permitir editar depois
        <p class="text-muted">
          {allowEdit
            ? 'Cada pessoa recebe um link privado para corrigir respostas e foto.'
            : 'O envio é único. Só você pode corrigir respostas por aqui.'}
        </p>
      </span>
      <Switch
        aria-label="Permitir editar depois"
        checked={allowEdit}
        on:change={() => (allowEdit = !allowEdit)}
      />
    </div>

    {#if error}
      <p class="form-error">{error}</p>
    {/if}

    <Button type="submit" variant="secondary" block disabled={submitting}>
      {submitting ? 'Salvando…' : 'Salvar alterações'}
    </Button>
  </form>
</div>

{#if onDeleteEvent}
  <div class="rail-card rail-section danger-zone">
    <span class="overline">Cuidado</span>
    <Button variant="danger" block disabled={deletingEvent} on:click={onDeleteEvent}>
      {deletingEvent ? 'Excluindo…' : 'Excluir evento'}
    </Button>
    <p class="text-muted danger-note">
      Apaga o evento inteiro: perguntas, respostas e fotos. Não pode ser desfeito.
    </p>
  </div>
{/if}

<style>
  .rail-card {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    padding: 16px;
  }

  .rail-section {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .answers-switch {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 16px;
    border-left: 3px solid var(--rail-accent);
  }

  .answers-switch strong {
    font-size: 0.875rem;
  }

  .answers-switch p {
    margin: 2px 0 0;
    font-size: 0.75rem;
  }

  .stat-row {
    display: flex;
    gap: 10px;
  }

  .stat-box {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 14px;
  }

  .stat-num {
    font-size: 1.625rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1;
  }

  .stat-label {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .config-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    font-size: 0.875rem;
  }

  .config-row-hint {
    align-items: flex-start;
  }

  .config-row-hint p {
    margin: 2px 0 0;
    font-size: 0.75rem;
    line-height: 1.35;
  }

  .danger-zone {
    border-color: var(--tint-danger);
  }

  .danger-note {
    margin: 0;
    font-size: 0.75rem;
    line-height: 1.35;
  }
</style>
