<script>
  export let eventId = '';
  export let questions = [];

  import { api } from '../lib/api.js';
  import { showToast } from '../lib/toastStore.js';
  import { formatDate } from '../lib/formatDate.js';
  import Button from './Button.svelte';

  let loading = true;
  let error = '';
  let participantCount = 0;
  let participants = [];
  let expandedId = null;
  let editing = null;
  let editValue = '';
  let saving = false;
  let saveError = '';

  async function load() {
    loading = true;
    try {
      const data = await api.events.responses.list(eventId);
      participantCount = data.participantCount;
      participants = data.participants;
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
  load();

  function toggleExpand(pid) {
    expandedId = expandedId === pid ? null : pid;
  }

  function questionOptions(qid) {
    const q = questions.find((item) => item.id === qid);
    return (q && q.options) || [];
  }

  function startEdit(participantId, answer) {
    editing = { participantId, questionId: answer.questionId };
    editValue = answer.questionType === 'OPEN_TEXT' ? answer.text : answer.optionId;
    saveError = '';
  }

  function cancelEdit() {
    editing = null;
    saveError = '';
  }

  async function saveEdit(answer) {
    saving = true;
    saveError = '';
    try {
      const body =
        answer.questionType === 'OPEN_TEXT' ? { text: editValue.trim() } : { optionId: editValue };
      await api.events.responses.updateAnswer(
        eventId,
        editing.participantId,
        answer.questionId,
        body
      );
      showToast('Resposta atualizada!');
      editing = null;
      await load();
    } catch (e) {
      saveError = e.message;
    } finally {
      saving = false;
    }
  }
</script>

<div class="responses-body">
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if error}
    <p class="form-error">{error}</p>
  {:else}
    <p class="text-muted">
      {participantCount}
      {participantCount === 1 ? 'pessoa respondeu' : 'pessoas responderam'}
    </p>

    {#if participants.length === 0}
      <p class="text-muted empty-note">Ninguém respondeu ainda.</p>
    {:else}
      <ul class="participant-list">
        {#each participants as p (p.id)}
          <li class="participant">
            <button
              type="button"
              class="participant-head"
              on:click={() => toggleExpand(p.id)}
              aria-expanded={expandedId === p.id}
            >
              {#if p.photo}
                <img class="avatar" src={p.photo} alt="" />
              {:else}
                <span class="avatar avatar-placeholder">{p.email[0].toUpperCase()}</span>
              {/if}
              <span class="participant-info">
                <strong>{p.email}</strong>
                <span class="text-muted">{formatDate(p.createdAt)}</span>
              </span>
              <span class="chevron" class:open={expandedId === p.id} aria-hidden="true">›</span>
            </button>

            {#if expandedId === p.id}
              <div class="answers">
                {#each p.answers as a (a.questionId)}
                  <div class="answer-row">
                    <span class="answer-question">{a.questionTitle}</span>

                    {#if editing && editing.participantId === p.id && editing.questionId === a.questionId}
                      {#if a.questionType === 'OPEN_TEXT'}
                        <textarea class="answer-edit-text" bind:value={editValue} rows="2"
                        ></textarea>
                      {:else}
                        <select class="answer-edit-select" bind:value={editValue}>
                          {#each questionOptions(a.questionId) as opt (opt.id)}
                            <option value={opt.id}>{opt.text}</option>
                          {/each}
                        </select>
                      {/if}
                      {#if saveError}
                        <p class="form-error">{saveError}</p>
                      {/if}
                      <div class="answer-edit-actions">
                        <Button type="button" disabled={saving} on:click={() => saveEdit(a)}>
                          {saving ? 'Salvando…' : 'Salvar'}
                        </Button>
                        <Button
                          variant="secondary"
                          type="button"
                          disabled={saving}
                          on:click={cancelEdit}
                        >
                          Cancelar
                        </Button>
                      </div>
                    {:else}
                      <div class="answer-value-row">
                        <span class="answer-value">
                          {a.questionType === 'OPEN_TEXT' ? a.text || '—' : a.optionText}
                        </span>
                        <button
                          type="button"
                          class="icon-btn"
                          title="Editar resposta"
                          aria-label={`Editar resposta de ${p.email} para ${a.questionTitle}`}
                          on:click={() => startEdit(p.id, a)}
                        >✎</button>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  {/if}
</div>

<style>
  .responses-body {
    margin-top: 14px;
  }

  .empty-note {
    margin: 0;
  }

  .participant-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .participant {
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg-input);
    overflow: hidden;
  }

  .participant-head {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    background: transparent;
    border: none;
    cursor: pointer;
    text-align: left;
    color: var(--text);
  }

  .avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;
  }

  .avatar-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--accent);
    color: #fff;
    font-weight: 600;
    font-size: 0.85rem;
  }

  .participant-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
    font-size: 0.9rem;
  }

  .chevron {
    color: var(--text-muted);
    transition: transform 0.15s ease;
    flex-shrink: 0;
  }

  .chevron.open {
    transform: rotate(90deg);
  }

  .answers {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 4px 14px 14px;
  }

  .answer-row {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }

  .answer-question {
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  .answer-value-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }

  .answer-value {
    font-size: 0.95rem;
  }

  .answer-edit-text {
    width: 100%;
    box-sizing: border-box;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text);
    padding: 8px 10px;
    font-size: 0.9rem;
    font-family: inherit;
    resize: vertical;
  }

  .answer-edit-select {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text);
    padding: 8px 10px;
    font-size: 0.9rem;
  }

  .answer-edit-actions {
    display: flex;
    gap: 8px;
  }
</style>
