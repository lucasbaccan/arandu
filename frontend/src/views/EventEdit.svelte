<script>
  export let id = '';

  import { flip } from 'svelte/animate';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { showToast } from '../lib/toastStore.js';
  import { formatDateTime } from '../lib/formatDate.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import CopyButton from '../components/CopyButton.svelte';
  import Switch from '../components/Switch.svelte';
  import QuestionForm, { MIN_OPTIONS, MAX_OPTIONS } from '../components/QuestionForm.svelte';
  import AvatarCropper from '../components/AvatarCropper.svelte';

  const statusLabels = {
    PREPARATION: 'Em preparação',
    OPEN_FOR_ANSWERS: 'Coletando respostas',
    CLOSED_FOR_ANSWERS: 'Respostas encerradas',
    PRESENTING: 'Ao vivo',
    FINISHED: 'Finalizado'
  };

  function formatTime(iso) {
    const d = new Date(iso);
    const hh = String(d.getUTCHours()).padStart(2, '0');
    const min = String(d.getUTCMinutes()).padStart(2, '0');
    return `${hh}:${min}`;
  }

  function statusVisual(s) {
    if (s === 'OPEN_FOR_ANSWERS') return { tint: 'var(--tint-purple)', color: 'var(--accent)' };
    if (s === 'PRESENTING') return { tint: 'var(--tint-success)', color: 'var(--success)' };
    return { tint: 'var(--bg-input)', color: 'var(--text-muted)' };
  }

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
  // Não é persistido: o backend ainda não tem um campo para isto — cada
  // participante sempre recebe um link de edição. Mantido local para
  // refletir a opção do design enquanto essa capacidade não existe na API.
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
        configShowRanking: showRanking
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

  function statusLabel(s) {
    return statusLabels[s] || s;
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
      <div class="editor-topbar">
        <button
          type="button"
          class="back-link"
          aria-label="Voltar para meus eventos"
          on:click={back}
        >‹ Eventos</button>
        <img class="topbar-logo" src="/img/arandu-logo.png" alt="Arandu" />
        <span class="topbar-title">{title || 'Editar evento'}</span>
        <span
          class="status-pill"
          style="background:{statusVisual(status).tint};color:{statusVisual(status).color}"
        >{statusLabel(status)}</span>
        <span class="topbar-spacer"></span>
        <div class="topbar-pin">
          <span class="topbar-pin-label">PIN</span>
          <span class="topbar-pin-value">{pinCode.toUpperCase()}</span>
          <CopyButton text={pinCode.toUpperCase()} label="Copiar PIN" />
        </div>
        <Button
          size="sm"
          variant={status === 'PRESENTING' ? 'success-invert' : 'accent-invert'}
          on:click={openOrganizerPanel}
        >
          ▶ Painel do organizador
        </Button>
      </div>

      <div class="edit-layout">
        <aside class="card panel settings-panel">
          <div class="stat-row">
            <div class="stat-box">
              <span class="stat-num">{questions.length}</span>
              <span class="stat-label">pergunta{questions.length === 1 ? '' : 's'}</span>
            </div>
            <div class="stat-box">
              <span class="stat-num">{participantCount}</span>
              <span class="stat-label">{participantCount === 1 ? 'pessoa respondeu' : 'pessoas responderam'}</span>
            </div>
          </div>

          <div
            class="answers-switch"
            style="background:{status === 'OPEN_FOR_ANSWERS' ? 'var(--tint-success)' : 'var(--bg-input)'}"
          >
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

          <div class="sidebar-section">
            <span class="sidebar-heading">Compartilhar</span>
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

          <div class="sidebar-section">
            <span class="sidebar-heading">Configurações</span>
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

              <Button type="submit" block disabled={submitting}>
                {submitting ? 'Salvando…' : 'Salvar alterações'}
              </Button>
            </form>
          </div>
        </aside>

        <div class="card panel questions-panel">
          <div class="tabs" role="tablist">
            <button
              type="button"
              role="tab"
              class="pill-tab"
              class:active={activeTab === 'questions'}
              aria-selected={activeTab === 'questions'}
              on:click={() => (activeTab = 'questions')}
            >
              Perguntas
            </button>
            <button
              type="button"
              role="tab"
              class="pill-tab"
              class:active={activeTab === 'responses'}
              aria-selected={activeTab === 'responses'}
              on:click={() => (activeTab = 'responses')}
            >
              Respostas
            </button>
            <span class="tabs-spacer"></span>
            {#if activeTab === 'questions' && questions.length > 0}
              <Button size="sm" type="button" on:click={() => openInsertAt(questions.length)}>
                + Adicionar pergunta
              </Button>
            {/if}
          </div>

          {#if activeTab === 'responses'}
            {#if participants.length === 0}
              <p class="text-muted empty-note">Ninguém respondeu ainda.</p>
            {:else}
              <div class="responses-split">
                <div class="people-col">
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
                          <span class="person-order">#{p.order}</span>
                          <span class="person-avatar">
                            {#if p.photo}
                              <img src={p.photo} alt="" />
                            {:else}
                              {(p.name || p.email)[0].toUpperCase()}
                            {/if}
                          </span>
                          <span class="person-info">
                            <span class="person-name">{p.name || p.email}</span>
                            <span class="person-hour text-muted">{formatTime(p.createdAt)}</span>
                          </span>
                        </button>
                      </li>
                    {/each}
                  </ul>
                </div>

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
      </div>
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

  .editor-topbar {
    height: 52px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 0 20px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
  }

  .back-link {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 0;
    border: none;
    background: transparent;
    font-family: inherit;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-muted);
    white-space: nowrap;
    cursor: pointer;
  }

  .back-link:hover {
    color: var(--accent);
  }

  .topbar-logo {
    flex-shrink: 0;
    height: 24px;
    width: auto;
  }

  .topbar-title {
    flex: 0 1 auto;
    min-width: 0;
    font-size: 0.95rem;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .status-pill {
    flex-shrink: 0;
    display: inline-block;
    padding: 3px 10px;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .topbar-spacer {
    flex: 1;
  }

  .topbar-pin {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 5px 8px 5px 12px;
    border-radius: 999px;
    white-space: nowrap;
    background: var(--bg-input);
  }

  .topbar-pin-label {
    font-size: 0.65rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    color: var(--text-muted);
  }

  .topbar-pin-value {
    font-size: 0.85rem;
    font-weight: 700;
    letter-spacing: 0.12em;
  }

  .edit-wrap {
    width: 100%;
    max-width: 1160px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .tabs {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-bottom: 12px;
    margin-bottom: 12px;
    border-bottom: 1px solid var(--border);
  }

  .tabs-spacer {
    flex: 1;
  }

  .pill-tab {
    padding: 9px 16px;
    border-radius: 8px;
    border: 2px solid var(--accent);
    background: transparent;
    color: var(--accent);
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.15s ease, color 0.15s ease;
  }

  .pill-tab.active {
    background: var(--accent);
    color: #fff;
  }

  .edit-layout {
    display: grid;
    grid-template-columns: 320px 1fr;
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

  .settings-panel {
    gap: 20px;
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
    padding: 12px 14px;
    border-radius: 10px;
    background: var(--bg-input);
  }

  .stat-num {
    font-size: 1.4rem;
    font-weight: 800;
    line-height: 1;
  }

  .stat-label {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .answers-switch {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 10px;
  }

  .answers-switch strong {
    font-size: 0.9rem;
  }

  .answers-switch p {
    margin: 2px 0 0;
    font-size: 0.8rem;
  }

  .sidebar-section {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .sidebar-heading {
    font-size: 0.85rem;
    font-weight: 700;
  }

  .share-link {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: var(--bg-input);
  }

  .share-link-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .share-link-label {
    font-size: 0.8rem;
    font-weight: 600;
  }

  .share-link-url {
    font-size: 0.7rem;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .config-row {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 0.85rem;
  }

  .config-row-hint {
    align-items: flex-start;
  }

  .config-row-hint p {
    margin: 2px 0 0;
    font-size: 0.75rem;
    line-height: 1.35;
  }

  .questions-panel {
    gap: 0;
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
    width: 34px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }

  .question-num {
    width: 26px;
    height: 26px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--bg-elev);
    color: var(--text-muted);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.8rem;
    font-weight: 800;
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
    align-items: flex-start;
    gap: 6px;
    flex-shrink: 0;
  }

  /* Respostas */

  .responses-split {
    display: flex;
    gap: 18px;
    align-items: flex-start;
  }

  .people-col {
    flex: 0 1 280px;
    min-width: 180px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .people-search {
    width: 100%;
    padding: 9px 12px;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: var(--bg-input);
    color: var(--text);
    font-family: inherit;
    font-size: 0.85rem;
  }

  .sort-toggle {
    display: flex;
    gap: 4px;
    padding: 3px;
    border-radius: 9px;
    background: var(--bg-input);
  }

  .sort-btn {
    flex: 1;
    padding: 6px 10px;
    border-radius: 7px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-family: inherit;
    font-size: 0.78rem;
    font-weight: 600;
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
    max-height: 480px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
    mask-image: linear-gradient(to bottom, #000 calc(100% - 24px), transparent);
  }

  .person-row {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border-radius: 10px;
    text-align: left;
    cursor: pointer;
    font-family: inherit;
    background: transparent;
    border: 1px solid var(--border);
    color: var(--text);
  }

  .person-row.selected {
    background: var(--tint-purple);
    border-color: var(--accent);
  }

  .person-order {
    width: 24px;
    flex-shrink: 0;
    text-align: right;
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--text-muted);
  }

  .person-avatar,
  .detail-avatar {
    width: 32px;
    height: 32px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--accent);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.8rem;
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
    gap: 1px;
  }

  .person-name {
    font-size: 0.85rem;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .person-hour {
    font-size: 0.7rem;
  }

  .detail-col {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 12px;
    overflow: hidden;
  }

  .detail-head {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    padding: 20px 22px;
    border-bottom: 1px solid var(--border);
  }

  .detail-avatar-btn {
    position: relative;
    width: 64px;
    height: 64px;
    flex-shrink: 0;
    padding: 0;
    border: none;
    background: transparent;
    cursor: pointer;
  }

  .detail-avatar {
    width: 64px;
    height: 64px;
    font-size: 1.4rem;
  }

  .detail-avatar-edit {
    position: absolute;
    right: -2px;
    bottom: -2px;
    width: 26px;
    height: 26px;
    border-radius: 50%;
    border: 2px solid var(--bg-elev);
    background: var(--bg-input);
    color: var(--orange);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
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
    font-size: 0.8rem;
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
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px 22px 20px;
    max-height: 480px;
    overflow-y: auto;
  }

  .answer-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 14px 16px;
    border-radius: 10px;
    background: var(--bg-input);
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
    font-size: 0.88rem;
    font-weight: 600;
  }

  .answer-options {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .answer-option {
    padding: 7px 13px;
    border-radius: 999px;
    cursor: pointer;
    font-family: inherit;
    font-size: 0.8rem;
    font-weight: 600;
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
    color: var(--text);
    transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  }

  .answer-option.selected {
    background: var(--accent);
    border-color: var(--accent);
    color: #fff;
  }

  .answer-option:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .answer-text-input {
    padding: 9px 12px;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: var(--bg-elev);
    font-family: inherit;
    font-size: 0.85rem;
    color: var(--text);
  }

  /* Modal (edição de foto) */

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
    padding: 32px 28px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 16px;
    box-shadow: var(--shadow);
  }

  .modal-head {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .modal-head strong {
    font-size: 1.2rem;
  }

  .modal-head p {
    margin: 0;
    font-size: 0.9rem;
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
    font-family: inherit;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
  }

  .modal-cancel:hover {
    color: var(--text);
  }
</style>
