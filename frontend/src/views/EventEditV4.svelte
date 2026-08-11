<script>
  export let id = '';

  import { tick } from 'svelte';
  import { flip } from 'svelte/animate';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { showToast } from '../lib/toastStore.js';
  import Button from '../components/Button.svelte';
  import CopyButton from '../components/CopyButton.svelte';
  import Segmented from '../components/Segmented.svelte';
  import QuestionForm, { MIN_OPTIONS, MAX_OPTIONS } from '../components/QuestionForm.svelte';
  import ResponsesPanel from '../components/ResponsesPanel.svelte';

  const statusLabels = {
    PREPARATION: 'Em preparação',
    OPEN_FOR_ANSWERS: 'Coletando respostas',
    CLOSED_FOR_ANSWERS: 'Respostas encerradas',
    PRESENTING: 'Ao vivo',
    FINISHED: 'Finalizado'
  };

  const answerStatusOptions = [
    { value: 'OPEN_FOR_ANSWERS', label: 'Coletando' },
    { value: 'CLOSED_FOR_ANSWERS', label: 'Fechado' }
  ];

  let loading = true;
  let error = '';
  let notFound = false;
  let activeTab = 'questions'; // questions | responses

  $: answerLink = `${window.location.origin}/answer/${id}`;
  $: audienceLink = pinCode
    ? `${window.location.origin}/audience/${id}?pin=${encodeURIComponent(pinCode.toUpperCase())}`
    : `${window.location.origin}/audience/${id}`;

  let title = '';
  let pinCode = '';
  let showRanking = false;
  let status = '';
  let submitting = false;

  let questions = [];
  let questionsLoading = true;
  let participantCount = 0;

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
      const [{ event }, { questions: qs }, { participantCount: pc }] = await Promise.all([
        api.events.get(id),
        api.events.questions.list(id),
        api.events.responses.list(id)
      ]);
      title = event.title;
      pinCode = event.pinCode;
      showRanking = event.configShowRanking;
      status = event.status;
      questions = qs;
      participantCount = pc;
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

  // --- edição inline: título ---
  let editingTitle = false;
  let titleDraft = '';
  let titleSaving = false;
  let titleError = '';
  let titleInputEl;

  async function startEditTitle() {
    titleDraft = title;
    titleError = '';
    editingTitle = true;
    await tick();
    titleInputEl?.focus();
    titleInputEl?.select();
  }

  function cancelEditTitle() {
    editingTitle = false;
    titleError = '';
  }

  async function saveTitle() {
    const trimmed = titleDraft.trim();
    if (!trimmed) {
      titleError = 'Informe o título do evento.';
      return;
    }
    titleSaving = true;
    try {
      await api.events.update(id, { title: trimmed, pinCode: '', configShowRanking: showRanking });
      title = trimmed;
      editingTitle = false;
      showToast('Título atualizado!');
    } catch (e) {
      titleError = e.message;
    } finally {
      titleSaving = false;
    }
  }

  function onTitleKeydown(e) {
    if (e.key === 'Enter') saveTitle();
    if (e.key === 'Escape') cancelEditTitle();
  }

  // --- edição inline: PIN ---
  let editingPin = false;
  let pinDraft = '';
  let pinSaving = false;
  let pinError = '';
  let pinInputEl;

  async function startEditPin() {
    pinDraft = pinCode;
    pinError = '';
    editingPin = true;
    await tick();
    pinInputEl?.focus();
    pinInputEl?.select();
  }

  function cancelEditPin() {
    editingPin = false;
    pinError = '';
  }

  async function savePin() {
    const trimmed = pinDraft.trim();
    if (!/^[a-zA-Z0-9_-]{1,25}$/.test(trimmed)) {
      pinError = 'O PIN deve ter 1 a 25 caracteres: letras, números, _ ou -.';
      return;
    }
    pinSaving = true;
    try {
      const { event } = await api.events.update(id, {
        title: title.trim(),
        pinCode: trimmed,
        configShowRanking: showRanking
      });
      pinCode = event.pinCode;
      editingPin = false;
      showToast('PIN atualizado!');
    } catch (e) {
      pinError = e.message;
    } finally {
      pinSaving = false;
    }
  }

  function onPinKeydown(e) {
    if (e.key === 'Enter') savePin();
    if (e.key === 'Escape') cancelEditPin();
  }

  async function handleSubmit() {
    submitting = true;
    try {
      await api.events.update(id, {
        title: title.trim(),
        pinCode: '',
        configShowRanking: showRanking
      });
      showToast('Alterações salvas!');
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      submitting = false;
    }
  }

  let statusBusy = false;
  let segmentResyncToken = 0;

  async function setAnswerStatus(next) {
    if (next === status || statusBusy) return;
    statusBusy = true;
    try {
      const { event } = await api.events.update(id, {
        title: title.trim(),
        pinCode: '',
        configShowRanking: showRanking,
        status: next
      });
      status = event.status;
      showToast(status === 'OPEN_FOR_ANSWERS' ? 'Respostas abertas!' : 'Respostas encerradas.');
    } catch (e) {
      showToast(e.message, 'error');
      segmentResyncToken += 1; // força o Segmented a re-sincronizar com o status real
    } finally {
      statusBusy = false;
    }
  }

  function validateQuestion(qTitle, qOptions, qType) {
    if (!qTitle.trim()) return 'Informe o texto da pergunta.';
    if (qTitle.trim().length > 300) return 'Pergunta muito longa (máximo 300 caracteres).';
    if (qType === 'OPEN_TEXT') return '';
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
    formError = validateQuestion(detail.title, detail.options, detail.type);
    if (formError) return;
    formBusy = true;
    try {
      const body = {
        title: detail.title.trim(),
        type: detail.type,
        options: detail.type === 'OPEN_TEXT' ? [] : detail.options.map((o) => o.trim()).filter(Boolean)
      };
      const { question } = await api.events.questions.create(id, body);
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
    formError = validateQuestion(detail.title, detail.options, detail.type);
    if (formError) return;
    formBusy = true;
    try {
      const body = {
        title: detail.title.trim(),
        options: detail.type === 'OPEN_TEXT' ? [] : detail.options.map((o) => o.trim()).filter(Boolean)
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
    if (!window.confirm(`Remover a pergunta "${q.title}"? Essa ação não pode ser desfeita.`)) {
      return;
    }
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
    const items = e.currentTarget.querySelectorAll(':scope > li.question-slot');
    let target = dragIndex;
    for (let k = 0; k < items.length; k++) {
      if (k === dragIndex) continue;
      const rect = items[k].getBoundingClientRect();
      const midpoint = rect.top + rect.height / 2;
      if (k > dragIndex && e.clientY >= midpoint) {
        target = Math.max(target, k);
      } else if (k < dragIndex && e.clientY <= midpoint) {
        target = Math.min(target, k);
      }
    }
    dropIndex = target;
  }

  function onDrop(e) {
    if (dragIndex === null) return;
    e.preventDefault();
    const to = dropIndex === null ? dragIndex : dropIndex;
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

  function openOrganizerPanel() {
    navigate(`/stage/${id}`);
  }

  function statusLabel(s) {
    return statusLabels[s] || s;
  }
</script>

<div class="variant-bar">
  <span class="variant-bar-label">Comparar layout:</span>
  <a class="variant-link" href="/events/{id}" on:click|preventDefault={() => navigate(`/events/${id}`)}>Original</a>
  <a class="variant-link" href="/eventos1/{id}" on:click|preventDefault={() => navigate(`/eventos1/${id}`)}>V1 · Status no topo</a>
  <a class="variant-link" href="/eventos2/{id}" on:click|preventDefault={() => navigate(`/eventos2/${id}`)}>V2 · Sidebar limpo</a>
  <a class="variant-link" href="/eventos3/{id}" on:click|preventDefault={() => navigate(`/eventos3/${id}`)}>V3 · Sem status aqui</a>
  <a class="variant-link current" href="/eventos4/{id}" on:click|preventDefault={() => navigate(`/eventos4/${id}`)}>V4 · Segmentado + stats</a>
</div>

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
      <div class="dash-head hero">
        <div class="dash-brand">
          <a
            class="dash-logo-link"
            href="/dashboard"
            aria-label="Voltar para meus eventos"
            on:click|preventDefault={back}
          >
            <img class="dash-logo" src="/img/arandu-completo.png" alt="Arandu" />
          </a>
          <div>
            <div class="title-edit-row">
              {#if editingTitle}
                <input
                  class="title-edit-input"
                  bind:value={titleDraft}
                  bind:this={titleInputEl}
                  on:keydown={onTitleKeydown}
                  disabled={titleSaving}
                />
                <button
                  class="icon-btn inline-edit-btn"
                  title="Salvar título"
                  aria-label="Salvar título"
                  disabled={titleSaving}
                  on:click={saveTitle}
                >✓</button>
                <button
                  class="icon-btn inline-edit-btn"
                  title="Cancelar"
                  aria-label="Cancelar edição do título"
                  disabled={titleSaving}
                  on:click={cancelEditTitle}
                >×</button>
              {:else}
                <h1 class="dash-title">{title || 'Editar evento'}</h1>
                <button
                  class="icon-btn inline-edit-btn"
                  title="Editar título"
                  aria-label="Editar título"
                  on:click={startEditTitle}
                >✎</button>
              {/if}
            </div>
            {#if titleError}<p class="form-error inline-edit-error">{titleError}</p>{/if}

            <p class="dash-user status-row">
              <span class="badge badge-{status.toLowerCase()}">{statusLabel(status)}</span>
              {#if status === 'OPEN_FOR_ANSWERS' || status === 'CLOSED_FOR_ANSWERS'}
                {#key segmentResyncToken}
                  <Segmented
                    options={answerStatusOptions}
                    value={status}
                    disabled={statusBusy}
                    on:change={(e) => setAnswerStatus(e.detail)}
                  />
                {/key}
              {/if}
              <span class="text-muted pin-inline">
                · PIN:
                {#if editingPin}
                  <input
                    class="pin-edit-input"
                    bind:value={pinDraft}
                    bind:this={pinInputEl}
                    on:keydown={onPinKeydown}
                    disabled={pinSaving}
                  />
                  <button
                    class="icon-btn inline-edit-btn"
                    title="Salvar PIN"
                    aria-label="Salvar PIN"
                    disabled={pinSaving}
                    on:click={savePin}
                  >✓</button>
                  <button
                    class="icon-btn inline-edit-btn"
                    title="Cancelar"
                    aria-label="Cancelar edição do PIN"
                    disabled={pinSaving}
                    on:click={cancelEditPin}
                  >×</button>
                {:else}
                  <strong>#{pinCode.toUpperCase()}</strong>
                  <CopyButton text={pinCode.toUpperCase()} label="Copiar PIN" />
                  <button
                    class="icon-btn inline-edit-btn"
                    title="Editar PIN"
                    aria-label="Editar PIN"
                    on:click={startEditPin}
                  >✎</button>
                {/if}
              </span>
              {#if pinError}<span class="form-error inline-edit-error">{pinError}</span>{/if}
              <span class="text-muted pin-inline">
                · <a href={answerLink} target="_blank" rel="noopener">Link de participação</a>
                <CopyButton text={answerLink} label="Copiar link de participação" />
              </span>
              <span class="text-muted pin-inline">
                · <a href={audienceLink} target="_blank" rel="noopener">Apresentação pública</a>
                <CopyButton text={audienceLink} label="Copiar link da apresentação pública" />
              </span>
            </p>
          </div>
        </div>
        <Button
          variant={status === 'PRESENTING' ? 'success-invert' : 'accent-invert'}
          on:click={openOrganizerPanel}
        >
          ▶ Painel do organizador
        </Button>
      </div>

      <div class="stats-row">
        <div class="stat-card stat-purple">
          <span class="stat-value">{questions.length}</span>
          <span class="stat-label">Pergunta{questions.length === 1 ? '' : 's'}</span>
        </div>
        <div class="stat-card stat-cyan">
          <span class="stat-value">{participantCount}</span>
          <span class="stat-label">{participantCount === 1 ? 'Resposta' : 'Respostas'}</span>
        </div>
      </div>

      <div class="edit-layout">
        <div class="card panel settings-panel">
          <h2>Configurações do evento</h2>

          <form class="form" novalidate on:submit|preventDefault={handleSubmit}>
            <label class="field check">
              <input type="checkbox" bind:checked={showRanking} />
              <span>Exibir ranking de pontos</span>
            </label>

            <div class="form-actions">
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Salvando…' : 'Salvar alterações'}
              </Button>
            </div>
          </form>
        </div>

        <div class="card panel questions-panel">
          <div class="tabs" role="tablist">
            <button
              type="button"
              role="tab"
              class="tab"
              class:active={activeTab === 'questions'}
              aria-selected={activeTab === 'questions'}
              on:click={() => (activeTab = 'questions')}
            >
              Perguntas
            </button>
            <button
              type="button"
              role="tab"
              class="tab"
              class:active={activeTab === 'responses'}
              aria-selected={activeTab === 'responses'}
              on:click={() => (activeTab = 'responses')}
            >
              Respostas
            </button>
          </div>

          {#if activeTab === 'responses'}
            <ResponsesPanel eventId={id} {questions} />
          {:else if questionsLoading}
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
              <li class="insert-slot" style="order: 0" class:active={insertAt === 0}>
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
                <li class="question-slot" style="order: {2 * i + 1}" animate:flip={{ duration: 220 }}>
                  {#if editingId === q.id}
                    <div class="insert-slot active">
                      <QuestionForm
                        heading="Editar pergunta"
                        submitLabel="Salvar pergunta"
                        showCancel
                        initialTitle={q.title}
                        initialOptions={q.options.map((o) => o.text)}
                        initialType={q.type}
                        typeEditable={false}
                        error={formError}
                        submitting={formBusy}
                        on:submit={(e) => handleUpdate(e.detail, q.id)}
                        on:cancel={closeForms}
                      />
                    </div>
                  {:else}
                    <div
                      class="question-item"
                      role="listitem"
                      class:dragging={dragIndex === i}
                      class:drop-top={dragIndex !== null && dropIndex === i && dragIndex !== i}
                      draggable={editingId === null && insertAt === null}
                      on:dragstart={(e) => onDragStart(e, i)}
                      on:dragend={onDragEnd}
                    >
                      <div class="reorder-controls">
                        <div class="move-buttons">
                          <button
                            type="button"
                            class="icon-btn icon-btn-move"
                            title="Mover para cima"
                            aria-label="Mover pergunta para cima"
                            disabled={i === 0}
                            on:click={() => moveQuestion(i, i - 1)}
                          >↑</button>
                          <button
                            type="button"
                            class="icon-btn icon-btn-move"
                            title="Mover para baixo"
                            aria-label="Mover pergunta para baixo"
                            disabled={i === questions.length - 1}
                            on:click={() => moveQuestion(i, i + 1)}
                          >↓</button>
                        </div>
                        <span class="drag-handle" title="Arraste para reordenar">⠿</span>
                      </div>
                      <div class="question-info">
                        <div class="question-title-row">
                          <strong class="question-title">{q.title}</strong>
                          {#if q.type !== 'OPEN_TEXT' && q.type !== 'GROUP'}
                            <span class="badge badge-individual">Individual</span>
                          {/if}
                        </div>
                        {#if q.type === 'OPEN_TEXT'}
                          <span class="badge badge-open-text">Resposta aberta</span>
                        {/if}
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
                          class="icon-btn icon-btn-edit"
                          title="Editar pergunta"
                          aria-label={`Editar pergunta ${q.title}`}
                          on:click={() => startEdit(q)}
                        >✎</button>
                        <button
                          type="button"
                          class="icon-btn icon-btn-delete"
                          title="Remover pergunta"
                          aria-label={`Remover pergunta ${q.title}`}
                          on:click={() => removeQuestion(q)}
                        >×</button>
                      </div>
                    </div>
                  {/if}
                </li>
              {/each}

              {#each questions as q, i (q.id)}
                {#if i < questions.length - 1}
                  <li
                    class="insert-slot"
                    style="order: {2 * (i + 1)}"
                    class:active={insertAt === i + 1}
                  >
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
                {/if}
              {/each}

              <li
                class="insert-slot insert-slot-end"
                style="order: {2 * questions.length}"
                class:active={insertAt === questions.length}
              >
                {#if insertAt === questions.length}
                  <QuestionForm
                    heading="Adicionar pergunta"
                    submitLabel="Adicionar"
                    showCancel
                    error={formError}
                    submitting={formBusy}
                    on:submit={(e) => handleInsert(e.detail, questions.length)}
                    on:cancel={closeForms}
                  />
                {:else}
                  <Button
                    variant="secondary"
                    type="button"
                    on:click={() => openInsertAt(questions.length)}
                  >
                    + Adicionar pergunta
                  </Button>
                {/if}
              </li>
            </ul>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  .variant-bar {
    position: sticky;
    top: 0;
    z-index: 40;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    padding: 10px 24px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border-strong);
  }

  .variant-bar-label {
    font-size: 0.8rem;
    color: var(--text-muted);
    font-weight: 600;
    margin-right: 4px;
  }

  .variant-link {
    font-size: 0.8rem;
    padding: 5px 10px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    color: var(--text-muted);
  }

  .variant-link:hover {
    color: var(--text);
    border-color: var(--accent);
    text-decoration: none;
  }

  .variant-link.current {
    background: var(--accent);
    border-color: var(--accent);
    color: #fff;
    font-weight: 600;
  }

  .dash-logo-link {
    display: block;
    line-height: 0;
    opacity: 1;
    transition: opacity 0.15s ease;
  }

  .dash-logo-link:hover {
    opacity: 0.8;
  }

  .dash-head.hero {
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    padding: 20px 24px;
  }

  .stats-row {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .stat-card {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 12px 18px;
    border-radius: 12px;
    min-width: 140px;
  }

  .stat-value {
    font-size: 1.6rem;
    font-weight: 800;
    line-height: 1;
  }

  .stat-label {
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  .stat-purple {
    background: var(--tint-purple);
  }

  .stat-cyan {
    background: var(--tint-cyan);
  }

  .title-edit-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .title-edit-input {
    font-size: 1.6rem;
    font-weight: 700;
    font-family: inherit;
    background: var(--bg-input);
    border: 1px solid var(--border-strong);
    border-radius: 8px;
    color: var(--text);
    padding: 4px 10px;
    outline: none;
    min-width: 220px;
  }

  .title-edit-input:focus {
    border-color: var(--accent);
  }

  .pin-edit-input {
    font-size: 0.85rem;
    font-family: inherit;
    font-weight: 700;
    text-transform: uppercase;
    background: var(--bg-input);
    border: 1px solid var(--border-strong);
    border-radius: 6px;
    color: var(--text);
    padding: 2px 6px;
    outline: none;
    width: 120px;
  }

  .pin-edit-input:focus {
    border-color: var(--accent);
  }

  .inline-edit-error {
    margin: 2px 0 0;
    font-size: 0.78rem;
  }

  .icon-btn.inline-edit-btn {
    width: 24px;
    height: 24px;
    font-size: 0.8rem;
    border: none;
    background: transparent;
    flex-shrink: 0;
  }

  .icon-btn.inline-edit-btn:hover:not(:disabled) {
    background: var(--bg-input);
    color: var(--text);
  }

  .status-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

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

  .tabs {
    display: flex;
    gap: 4px;
    border-bottom: 1px solid var(--border);
  }

  .tab {
    padding: 10px 16px;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-muted);
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
    transition: color 0.15s ease, border-color 0.15s ease;
  }

  .tab:hover {
    color: var(--text);
  }

  .tab.active {
    color: var(--accent);
    border-bottom-color: var(--accent);
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

  .insert-slot {
    display: flex;
    flex-direction: column;
    align-items: stretch;
  }

  .insert-slot.active {
    margin: 6px 0;
  }

  .insert-slot-end {
    margin-top: 8px;
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

  .reorder-controls {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  .move-buttons {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .icon-btn-move {
    font-size: 1.3rem;
    color: var(--accent);
    border-color: rgba(43, 0, 187, 0.35);
    transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease,
      transform 0.15s ease;
  }

  .icon-btn-move:hover:not(:disabled) {
    background: var(--accent);
    color: #fff;
    border-color: var(--accent);
    transform: scale(1.08);
  }

  .icon-btn-edit {
    font-size: 1.3rem;
    color: var(--orange);
    border-color: rgba(255, 117, 0, 0.35);
    transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease,
      transform 0.15s ease;
  }

  .icon-btn-edit:hover:not(:disabled) {
    background: var(--orange);
    color: var(--text);
    border-color: var(--orange);
    transform: scale(1.08);
  }

  .icon-btn-delete {
    font-size: 1.3rem;
    color: var(--danger);
    border-color: rgba(229, 72, 77, 0.35);
    transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease,
      transform 0.15s ease;
  }

  .icon-btn-delete:hover:not(:disabled) {
    background: var(--danger);
    color: #fff;
    border-color: var(--danger);
    transform: scale(1.08);
  }

  .drag-handle {
    color: var(--text-muted);
    font-size: 1.4rem;
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
    background: rgba(104, 103, 122, 0.12);
    flex-shrink: 0;
  }

  .badge-open-text {
    align-self: flex-start;
    color: var(--orange);
    border-color: var(--orange);
    background: rgba(255, 117, 0, 0.12);
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
