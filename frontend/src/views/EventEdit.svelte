<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { showToast } from '../lib/toastStore.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import CopyButton from '../components/CopyButton.svelte';
  import Switch from '../components/Switch.svelte';
  import QuestionForm, { MIN_OPTIONS, MAX_OPTIONS } from '../components/QuestionForm.svelte';

  const statusLabels = {
    PREPARATION: 'Em preparação',
    OPEN_FOR_ANSWERS: 'Coletando respostas',
    CLOSED_FOR_ANSWERS: 'Respostas encerradas',
    PRESENTING: 'Ao vivo',
    FINISHED: 'Finalizado'
  };

  let loading = true;
  let error = '';
  let notFound = false;

  $: answerLink = `${window.location.origin}/answer/${id}`;

  let title = '';
  let pinCode = '';
  let changePin = false;
  let showRanking = false;
  let status = '';
  let submitting = false;

  let questions = [];
  let questionsLoading = true;

  let formError = '';
  let formBusy = false;
  let editingId = null;
  let insertAt = null;

  let dragIndex = null;
  let dropIndex = null;
  let reorderBusy = false;

  async function loadQuestions() {
    const { questions: qs } = await api.events.questions.list(id);
    questions = qs;
  }

  async function load() {
    try {
      const [{ event }, { questions: qs }] = await Promise.all([
        api.events.get(id),
        api.events.questions.list(id)
      ]);
      title = event.title;
      pinCode = event.pinCode;
      showRanking = event.configShowRanking;
      status = event.status;
      questions = qs;
    } catch (e) {
      if (e.status === 404) {
        notFound = true;
      } else {
        error = e.message;
      }
    } finally {
      loading = false;
      questionsLoading = false;
    }
  }
  load();

  function validate() {
    if (!title.trim()) return 'Informe o título do evento.';
    if (changePin && !/^[a-zA-Z0-9_-]{1,25}$/.test(pinCode.trim())) {
      return 'O PIN deve ter 1 a 25 caracteres: letras, números, _ ou -.';
    }
    return '';
  }

  async function handleSubmit() {
    error = validate();
    if (error) return;
    submitting = true;
    try {
      await api.events.update(id, {
        title: title.trim(),
        pinCode: changePin ? pinCode.trim() : '',
        configShowRanking: showRanking
      });
      changePin = false;
      showToast('Alterações salvas!');
    } catch (e) {
      showToast(e.message, 'error');
      error = e.message;
    } finally {
      submitting = false;
    }
  }

  let statusBusy = false;

  async function toggleAnswersOpen() {
    const nextStatus = status === 'OPEN_FOR_ANSWERS' ? 'CLOSED_FOR_ANSWERS' : 'OPEN_FOR_ANSWERS';
    statusBusy = true;
    try {
      const { event } = await api.events.update(id, {
        title: title.trim(),
        pinCode: '',
        configShowRanking: showRanking,
        status: nextStatus
      });
      status = event.status;
      showToast(status === 'OPEN_FOR_ANSWERS' ? 'Respostas abertas!' : 'Respostas encerradas.');
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      statusBusy = false;
    }
  }

  function validateQuestion(qTitle, qOptions) {
    if (!qTitle.trim()) return 'Informe o texto da pergunta.';
    if (qTitle.trim().length > 300) return 'Pergunta muito longa (máximo 300 caracteres).';
    const opts = qOptions.map((o) => o.trim()).filter(Boolean);
    if (opts.length < MIN_OPTIONS) return `Informe pelo menos ${MIN_OPTIONS} opções.`;
    if (opts.length > MAX_OPTIONS) return `Máximo de ${MAX_OPTIONS} opções.`;
    if (opts.some((o) => o.length > 120)) return 'Opção muito longa (máximo 120 caracteres).';
    if (new Set(opts).size !== opts.length) return 'As opções devem ser diferentes entre si.';
    return '';
  }

  function startEdit(q) {
    insertAt = null;
    formError = '';
    editingId = q.id;
  }

  function openInsertAt(pos) {
    editingId = null;
    formError = '';
    insertAt = pos;
  }

  function closeForms() {
    editingId = null;
    insertAt = null;
    formError = '';
  }

  async function handleInsert(detail, pos) {
    formError = validateQuestion(detail.title, detail.options);
    if (formError) return;
    formBusy = true;
    try {
      const body = {
        title: detail.title.trim(),
        options: detail.options.map((o) => o.trim()).filter(Boolean)
      };
      const { question } = await api.events.questions.create(id, { ...body, type: 'GROUP' });
      const next = [...questions];
      next.splice(pos, 0, question);
      questions = next;
      if (pos < questions.length - 1) {
        await persistOrder();
      }
      showToast('Pergunta adicionada!');
      closeForms();
    } catch (e) {
      showToast(e.message, 'error');
      formError = e.message;
    } finally {
      formBusy = false;
    }
  }

  async function handleUpdate(detail, questionId) {
    formError = validateQuestion(detail.title, detail.options);
    if (formError) return;
    formBusy = true;
    try {
      const body = {
        title: detail.title.trim(),
        options: detail.options.map((o) => o.trim()).filter(Boolean)
      };
      const { question } = await api.events.questions.update(id, questionId, body);
      questions = questions.map((x) => (x.id === questionId ? question : x));
      showToast('Pergunta atualizada!');
      closeForms();
    } catch (e) {
      showToast(e.message, 'error');
      formError = e.message;
    } finally {
      formBusy = false;
    }
  }

  async function removeQuestion(q) {
    try {
      await api.events.questions.remove(id, q.id);
      questions = questions.filter((x) => x.id !== q.id);
      showToast('Pergunta removida.');
    } catch (e) {
      showToast(e.message, 'error');
    }
  }

  function moveQuestion(from, to) {
    if (to < 0 || to >= questions.length || from === to) return;
    const next = [...questions];
    const [moved] = next.splice(from, 1);
    next.splice(to, 0, moved);
    questions = next;
    persistOrder();
  }

  async function persistOrder() {
    if (reorderBusy || questions.length < 2) return;
    reorderBusy = true;
    try {
      await api.events.questions.reorder(
        id,
        questions.map((q) => q.id)
      );
    } catch (e) {
      showToast(e.message, 'error');
      try {
        await loadQuestions();
      } catch {
        // mantém a ordem atual se a recarga falhar
      }
    } finally {
      reorderBusy = false;
    }
  }

  function onDragStart(e, index) {
    if (editingId !== null || insertAt !== null) return;
    dragIndex = index;
    dropIndex = index;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', String(index));
    }
  }

  function onDragOver(e) {
    if (dragIndex === null) return;
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }
    const items = e.currentTarget.querySelectorAll(':scope > li.question-item');
    let target = items.length;
    for (let k = 0; k < items.length; k++) {
      const rect = items[k].getBoundingClientRect();
      if (e.clientY < rect.top + rect.height / 2) {
        target = k;
        break;
      }
    }
    dropIndex = target;
  }

  function onDrop(e) {
    if (dragIndex === null) return;
    e.preventDefault();
    let to = dropIndex;
    if (to === null) to = questions.length;
    if (dragIndex < to) to -= 1;
    moveQuestion(dragIndex, to);
    dragIndex = null;
    dropIndex = null;
  }

  function onDragEnd() {
    dragIndex = null;
    dropIndex = null;
  }

  function back() {
    navigate('/dashboard');
  }

  function statusLabel(s) {
    return statusLabels[s] || s;
  }
</script>

<main class="page page-wide">
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if notFound}
    <div class="card panel">
      <h1>Evento não encontrado</h1>
      <p class="subtitle">Ele pode ter sido removido ou você não tem acesso.</p>
      <Button block variant="secondary" on:click={back}>Voltar</Button>
    </div>
  {:else}
    <div class="edit-wrap">
      <div class="dash-head">
        <div class="dash-brand">
          <Button variant="secondary" on:click={back}>Voltar</Button>
          <div>
            <h1 class="dash-title">{title || 'Editar evento'}</h1>
            <p class="dash-user">
              <span class="badge badge-{status.toLowerCase()}">{statusLabel(status)}</span>
              <span class="text-muted pin-inline">
                · PIN: <strong>{pinCode.toUpperCase()}</strong>
                <CopyButton text={pinCode.toUpperCase()} label="Copiar PIN" />
              </span>
              <span class="text-muted pin-inline">
                · <a href={answerLink} target="_blank" rel="noopener">Link de participação</a>
                <CopyButton text={answerLink} label="Copiar link de participação" />
              </span>
            </p>
          </div>
        </div>
      </div>

      <div class="edit-layout">
        <div class="card panel settings-panel">
          <h2>Configurações do evento</h2>
          <p class="text-muted questions-count">
            {questions.length} pergunta{questions.length === 1 ? '' : 's'} adicionada{questions.length === 1 ? '' : 's'}
          </p>

          <div class="answers-switch">
            <div>
              <strong>Respostas {status === 'OPEN_FOR_ANSWERS' ? 'abertas' : 'fechadas'}</strong>
              <p class="text-muted">Participantes só respondem enquanto estiver aberto.</p>
            </div>
            <Switch
              checked={status === 'OPEN_FOR_ANSWERS'}
              disabled={statusBusy}
              on:change={toggleAnswersOpen}
            />
          </div>

          <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
            <Input
              label="Título"
              bind:value={title}
              placeholder="Ex: Conecta DevOps 2026"
              autocomplete="off"
              required
            />

            <label class="field check">
              <input type="checkbox" bind:checked={changePin} />
              <span>Definir novo PIN</span>
            </label>
            {#if changePin}
              <Input
                label="Novo PIN"
                bind:value={pinCode}
                placeholder="Ex: dev-team"
                hint="1 a 25 caracteres: letras, números, _ ou -"
                uppercase
              />
            {/if}

            <label class="field check">
              <input type="checkbox" bind:checked={showRanking} />
              <span>Exibir ranking de pontos</span>
            </label>

            {#if error}
              <p class="form-error">{error}</p>
            {/if}

            <div class="form-actions">
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Salvando…' : 'Salvar alterações'}
              </Button>
            </div>
          </form>
        </div>

        <div class="card panel questions-panel">
          <div class="questions-head">
            <h2>Perguntas da dinâmica</h2>
            <p class="text-muted">Arraste para reordenar ou clique em "+" para inserir uma pergunta em qualquer posição.</p>
          </div>

          {#if questionsLoading}
            <p class="text-muted">Carregando perguntas…</p>
          {:else if questions.length === 0}
            <div class="empty-note">
              <p class="text-muted">Nenhuma pergunta ainda.</p>
              {#if insertAt === 0}
                <QuestionForm
                  heading="Adicionar pergunta"
                  submitLabel="Adicionar"
                  showCancel
                  error={formError}
                  submitting={formBusy}
                  on:submit={(e) => handleInsert(e.detail, 0)}
                  on:cancel={closeForms}
                />
              {:else}
                <Button variant="secondary" type="button" on:click={() => openInsertAt(0)}>
                  + Adicionar pergunta
                </Button>
              {/if}
            </div>
          {:else}
            <ul
              class="question-list"
              on:dragover={onDragOver}
              on:drop={onDrop}
            >
              <li class="insert-slot" class:active={insertAt === 0}>
                {#if insertAt === 0}
                  <QuestionForm
                    heading="Adicionar pergunta"
                    submitLabel="Adicionar"
                    showCancel
                    error={formError}
                    submitting={formBusy}
                    on:submit={(e) => handleInsert(e.detail, 0)}
                    on:cancel={closeForms}
                  />
                {:else}
                  <button
                    type="button"
                    class="insert-btn"
                    aria-label="Adicionar pergunta no início"
                    on:click={() => openInsertAt(0)}
                  >+ Adicionar pergunta aqui</button>
                {/if}
              </li>

              {#each questions as q, i (q.id)}
                {#if editingId === q.id}
                  <li class="insert-slot active">
                    <QuestionForm
                      heading="Editar pergunta"
                      submitLabel="Salvar pergunta"
                      showCancel
                      initialTitle={q.title}
                      initialOptions={q.options.map((o) => o.text)}
                      error={formError}
                      submitting={formBusy}
                      on:submit={(e) => handleUpdate(e.detail, q.id)}
                      on:cancel={closeForms}
                    />
                  </li>
                {:else}
                  <li
                    class="question-item"
                    class:dragging={dragIndex === i}
                    class:drop-top={dragIndex !== null && dropIndex === i && dragIndex !== i}
                    draggable={editingId === null && insertAt === null}
                    on:dragstart={(e) => onDragStart(e, i)}
                    on:dragend={onDragEnd}
                  >
                    <span class="drag-handle" title="Arraste para reordenar">⠿</span>
                    <div class="question-info">
                      <div class="question-title-row">
                        <strong class="question-title">{q.title}</strong>
                        {#if q.type !== 'GROUP'}
                          <span class="badge badge-individual">Individual</span>
                        {/if}
                      </div>
                      {#if q.options.length > 0}
                        <ul class="option-chips">
                          {#each q.options as opt (opt.id)}
                            <li class="option-chip">{opt.text}</li>
                          {/each}
                        </ul>
                      {/if}
                    </div>
                    <div class="question-actions">
                      <button
                        type="button"
                        class="icon-btn"
                        title="Mover para cima"
                        aria-label="Mover pergunta para cima"
                        disabled={i === 0}
                        on:click={() => moveQuestion(i, i - 1)}
                      >↑</button>
                      <button
                        type="button"
                        class="icon-btn"
                        title="Mover para baixo"
                        aria-label="Mover pergunta para baixo"
                        disabled={i === questions.length - 1}
                        on:click={() => moveQuestion(i, i + 1)}
                      >↓</button>
                      <button
                        type="button"
                        class="icon-btn"
                        title="Editar pergunta"
                        aria-label={`Editar pergunta ${q.title}`}
                        on:click={() => startEdit(q)}
                      >✎</button>
                      <button
                        type="button"
                        class="icon-btn danger"
                        title="Remover pergunta"
                        aria-label={`Remover pergunta ${q.title}`}
                        on:click={() => removeQuestion(q)}
                      >×</button>
                    </div>
                  </li>
                {/if}

                <li class="insert-slot" class:active={insertAt === i + 1}>
                  {#if insertAt === i + 1}
                    <QuestionForm
                      heading="Adicionar pergunta"
                      submitLabel="Adicionar"
                      showCancel
                      error={formError}
                      submitting={formBusy}
                      on:submit={(e) => handleInsert(e.detail, i + 1)}
                      on:cancel={closeForms}
                    />
                  {:else}
                    <button
                      type="button"
                      class="insert-btn"
                      aria-label={`Adicionar pergunta após "${q.title}"`}
                      on:click={() => openInsertAt(i + 1)}
                    >+ Adicionar pergunta aqui</button>
                  {/if}
                </li>
              {/each}
              {#if dragIndex !== null && dropIndex === questions.length}
                <li class="drop-end" aria-hidden="true"></li>
              {/if}
            </ul>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  .pin-inline {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    vertical-align: middle;
  }

  .edit-wrap {
    width: 100%;
    max-width: 1100px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .edit-layout {
    display: grid;
    grid-template-columns: minmax(280px, 360px) 1fr;
    gap: 16px;
    align-items: start;
  }

  @media (max-width: 880px) {
    .edit-layout {
      grid-template-columns: 1fr;
    }
  }

  .panel {
    max-width: none;
    width: 100%;
  }

  .panel h2 {
    font-size: 1.05rem;
  }

  .form-actions {
    display: flex;
    gap: 10px;
  }

  .questions-panel {
    gap: 16px;
  }

  .questions-head p {
    margin: -8px 0 0;
  }

  .questions-count {
    margin: -10px 0 0;
    font-size: 0.85rem;
  }

  .answers-switch {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 10px;
  }

  .answers-switch strong {
    font-size: 0.9rem;
  }

  .answers-switch p {
    margin: 2px 0 0;
    font-size: 0.8rem;
  }

  .empty-note {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .empty-note p {
    margin: 0;
  }

  .question-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .question-item {
    display: flex;
    align-items: center;
    gap: 12px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 12px 14px;
    cursor: grab;
    transition: border-color 0.15s ease, opacity 0.15s ease, transform 0.15s ease;
  }

  .question-item.dragging {
    opacity: 0.45;
    border-style: dashed;
    border-color: var(--accent);
  }

  .question-item.drop-top {
    transform: translateX(10px);
    border-color: var(--accent);
  }

  .drop-end {
    height: 2px;
    background: var(--accent);
    border-radius: 2px;
    margin-top: -2px;
  }

  .insert-slot {
    display: flex;
  }

  .insert-slot.active {
    margin: 6px 0;
  }

  .insert-btn {
    flex: 1;
    width: 100%;
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 0.75rem;
    cursor: pointer;
    padding: 6px 0;
    opacity: 0;
    transition: opacity 0.15s ease, color 0.15s ease;
  }

  .insert-slot:hover .insert-btn,
  .insert-btn:focus-visible {
    opacity: 1;
  }

  .insert-btn::before {
    content: '';
    display: block;
    height: 1px;
    background: var(--accent);
    margin-bottom: 4px;
  }

  .insert-btn:hover,
  .insert-btn:focus-visible {
    color: var(--accent);
  }

  .drag-handle {
    color: var(--text-muted);
    font-size: 1.2rem;
    line-height: 1;
    flex-shrink: 0;
  }

  .question-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .question-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .question-title {
    font-size: 0.95rem;
    line-height: 1.35;
  }

  .badge-individual {
    color: var(--text-muted);
    border-color: var(--border);
    background: rgba(139, 152, 165, 0.12);
    flex-shrink: 0;
  }

  .option-chips {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .option-chip {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 3px 10px;
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  .question-actions {
    display: flex;
    flex-direction: column;
    gap: 6px;
    flex-shrink: 0;
  }

</style>
