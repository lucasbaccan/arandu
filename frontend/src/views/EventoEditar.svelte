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
  import RailContent from '../components/RailContent.svelte';

  function formatTime(iso) {
    const d = new Date(iso);
    const hh = String(d.getUTCHours()).padStart(2, '0');
    const min = String(d.getUTCMinutes()).padStart(2, '0');
    return `${hh}:${min}`;
  }

  let loading = true;
  let error = '';
  let loadError = '';
  let notFound = false;
  let activeTab = 'questions'; // questions | responses | settings
  let railCollapsed = false;
  let isMobile = false;

  $: tabsList = [
    { value: 'questions', label: 'Perguntas', count: questions.length },
    { value: 'responses', label: 'Respostas', count: participantCount },
    ...(isMobile ? [{ value: 'settings', label: 'Configuração' }] : [])
  ];

  $: linkResposta = `${window.location.origin}/responder/${id}`;
  $: linkPlateia = pinCode
    ? `${window.location.origin}/plateia/${id}?pin=${encodeURIComponent(pinCode.toUpperCase())}`
    : `${window.location.origin}/plateia/${id}`;

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
    const { questions: qs } = await api.eventos.perguntas.listar(id);
    questions = qs;
  }

  async function loadResponses() {
    const { participantCount: pc, participants: parts } = await api.eventos.respostas.listar(id);
    participantCount = pc;
    participants = parts;
    if (!selectedParticipantId || !parts.some((p) => p.id === selectedParticipantId)) {
      selectedParticipantId = parts[0] ? parts[0].id : null;
    }
  }

  async function load() {
    try {
      const [{ event }, { questions: qs }] = await Promise.all([
        api.eventos.buscar(id),
        api.eventos.perguntas.listar(id),
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
        loadError = e.message;
      }
    } finally {
      loading = false;
      questionsLoading = false;
    }
  }
  load();

  if (window.matchMedia) {
    const mobileMq = window.matchMedia('(max-width: 980px)');
    isMobile = mobileMq.matches;
    mobileMq.addEventListener('change', (e) => {
      isMobile = e.matches;
    });
  }

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
      await api.eventos.atualizar(id, {
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
      const { event } = await api.eventos.atualizar(id, {
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
      const { question } = await api.eventos.perguntas.criar(id, body);
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
      const { question } = await api.eventos.perguntas.atualizar(id, questionId, body);
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
      await api.eventos.perguntas.remover(id, q.id);
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
      await api.eventos.perguntas.reordenar(
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
    navigate('/painel');
  }

  function abrirPalco() {
    navigate(`/palco/${id}`);
  }

  let deletingEvent = false;

  async function deleteEvent() {
    if (
      !window.confirm(
        `Excluir o evento "${title}"? Todas as perguntas, respostas e fotos serão apagadas. Essa ação não pode ser desfeita.`
      )
    ) {
      return;
    }
    deletingEvent = true;
    try {
      await api.eventos.deletar(id);
      showToast('Evento excluído.');
      navigate('/painel');
    } catch (e) {
      showToast(e.message, 'error');
      deletingEvent = false;
    }
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
      await api.eventos.respostas.atualizarResposta(id, participantId, questionId, { optionId });
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
      await api.eventos.respostas.atualizarResposta(id, participantId, questionId, { text });
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
      await api.eventos.respostas.atualizarFoto(id, selectedParticipant.id, { photo: photoDraft });
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
      await api.eventos.respostas.atualizarFoto(id, selectedParticipant.id, { photo: '' });
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
        { label: 'Eventos', href: '/painel' },
        { label: title || 'Evento', href: `/evento/${id}` }
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
        <Button size="sm" on:click={abrirPalco}>Abrir painel ao vivo</Button>
      </svelte:fragment>
    </CrumbBar>

    {#if loadError}
      <p class="form-error edit-top-error">{loadError}</p>
    {/if}

    <div class="edit-layout" class:rail-collapsed={railCollapsed}>
      <div class="questions-col">
        <Tabs
          bind:value={activeTab}
          tabs={tabsList}
        >
          <svelte:fragment slot="actions">
            {#if activeTab === 'questions' && questions.length > 0}
              <Button size="sm" type="button" on:click={() => openInsertAt(questions.length)}>
                + Adicionar pergunta
              </Button>
            {/if}
          </svelte:fragment>
        </Tabs>

          {#if isMobile && activeTab === 'settings'}
            <div class="rail-mobile">
              <RailContent
                {status}
                {statusBusy}
                questionCount={questions.length}
                {participantCount}
                {linkResposta}
                {linkPlateia}
                bind:title
                bind:pinCode
                bind:showRanking
                bind:allowEdit
                {error}
                {submitting}
                {toggleAnswersOpen}
                {handleSubmit}
                onDeleteEvent={deleteEvent}
                {deletingEvent}
              />
            </div>
          {:else if activeTab === 'responses'}
            {#if participants.length === 0}
              <p class="text-muted empty-note">Ninguém respondeu ainda.</p>
            {:else}
              <div class="responses-split" bind:this={responsesSplitEl}>
                <div class="people-col" style="--people-basis: {peopleColWidth}px">
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
                            <span class="person-name" title={p.name || p.email}>{p.name || p.email}</span>
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
                        <strong class="detail-name" title={selectedParticipant.name || selectedParticipant.email}>{selectedParticipant.name || selectedParticipant.email}</strong>
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
                          text={`${window.location.origin}/responder/${id}?edit=${selectedParticipant.editToken}`}
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
                        ><span class="msr icon-btn-move-glyph">arrow_upward</span></button>
                        <button
                          type="button"
                          class="icon-btn icon-btn-move"
                          title="Mover para baixo"
                          aria-label="Mover pergunta para baixo"
                          disabled={i === questions.length - 1}
                          on:click={() => moveQuestion(i, i + 1)}
                        ><span class="msr icon-btn-move-glyph">arrow_downward</span></button>
                        <button
                          type="button"
                          class="icon-btn icon-btn-edit"
                          title="Editar pergunta"
                          aria-label={`Editar pergunta ${q.title}`}
                          on:click={() => startEdit(q)}
                        ><span class="msr icon-btn-edit-glyph">edit_square</span></button>
                        <button
                          type="button"
                          class="icon-btn icon-btn-delete"
                          title="Remover pergunta"
                          aria-label={`Remover pergunta ${q.title}`}
                          on:click={() => removeQuestion(q)}
                        ><span class="msr icon-btn-delete-glyph">delete</span></button>
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

          <footer class="edit-footer" aria-hidden="true"></footer>
      </div>

      <aside class="side-rail" class:collapsed={railCollapsed}>
        <div class="rail-top">
          <button
            type="button"
            class="icon-btn rail-toggle"
            aria-label={railCollapsed ? 'Expandir painel lateral' : 'Recolher painel lateral'}
            title={railCollapsed ? 'Expandir painel lateral' : 'Recolher painel lateral'}
            aria-expanded={!railCollapsed}
            on:click={() => (railCollapsed = !railCollapsed)}
          >
            {#if railCollapsed}
              <span class="msr rail-toggle-icon">right_panel_open</span>
            {:else}
              <span class="msr rail-toggle-icon">right_panel_close</span>
            {/if}
          </button>
        </div>

        {#if !railCollapsed}
          <div class="rail-scroll">
            <RailContent
              {status}
              {statusBusy}
              questionCount={questions.length}
              {participantCount}
              {linkResposta}
              {linkPlateia}
              bind:title
              bind:pinCode
              bind:showRanking
              bind:allowEdit
              {error}
              {submitting}
              {toggleAnswersOpen}
              {handleSubmit}
              onDeleteEvent={deleteEvent}
              {deletingEvent}
            />
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
    position: relative;
    display: grid;
    grid-template-columns: 1fr 320px;
    grid-template-rows: minmax(0, 1fr);
    gap: 20px;
    padding: 24px;
    align-items: stretch;
    transition: grid-template-columns 0.16s ease;
  }

  .edit-layout.rail-collapsed {
    grid-template-columns: 1fr;
  }

  /* Com o trilho colapsado, o botão flutuante ocupa o canto superior direito;
     reserva espaço pra não cobrir o "Adicionar pergunta" do Tabs. */
  .edit-layout.rail-collapsed .questions-col {
    padding-right: 60px;
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
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 0;
    overflow: hidden;
  }

  .rail-mobile {
    display: none;
    flex-direction: column;
    gap: 12px;
  }

  /* Espaço inferior no mobile: a página rola inteira e o conteúdo (lista de
     perguntas / respostas / configurações) chegava colado na borda de baixo. */
  .edit-footer {
    display: none;
  }

  .rail-mobile :global(.answers-switch),
  .rail-mobile :global(.rail-card) {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    padding: 16px;
  }

  .rail-top {
    flex-shrink: 0;
    display: flex;
    justify-content: flex-end;
  }

  .rail-scroll {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
    overflow-y: auto;
  }

  .rail-toggle {
    flex-shrink: 0;
    display: grid;
    place-items: center;
  }

  .rail-toggle-icon {
    font-size: 20px;
    font-variation-settings: 'FILL' 0, 'wght' 500, 'GRAD' 0, 'opsz' 24;
  }

  .side-rail.collapsed {
    position: absolute;
    top: 24px;
    right: 24px;
    width: 48px;
    overflow: visible;
  }

  .side-rail.collapsed .rail-top {
    justify-content: center;
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
    color: var(--accent);
    background: var(--accent-soft);
  }

  .question-actions .icon-btn:hover:not(:disabled) {
    border-color: transparent;
    background: var(--accent);
    color: var(--on-accent);
  }

  .question-actions .icon-btn-edit {
    background: var(--tint-cyan);
    color: var(--cyan-text);
  }

  .question-actions .icon-btn-edit:hover:not(:disabled) {
    background: var(--cyan);
    color: var(--on-accent);
  }

  .icon-btn-edit-glyph {
    font-size: 18px;
    font-variation-settings: 'FILL' 0, 'wght' 500, 'GRAD' 0, 'opsz' 24;
  }

  .icon-btn-move-glyph,
  .icon-btn-delete-glyph {
    font-size: 18px;
    font-variation-settings: 'FILL' 0, 'wght' 500, 'GRAD' 0, 'opsz' 24;
  }

  .question-actions .icon-btn-delete {
    background: var(--tint-danger);
    color: var(--danger);
  }

  .question-actions .icon-btn-delete:hover:not(:disabled) {
    border-color: transparent;
    background: var(--danger);
    color: var(--on-accent);
  }

  .question-actions .icon-btn:disabled {
    background: var(--surface-muted);
    color: var(--text-subtle);
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
    flex: 0 0 var(--people-basis, 250px);
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

    .rail-scroll {
      overflow: visible;
      flex: none;
    }

    /* No mobile as configurações viram a aba "Configuração"; o trilho
       lateral não existe mais (ficaria sobre as perguntas). */
    .side-rail {
      display: none;
    }

    .rail-mobile {
      display: flex;
    }

    .edit-footer {
      display: block;
      height: 32px;
      flex: none;
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

  /* --- Mobile --- */

  /* Badges "Individual"/"Resposta aberta" usavam classes sem definição (caiam
     no .badge neutro). Agora diferenciam visualmente, com tokens tema-aware. */
  .badge-individual {
    color: var(--success-text);
    background: var(--tint-success);
  }

  .badge-open-text {
    color: var(--purple-text);
    background: var(--tint-purple);
  }

  @media (max-width: 760px) {
    /* Lista de participantes: o basis inline virava ALTURA fixa de 250px e o
       excedente ficava pintado atrás do painel de detalhes. Libera a altura e
       devolve a rolagem interna à lista. */
    .people-col {
      flex: 0 0 auto;
      height: auto;
      /* min() evita 1 linha visível em landscape (45vh de ~375px de altura). */
      max-height: min(45vh, 320px);
    }

    .people-list {
      flex: 1 1 auto;
      min-height: 0;
      max-height: none;
      overflow-y: auto;
    }

    /* Respostas da pessoa: o badge "Individual"/"Resposta aberta" desce para
       baixo do título da pergunta, antes da resposta, em vez de ficar ao lado. */
    .answer-card-head {
      flex-direction: column;
      align-items: flex-start;
      gap: 6px;
    }

    .answer-card-question {
      flex: none;
      width: 100%;
    }
  }

  @media (max-width: 560px) {
    /* Linha da pergunta: texto + chip + 4 ações espremidos em ~295px.
       Empilha as ações abaixo do texto com alvos de 44px (o ↑↓ é o único
       reordenador que funciona em touch — drag & drop nativo não). */
    .question-item {
      flex-wrap: wrap;
      gap: 10px;
    }

    .reorder-controls {
      flex-direction: column;
      width: auto;
    }

    .question-actions {
      width: 100%;
      justify-content: flex-end;
    }

    .sort-btn {
      padding: 10px 10px;
    }

    /* 16px evita o zoom automático do iOS ao focar. */
    .people-search {
      font-size: 1rem;
    }

    .answer-text-input {
      font-size: 1rem;
    }

    .modal-cancel {
      padding: 12px 8px;
    }

    .modal-card {
      padding: 20px;
    }
  }

  /* Touch: alvos de 44px nas ações da pergunta (↑↓✎✕), sort e opções de
     resposta — vale em QUALQUER largura (tablets/landscape inclusos), não só
     ≤560px: a regra escopada de 30px sobrescrevia o bump global do app.css. */
  @media (pointer: coarse) {
    .question-actions .icon-btn {
      width: 44px;
      height: 44px;
    }

    .sort-btn {
      min-height: 44px;
    }

    .answer-option {
      min-height: 44px;
    }

    .people-search {
      font-size: 1rem;
    }
  }

  /* Em touch não existe hover: os divisores de inserção ficavam invisíveis.
     (Desaninhado do ≤560px: antes só valia em retrato; tablets e landscape
     seguiam sem os divisores.) */
  @media (hover: none) {
    .insert-btn {
      opacity: 0.85;
      padding: 10px 0;
    }
  }
</style>
