<script>
  export let id = '';

  import { flip } from 'svelte/animate';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { showToast } from '../lib/toastStore.js';
  import { formatDateTime } from '../lib/formatDate.js';
  import { statusInfo, questionKindInfo } from '../lib/eventStatus.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import CopyButton from '../components/CopyButton.svelte';
  import Switch from '../components/Switch.svelte';
  import Chip from '../components/Chip.svelte';
  import CrumbBar from '../components/CrumbBar.svelte';
  import PinChip from '../components/PinChip.svelte';
  import Tabs from '../components/Tabs.svelte';
  import TopBar from '../components/TopBar.svelte';
  import QuestionForm, { MIN_OPTIONS, MAX_OPTIONS } from '../components/QuestionForm.svelte';
  import AvatarCropper from '../components/AvatarCropper.svelte';

  function formatTime(iso) {
    const d = new Date(iso);
    const hh = String(d.getUTCHours()).padStart(2, '0');
    const min = String(d.getUTCMinutes()).padStart(2, '0');
    return `${hh}:${min}`;
  }

  let loading = true;
  let error = '';
  let notFound = false;
  let activeTab = 'questions'; // questions | responses
  let railCollapsed = false;

  $: answerLink = `${window.location.origin}/answer/${id}`;
  $: audienceLink = pinCode
    ? `${window.location.origin}/audience/${id}?pin=${encodeURIComponent(pinCode.toUpperCase())}`
    : `${window.location.origin}/audience/${id}`;

  let title = '';
  let pinCode = '';
  let showRanking = false;
  let allowEdit = true;
  let status = '';
  let submitting = false;

  let questions = [];
  let questionsLoading = true;
  let participantCount = 0;
  let participants = [];

  let formError = '';
  let formBusy = false;
  let editingId = null;
  let insertAt = null;

  let dragIndex = null;
  let dropIndex = null;
  let reorderBusy = false;

  // Respostas
  let personSearch = '';
  let sortMode = 'sequence'; // sequence | name
  let selectedParticipantId = null;
  let photoDialogOpen = false;
  let photoDraft = '';
  let savingPhoto = false;
  let savingAnswerKey = '';

  // Coluna de pessoas: arrastável, mas nunca menor que o tamanho de base —
  // o valor que já era o padrão da tela antes de existir o redimensionamento.
  const PEOPLE_COL_MIN = 250;
  let peopleColWidth = PEOPLE_COL_MIN;
  let responsesSplitEl;

  function peopleColMax() {
    const containerWidth = responsesSplitEl ? responsesSplitEl.getBoundingClientRect().width : Infinity;
    // Reserva um mínimo pro detail-col não desaparecer atrás da coluna de pessoas.
    return Math.max(PEOPLE_COL_MIN, containerWidth - 220);
  }

  function startPeopleColResize(e) {
    e.preventDefault();
    const startX = e.clientX;
    const startWidth = peopleColWidth;
    const maxWidth = peopleColMax();

    function onMove(ev) {
      const next = startWidth + (ev.clientX - startX);
      peopleColWidth = Math.min(maxWidth, Math.max(PEOPLE_COL_MIN, next));
    }
    function onUp() {
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
    }
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
  }

  function handlePeopleColResizeKeydown(e) {
    if (e.key === 'ArrowLeft') {
      e.preventDefault();
      peopleColWidth = Math.max(PEOPLE_COL_MIN, peopleColWidth - 16);
    } else if (e.key === 'ArrowRight') {
      e.preventDefault();
      peopleColWidth = Math.min(peopleColMax(), peopleColWidth + 16);
    }
  }

  async function loadQuestions() {
    const { questions: qs } = await api.events.questions.list(id);
    questions = qs;
  }

  async function loadResponses() {
    const { participantCount: pc, participants: parts } = await api.events.responses.list(id);
    participantCount = pc;
    participants = parts;
    if (!selectedParticipantId || !parts.some((p) => p.id === selectedParticipantId)) {
      selectedParticipantId = parts[0] ? parts[0].id : null;
    }
  }

  async function load() {
    try {
      const [{ event }, { questions: qs }] = await Promise.all([
        api.events.get(id),
        api.events.questions.list(id),
        loadResponses()
      ]);
      title = event.title;
      pinCode = event.pinCode;
      showRanking = event.configShowRanking;
      allowEdit = event.allowEdit;
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
    if (!/^[a-zA-Z0-9_-]{1,25}$/.test(pinCode.trim())) {
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
        pinCode: pinCode.trim(),
        configShowRanking: showRanking,
        allowEdit
      });
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
        allowEdit,
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
    // Direction-aware: dragging past an item's midpoint swaps with it
    // immediately, instead of requiring an overshoot into the next item.
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

  // Recarrega as respostas sempre que a aba é aberta, para refletir
  // participações novas desde o carregamento inicial da página.
  $: if (activeTab === 'responses' && !loading) {
    loadResponses();
  }

  // --- Respostas ---

  $: sortedPeople = participants
    .map((p, i) => ({ ...p, order: i + 1 }))
    .filter((p) => {
      const t = personSearch.trim().toLowerCase();
      if (!t) return true;
      return (
        (p.name || '').toLowerCase().includes(t) || (p.email || '').toLowerCase().includes(t)
      );
    })
    .sort((a, b) => (sortMode === 'name' ? (a.name || a.email).localeCompare(b.name || b.email) : a.order - b.order));

  $: selectedParticipant =
    participants.find((p) => p.id === selectedParticipantId) || null;
  $: selectedOrder = selectedParticipant
    ? participants.findIndex((p) => p.id === selectedParticipant.id) + 1
    : 0;

  function selectPerson(pid) {
    selectedParticipantId = pid;
  }

  function questionOptionsFor(questionId) {
    const q = questions.find((item) => item.id === questionId);
    return (q && q.options) || [];
  }

  async function pickAnswerOption(participantId, questionId, optionId) {
    const key = `${participantId}:${questionId}`;
    savingAnswerKey = key;
    try {
      await api.events.responses.updateAnswer(id, participantId, questionId, { optionId });
      await loadResponses();
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      savingAnswerKey = '';
    }
  }

  async function saveAnswerText(participantId, questionId, text) {
    const key = `${participantId}:${questionId}`;
    savingAnswerKey = key;
    try {
      await api.events.responses.updateAnswer(id, participantId, questionId, { text });
      await loadResponses();
      showToast('Resposta atualizada!');
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      savingAnswerKey = '';
    }
  }

  function openPhotoDialog() {
    photoDraft = '';
    photoDialogOpen = true;
  }

  function closePhotoDialog() {
    photoDialogOpen = false;
    photoDraft = '';
  }

  function onPhotoDraftChange(e) {
    photoDraft = e.detail;
  }

  async function savePhoto() {
    if (!selectedParticipant) return;
    savingPhoto = true;
    try {
      await api.events.responses.updatePhoto(id, selectedParticipant.id, { photo: photoDraft });
      showToast('Foto atualizada!');
      closePhotoDialog();
      await loadResponses();
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      savingPhoto = false;
    }
  }

  async function removePhoto() {
    if (!selectedParticipant) return;
    savingPhoto = true;
    try {
      await api.events.responses.updatePhoto(id, selectedParticipant.id, { photo: '' });
      showToast('Foto removida.');
      closePhotoDialog();
      await loadResponses();
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      savingPhoto = false;
    }
  }
</script>

<main class="shell">
  {#if loading}
    <div class="edit-center"><p class="text-muted">Carregando…</p></div>
  {:else if notFound}
    <div class="edit-center">
      <div class="card">
        <h1>Evento não encontrado</h1>
        <p class="subtitle">Ele pode ter sido removido ou você não tem acesso.</p>
        <Button block variant="secondary" on:click={back}>Voltar</Button>
      </div>
    </div>
  {:else}
    <TopBar area="Organizador" />

    <CrumbBar
      crumbs={[
        { label: 'Eventos', href: '/dashboard' },
        { label: title || 'Evento', href: `/events/${id}` },
        { label: 'Editar' }
      ]}
    >
      <Chip
        slot="status"
        dot
        label={statusInfo(status).label}
        tint={statusInfo(status).tint}
        color={statusInfo(status).color}
      />
      <svelte:fragment slot="actions">
        <PinChip pin={pinCode} variant="boxed" />
        <Button size="sm" on:click={openOrganizerPanel}>Abrir painel ao vivo</Button>
      </svelte:fragment>
    </CrumbBar>

    <div class="edit-layout" class:rail-collapsed={railCollapsed}>
      <div class="questions-col">
        <Tabs
          bind:value={activeTab}
          tabs={[
            { value: 'questions', label: 'Perguntas', count: questions.length },
            { value: 'responses', label: 'Respostas', count: participantCount }
          ]}
        >
          <svelte:fragment slot="actions">
            {#if activeTab === 'questions' && questions.length > 0}
              <Button size="sm" type="button" on:click={() => openInsertAt(questions.length)}>
                + Adicionar pergunta
              </Button>
            {/if}
          </svelte:fragment>
        </Tabs>

          {#if activeTab === 'responses'}
            {#if participants.length === 0}
              <p class="text-muted empty-note">Ninguém respondeu ainda.</p>
            {:else}
              <div class="responses-split" bind:this={responsesSplitEl}>
                <div class="people-col" style="flex-basis: {peopleColWidth}px">
                  <input
                    class="people-search"
                    type="text"
                    bind:value={personSearch}
                    placeholder="Buscar pessoa ou e-mail"
                  />
                  <div class="sort-toggle">
                    <button
                      type="button"
                      class="sort-btn"
                      class:active={sortMode === 'sequence'}
                      on:click={() => (sortMode = 'sequence')}
                    >Ordem de envio</button>
                    <button
                      type="button"
                      class="sort-btn"
                      class:active={sortMode === 'name'}
                      on:click={() => (sortMode = 'name')}
                    >Nome</button>
                  </div>
                  <ul class="people-list">
                    {#each sortedPeople as p (p.id)}
                      <li>
                        <button
                          type="button"
                          class="person-row"
                          class:selected={selectedParticipant && selectedParticipant.id === p.id}
                          on:click={() => selectPerson(p.id)}
                        >
                          <span class="person-avatar">
                            {#if p.photo}
                              <img src={p.photo} alt="" />
                            {:else}
                              {(p.name || p.email)[0].toUpperCase()}
                            {/if}
                          </span>
                          <span class="person-info">
                            <span class="person-name">{p.name || p.email}</span>
                            <span class="person-meta text-muted">
                              <span class="person-order">#{p.order}</span>
                              <span class="person-meta-sep" aria-hidden="true">·</span>
                              <span class="person-hour">{formatTime(p.createdAt)}</span>
                            </span>
                          </span>
                        </button>
                      </li>
                    {/each}
                  </ul>
                </div>

                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
                <div
                  class="col-resizer"
                  role="separator"
                  aria-orientation="vertical"
                  aria-label="Redimensionar lista de pessoas"
                  tabindex="0"
                  on:mousedown={startPeopleColResize}
                  on:keydown={handlePeopleColResizeKeydown}
                ></div>

                {#if selectedParticipant}
                  <div class="detail-col">
                    <div class="detail-head">
                      <button
                        type="button"
                        class="detail-avatar-btn"
                        title="Editar foto"
                        on:click={openPhotoDialog}
                      >
                        <span class="detail-avatar">
                          {#if selectedParticipant.photo}
                            <img src={selectedParticipant.photo} alt="" />
                          {:else}
                            {(selectedParticipant.name || selectedParticipant.email)[0].toUpperCase()}
                          {/if}
                        </span>
                        <span class="detail-avatar-edit" aria-hidden="true">✎</span>
                      </button>
                      <div class="detail-identity">
                        <strong class="detail-name">{selectedParticipant.name || selectedParticipant.email}</strong>
                        <div class="detail-meta">
                          {#if selectedParticipant.name}
                            <span>{selectedParticipant.email}</span>
                          {/if}
                          <span class="detail-meta-order">envio #{selectedOrder}</span>
                          <span>{formatDateTime(selectedParticipant.createdAt)}</span>
                        </div>
                      </div>
                      <span class="detail-edit-link">
                        <CopyButton
                          text={`${window.location.origin}/answer/${id}?edit=${selectedParticipant.editToken}`}
                          label="Copiar link de edição"
                        />
                      </span>
                    </div>

                    <div class="detail-answers">
                      {#each selectedParticipant.answers as a (a.questionId)}
                        <div class="answer-card">
                          <div class="answer-card-head">
                            <span class="answer-card-question">{a.questionTitle}</span>
                            <span
                              class="badge"
                              class:badge-individual={a.questionType !== 'OPEN_TEXT'}
                              class:badge-open-text={a.questionType === 'OPEN_TEXT'}
                            >
                              {a.questionType === 'OPEN_TEXT' ? 'Resposta aberta' : 'Individual'}
                            </span>
                          </div>

                          {#if a.questionType === 'OPEN_TEXT'}
                            <input
                              class="answer-text-input"
                              value={a.text}
                              disabled={savingAnswerKey === `${selectedParticipant.id}:${a.questionId}`}
                              on:change={(e) =>
                                saveAnswerText(selectedParticipant.id, a.questionId, e.currentTarget.value)}
                              placeholder="Sem resposta"
                            />
                          {:else}
                            <div class="answer-options">
                              {#each questionOptionsFor(a.questionId) as opt (opt.id)}
                                <button
                                  type="button"
                                  class="answer-option"
                                  class:selected={opt.id === a.optionId}
                                  disabled={savingAnswerKey === `${selectedParticipant.id}:${a.questionId}`}
                                  on:click={() => pickAnswerOption(selectedParticipant.id, a.questionId, opt.id)}
                                >
                                  {opt.text}
                                </button>
                              {/each}
                            </div>
                          {/if}
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            {/if}
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
              <!--
                The question rows and the "insert here" dividers are rendered as two
                separate sibling each-blocks (Svelte's animate:flip requires the
                animated element to be the each-block's immediate, sole child — it
                can't share an iteration with the divider li). Visual interleaving
                is done purely with the CSS `order` property below: divider "after
                position p" gets order 2p, question i gets order 2i+1.
              -->
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
                        <span class="question-num">{i + 1}</span>
                        <span class="drag-handle" title="Arraste para reordenar">⠿</span>
                      </div>
                      <div class="question-info">
                        <div class="question-title-row">
                          <strong class="question-title">{q.title}</strong>
                          <Chip
                            shape="square"
                            label={questionKindInfo(q.type).label}
                            tint={questionKindInfo(q.type).tint}
                            color={questionKindInfo(q.type).color}
                          />
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

      <aside class="side-rail" class:collapsed={railCollapsed}>
        <button
          type="button"
          class="icon-btn rail-toggle"
          aria-label={railCollapsed ? 'Expandir painel lateral' : 'Recolher painel lateral'}
          title={railCollapsed ? 'Expandir painel lateral' : 'Recolher painel lateral'}
          aria-expanded={!railCollapsed}
          on:click={() => (railCollapsed = !railCollapsed)}
        >{railCollapsed ? '‹' : '›'}</button>

        {#if !railCollapsed}
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
              <span class="stat-num">{questions.length}</span>
              <span class="stat-label">pergunta{questions.length === 1 ? '' : 's'}</span>
            </div>
            <div class="rail-card stat-box">
              <span class="stat-num">{participantCount}</span>
              <span class="stat-label">{participantCount === 1 ? 'respondeu' : 'responderam'}</span>
            </div>
          </div>

          <div class="rail-card rail-section">
            <span class="overline">Compartilhar</span>
            <div class="share-link">
              <div class="share-link-info">
                <span class="share-link-label">Link de participação</span>
                <span class="share-link-url">{answerLink}</span>
              </div>
              <CopyButton text={answerLink} label="Copiar link de participação" />
            </div>
            <div class="share-link">
              <div class="share-link-info">
                <span class="share-link-label">Apresentação pública</span>
                <span class="share-link-url">{audienceLink}</span>
              </div>
              <CopyButton text={audienceLink} label="Copiar link da apresentação pública" />
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

              <!-- Secundário: a ação primária desta tela é abrir o painel ao vivo,
                   no breadcrumb. -->
              <Button type="submit" variant="secondary" block disabled={submitting}>
                {submitting ? 'Salvando…' : 'Salvar alterações'}
              </Button>
            </form>
          </div>
        {/if}
      </aside>
    </div>
  {/if}
</main>

{#if photoDialogOpen && selectedParticipant}
  <div class="modal-overlay" role="presentation" on:click|self={closePhotoDialog}>
    <div class="modal-card">
      <div class="modal-head">
        <strong>Foto de {selectedParticipant.name || selectedParticipant.email}</strong>
        <p class="text-muted">Envie uma foto e arraste para posicionar o rosto no círculo.</p>
      </div>
      <AvatarCropper on:change={onPhotoDraftChange} />
      <div class="modal-actions">
        {#if selectedParticipant.photo}
          <Button variant="danger" type="button" block disabled={savingPhoto} on:click={removePhoto}>
            Remover foto
          </Button>
        {/if}
        <Button type="button" block disabled={savingPhoto || !photoDraft} on:click={savePhoto}>
          {savingPhoto ? 'Salvando…' : 'Salvar foto'}
        </Button>
      </div>
      <button type="button" class="modal-cancel" on:click={closePhotoDialog}>Cancelar</button>
    </div>
  </div>
{/if}

<style>
  .shell {
    flex: none;
    height: 100vh;
    height: 100dvh;
  }

  .edit-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .edit-layout {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 1fr 320px;
    grid-template-rows: minmax(0, 1fr);
    gap: 20px;
    padding: 24px;
    align-items: stretch;
    transition: grid-template-columns 0.16s ease;
  }

  .edit-layout.rail-collapsed {
    grid-template-columns: 1fr 48px;
  }

  /* O conteúdo manda: perguntas ocupam a coluna larga, configurações viram um
     trilho de apoio à direita. Cada coluna rola por conta própria dentro da
     altura fixa do shell, então a coluna lateral nunca sai da tela. */
  .questions-col {
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    overflow: hidden;
  }

  .side-rail {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 0;
    overflow-y: auto;
  }

  .rail-toggle {
    flex-shrink: 0;
    align-self: flex-end;
  }

  .side-rail.collapsed {
    align-items: center;
    overflow: visible;
  }

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

  .share-link {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 8px;
    background: var(--surface-muted);
  }

  .share-link-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .share-link-label {
    font-size: 0.75rem;
    font-weight: 700;
  }

  .share-link-url {
    font-size: 0.6875rem;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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

  /* --- Perguntas --- */

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
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
    overflow-y: auto;
  }

  .question-item {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    padding: 16px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    cursor: grab;
    transition: border-color 0.15s ease, box-shadow 0.15s ease, opacity 0.15s ease,
      transform 0.15s ease;
  }

  .question-item:hover {
    border-color: var(--accent);
    box-shadow: var(--shadow-hover);
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
    font-family: var(--font-ui);
    font-size: 0.75rem;
    font-weight: 600;
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
    width: 24px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    color: var(--text-subtle);
  }

  .question-num {
    width: 22px;
    height: 22px;
    flex-shrink: 0;
    border-radius: 6px;
    background: var(--surface-muted);
    color: var(--text-muted);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 800;
  }

  .drag-handle {
    font-size: 0.875rem;
    line-height: 1;
    flex-shrink: 0;
  }

  .question-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .question-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .question-title {
    font-size: 0.9375rem;
    line-height: 1.35;
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
    padding: 4px 10px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--surface-muted);
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .question-actions {
    display: flex;
    align-items: flex-start;
    gap: 4px;
    flex-shrink: 0;
  }

  .question-actions .icon-btn {
    width: 30px;
    height: 30px;
    border-color: transparent;
    color: var(--text-subtle);
  }

  .question-actions .icon-btn:hover:not(:disabled) {
    border-color: var(--accent);
    color: var(--accent);
  }

  .question-actions .icon-btn-delete:hover:not(:disabled) {
    border-color: var(--danger);
    color: var(--danger);
  }

  /* --- Respostas --- */

  .responses-split {
    flex: 1;
    min-height: 0;
    display: flex;
    gap: 0;
    align-items: stretch;
  }

  @media (max-width: 760px) {
    .responses-split {
      flex-direction: column;
    }

    .col-resizer {
      display: none;
    }
  }

  .people-col {
    flex: 0 0 250px;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .col-resizer {
    flex: 0 0 17px;
    align-self: stretch;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: col-resize;
    touch-action: none;
  }

  .col-resizer::before {
    content: '';
    width: 3px;
    height: 100%;
    border-radius: 999px;
    background: var(--border);
    transition: background 0.15s ease;
  }

  .col-resizer:hover::before,
  .col-resizer:focus-visible::before {
    background: var(--accent);
  }

  .col-resizer:focus-visible {
    outline: none;
  }

  .people-search {
    width: 100%;
    padding: 10px 12px;
    border-radius: var(--radius-control);
    border: 1.5px solid var(--border-strong);
    background: var(--bg-elev);
    color: var(--text);
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    outline: none;
    transition: border-color 0.15s ease;
  }

  .people-search:focus {
    border-color: var(--accent);
  }

  .sort-toggle {
    display: flex;
    gap: 4px;
    padding: 3px;
    border-radius: var(--radius-control);
    background: var(--surface-muted);
  }

  .sort-btn {
    flex: 1;
    padding: 6px 10px;
    border-radius: 7px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.75rem;
    font-weight: 700;
    cursor: pointer;
  }

  .sort-btn.active {
    background: var(--bg-elev);
    color: var(--text);
    box-shadow: 0 1px 4px rgba(23, 21, 42, 0.1);
  }

  .people-list {
    list-style: none;
    margin: 0;
    padding: 0;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .person-row {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border-radius: var(--radius-row);
    text-align: left;
    cursor: pointer;
    font-family: var(--font-ui);
    background: var(--bg-elev);
    border: 1px solid var(--border);
    color: var(--text);
    transition: border-color 0.15s ease;
  }

  .person-row:hover {
    border-color: var(--accent);
  }

  .person-row.selected {
    background: var(--tint-purple);
    border-color: var(--accent);
  }

  .person-avatar,
  .detail-avatar {
    width: 36px;
    height: 36px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--accent);
    color: var(--on-accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.875rem;
    font-weight: 800;
    overflow: hidden;
  }

  .person-avatar img,
  .detail-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .person-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .person-name {
    font-size: 0.875rem;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .person-meta {
    display: flex;
    align-items: baseline;
    gap: 5px;
    font-size: 0.75rem;
  }

  .person-order {
    font-weight: 700;
    color: var(--text-subtle);
  }

  .person-meta-sep {
    color: var(--text-subtle);
  }

  .detail-col {
    flex: 1;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    overflow: hidden;
  }

  .detail-head {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    padding: 12px 22px;
    border-bottom: 1px solid var(--border);
  }

  .detail-avatar-btn {
    position: relative;
    width: 46px;
    height: 46px;
    flex-shrink: 0;
    padding: 0;
    border: none;
    background: transparent;
    cursor: pointer;
  }

  .detail-avatar {
    width: 46px;
    height: 46px;
    font-size: 1.0625rem;
  }

  .detail-avatar-edit {
    position: absolute;
    right: -2px;
    bottom: -2px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 2px solid var(--bg-elev);
    background: var(--surface-muted);
    color: var(--orange);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.625rem;
  }

  .detail-identity {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .detail-name {
    font-size: 1rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .detail-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 4px 10px;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .detail-meta-order {
    font-weight: 700;
  }

  .detail-edit-link {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    height: 32px;
  }

  .detail-answers {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px 22px 20px;
    overflow-y: auto;
  }

  .answer-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 14px 16px;
    border-radius: var(--radius-control);
    background: var(--surface-muted);
    border: 1px solid var(--border);
  }

  .answer-card-head {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .answer-card-question {
    flex: 1;
    min-width: 0;
    font-size: 0.875rem;
    font-weight: 700;
  }

  .answer-options {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .answer-option {
    padding: 7px 13px;
    border-radius: var(--radius-control);
    cursor: pointer;
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 700;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    color: var(--text-muted);
    transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  }

  .answer-option:hover:not(:disabled) {
    border-color: var(--accent);
    color: var(--text);
  }

  .answer-option.selected {
    border: 2px solid var(--accent);
    color: var(--text);
    padding: 6px 12px;
  }

  .answer-option:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .answer-text-input {
    padding: 10px 12px;
    border-radius: var(--radius-control);
    border: 1.5px solid var(--border-strong);
    background: var(--bg-elev);
    font-family: var(--font-ui);
    font-size: 0.875rem;
    color: var(--text);
    outline: none;
  }

  .answer-text-input:focus {
    border-color: var(--accent);
  }

  /* --- Modal (edição de foto) --- */

  .modal-overlay {
    position: fixed;
    inset: 0;
    z-index: 60;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    overflow-y: auto;
    background: rgba(23, 21, 42, 0.45);
  }

  .modal-card {
    width: 420px;
    max-width: 100%;
    max-height: calc(100vh - 48px);
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding: 32px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
  }

  .modal-head {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .modal-head strong {
    font-size: 1.125rem;
  }

  .modal-head p {
    margin: 0;
    font-size: 0.875rem;
  }

  .modal-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .modal-cancel {
    align-self: center;
    padding: 6px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 700;
    cursor: pointer;
  }

  .modal-cancel:hover {
    color: var(--text);
  }

  /* Em telas estreitas a coluna lateral desce pra baixo das perguntas (ver
     .edit-layout acima) — nesse layout empilhado não faz sentido cada bloco
     rolar por conta própria, então aqui isso é desfeito e quem rola é a
     página toda, de cima a baixo, como qualquer tela pública. */
  @media (max-width: 980px) {
    .edit-layout {
      display: flex;
      flex-direction: column;
      overflow-y: auto;
    }

    .questions-col,
    .side-rail {
      overflow: visible;
    }

    .question-list,
    .responses-split,
    .people-list,
    .detail-answers {
      flex: none;
      max-height: none;
      overflow: visible;
    }
  }
</style>
