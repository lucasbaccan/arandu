<script>
  export let eventId = '';
  export let questions = [];

  import { api } from '../lib/api.js';
  import { showToast } from '../lib/toastStore.js';
  import { formatDateTime } from '../lib/formatDate.js';
  import Button from './Button.svelte';
  import CopyButton from './CopyButton.svelte';
  import AvatarCropper from './AvatarCropper.svelte';

  let loading = true;
  let error = '';
  let participantCount = 0;
  let participants = [];
  let expandedId = null;
  let editing = null;
  let editValue = '';
  let saving = false;
  let saveError = '';

  let editingPhotoFor = null;
  let photoDraft = '';
  let savingPhoto = false;
  let photoSaveError = '';

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

  function startEditPhoto(p) {
    expandedId = p.id;
    editingPhotoFor = p.id;
    photoDraft = p.photo;
    photoSaveError = '';
  }

  function cancelEditPhoto() {
    editingPhotoFor = null;
    photoSaveError = '';
  }

  function onPhotoDraftChange(e) {
    photoDraft = e.detail;
  }

  async function savePhoto(p) {
    savingPhoto = true;
    photoSaveError = '';
    try {
      await api.events.responses.updatePhoto(eventId, p.id, { photo: photoDraft });
      showToast('Foto atualizada!');
      editingPhotoFor = null;
      await load();
    } catch (e) {
      photoSaveError = e.message;
    } finally {
      savingPhoto = false;
    }
  }

  function editLink(p) {
    return `${window.location.origin}/answer/${eventId}?edit=${p.editToken}`;
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
            <div class="participant-head">
              <button
                type="button"
                class="participant-head-main"
                on:click={() => toggleExpand(p.id)}
                aria-expanded={expandedId === p.id}
              >
                <span class="avatar-wrap">
                  {#if p.photo}
                    <img class="avatar" src={p.photo} alt="" />
                  {:else}
                    <span class="avatar avatar-placeholder">{p.email[0].toUpperCase()}</span>
                  {/if}
                </span>
                <span class="participant-info">
                  <strong>{p.email}</strong>
                  <span class="text-muted">{formatDateTime(p.createdAt)}</span>
                </span>
                <span class="chevron" class:open={expandedId === p.id} aria-hidden="true">›</span>
              </button>
              <button
                type="button"
                class="icon-btn photo-edit-btn"
                title="Editar foto"
                aria-label={`Editar foto de ${p.email}`}
                on:click={() => startEditPhoto(p)}
              >✎</button>
            </div>

            {#if expandedId === p.id}
              <div class="answers">
                <div class="edit-link-row">
                  <span class="text-muted">Link de edição:</span>
                  <CopyButton text={editLink(p)} label="Copiar link de edição" />
                </div>

                {#if editingPhotoFor === p.id}
                  <div class="photo-editor">
                    <AvatarCropper on:change={onPhotoDraftChange} />
                    {#if photoSaveError}
                      <p class="form-error">{photoSaveError}</p>
                    {/if}
                    <div class="answer-edit-actions">
                      <Button type="button" disabled={savingPhoto} on:click={() => savePhoto(p)}>
                        {savingPhoto ? 'Salvando…' : 'Salvar foto'}
                      </Button>
                      <Button
                        variant="secondary"
                        type="button"
                        disabled={savingPhoto}
                        on:click={cancelEditPhoto}
                      >
                        Cancelar
                      </Button>
                    </div>
                  </div>
                {/if}

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
    border: 1px solid var(--border-strong);
    border-radius: 10px;
    background: var(--bg-input);
    overflow: hidden;
  }

  .participant-head {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 8px 6px 12px;
  }

  .participant-head-main {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 4px 0;
    background: transparent;
    border: none;
    cursor: pointer;
    text-align: left;
    color: var(--text);
    transition: background-color 0.15s ease;
  }

  .participant-head-main:hover {
    background: rgba(23, 21, 42, 0.05);
  }

  .avatar-wrap {
    flex-shrink: 0;
  }

  .photo-edit-btn {
    flex-shrink: 0;
    color: var(--orange);
    border-color: rgba(255, 117, 0, 0.35);
    transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease;
  }

  .photo-edit-btn:hover:not(:disabled) {
    background: var(--orange);
    color: var(--text);
    border-color: var(--orange);
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

  .edit-link-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.85rem;
  }

  .photo-editor {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
    border-radius: 10px;
  }

  .answer-row {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding-top: 10px;
    border-top: 1px solid var(--border-strong);
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
    border: 1px solid var(--border-strong);
    border-radius: 8px;
    color: var(--text);
    padding: 8px 10px;
    font-size: 0.9rem;
    font-family: inherit;
    resize: vertical;
  }

  .answer-edit-select {
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
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
