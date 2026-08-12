<script>
  export let id = '';

  import { onDestroy } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { flip } from 'svelte/animate';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import Switch from '../components/Switch.svelte';
  import ReactionBurstLayer from '../components/ReactionBurstLayer.svelte';
  import { fireReaction } from '../lib/reactionStore.js';

  let loading = true;
  let error = '';

  let event = null;
  let questions = [];
  let participants = [];
  let currentIndex = 0;

  let blanked = false;
  let answersHidden = false;
  let interactionsEnabled = true;
  let message = '';
  let messageDraft = '';
  let qaInbox = [];
  let adminEventSource = null;

  onDestroy(() => {
    if (adminEventSource) adminEventSource.close();
  });

  const stageStatusMeta = {
    PREPARATION: { label: 'Em preparação', tint: 'var(--tint-orange)', color: 'var(--orange)' },
    OPEN_FOR_ANSWERS: { label: 'Coletando respostas', tint: 'var(--tint-cyan)', color: 'var(--cyan-hover)' },
    CLOSED_FOR_ANSWERS: { label: 'Respostas encerradas', tint: 'rgba(104, 103, 122, 0.12)', color: 'var(--text-muted)' },
    PRESENTING: { label: 'Ao vivo', tint: 'var(--tint-success)', color: 'var(--success)' },
    FINISHED: { label: 'Finalizado', tint: 'rgba(104, 103, 122, 0.12)', color: 'var(--text-muted)' }
  };

  $: eventStatusInfo = event
    ? stageStatusMeta[event.status] || { label: event.status, tint: 'var(--bg-input)', color: 'var(--text-muted)' }
    : null;

  // Modo apresentação: some com os controles e números de admin (contagens,
  // pergunta X de Y, botões) pra o organizador poder compartilhar a tela
  // (ex: Google Meet) mostrando só os rostos e as respostas. Navegação e
  // ações continuam disponíveis pelo teclado.
  let presentationMode = false;

  function togglePresentationMode() {
    presentationMode = !presentationMode;
  }

  function handleKeydown(e) {
    // Atalhos são globais (svelte:window), mas não podem competir com
    // digitação normal em campos de texto — ex: o aviso pra tela dos
    // participantes, onde espaço/R/P precisam virar caracteres, não ações.
    const tag = e.target.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA' || e.target.isContentEditable) {
      return;
    }
    if (e.key === 'Escape') {
      if (presentationMode) presentationMode = false;
      return;
    }
    if (!currentQuestion) return;
    if (e.key === 'ArrowRight' || e.key === ' ') {
      e.preventDefault();
      goNext();
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault();
      goPrev();
    } else if (e.key.toLowerCase() === 'r') {
      revealAll();
    } else if (e.key.toLowerCase() === 'p') {
      togglePresentationMode();
    }
  }

  // revealed[questionId] = Set com os ids dos participantes já revelados nessa pergunta
  let revealed = {};

  load();

  async function load() {
    try {
      const [{ event: ev }, { questions: qs }, { participants: ps }] = await Promise.all([
        api.events.get(id),
        api.events.questions.list(id),
        api.events.responses.list(id)
      ]);
      event = ev;
      questions = qs;
      participants = ps;
      revealed = Object.fromEntries(qs.map((q) => [q.id, new Set()]));
      if (qs.length > 0) syncQuestion(qs[0].id);
      connectAdminStream();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function back() {
    navigate(`/events/${id}`);
  }

  // Espelha o estado ao vivo pro servidor (best-effort — não bloqueia nem
  // quebra a UI local se a rede falhar), pra quem está assistindo em
  // /audience/:id ver a mesma coisa em tempo real.
  function syncQuestion(questionId) {
    api.events.live.setQuestion(id, questionId).catch(() => {});
  }

  // Conexão só de leitura: traz de volta blank/aviso/interações se a página
  // recarregar, e entrega as reações e mensagens de Q&A que chegam ao vivo.
  function connectAdminStream() {
    if (adminEventSource) adminEventSource.close();
    adminEventSource = new EventSource(api.events.live.adminStreamUrl(id));
    adminEventSource.onmessage = (e) => {
      const snap = JSON.parse(e.data);
      blanked = snap.blanked;
      answersHidden = snap.answersHidden;
      interactionsEnabled = snap.interactionsEnabled;
      message = snap.message;
      messageDraft = snap.message;
      qaInbox = snap.qaInbox;
    };
    adminEventSource.addEventListener('reaction', (e) => {
      fireReaction(JSON.parse(e.data).emoji);
    });
  }

  function toggleBlanked() {
    blanked = !blanked;
    api.events.live.setBlanked(id, blanked).catch(() => {});
  }

  function toggleAnswersHidden() {
    answersHidden = !answersHidden;
    api.events.live.setAnswersHidden(id, answersHidden).catch(() => {});
  }

  function toggleInteractions() {
    interactionsEnabled = !interactionsEnabled;
    api.events.live.setInteractionsEnabled(id, interactionsEnabled).catch(() => {});
  }

  function sendMessage() {
    message = messageDraft.trim();
    api.events.live.setMessage(id, message).catch(() => {});
  }

  function clearMessage() {
    message = '';
    messageDraft = '';
    api.events.live.setMessage(id, '').catch(() => {});
  }

  function dismissQA(messageId) {
    qaInbox = qaInbox.filter((m) => m.id !== messageId);
    api.events.live.dismissQA(id, messageId).catch(() => {});
  }

  function openAudienceScreen() {
    window.open(`/audience/${id}?pin=${encodeURIComponent(event.pinCode.toUpperCase())}`, '_blank');
  }

  $: currentQuestion = questions[currentIndex];
  $: revealedIds = currentQuestion ? revealed[currentQuestion.id] : new Set();

  function answerFor(p, q) {
    return p.answers.find((a) => a.questionId === q.id) || null;
  }

  $: pending = currentQuestion ? participants.filter((p) => !revealedIds.has(p.id)) : [];

  $: choiceGroups =
    currentQuestion && currentQuestion.type !== 'OPEN_TEXT'
      ? currentQuestion.options.map((opt) => ({
          label: opt.text,
          participants: participants.filter((p) => {
            if (!revealedIds.has(p.id)) return false;
            const a = answerFor(p, currentQuestion);
            return a && a.optionId === opt.id;
          })
        }))
      : [];

  // Normaliza a resposta aberta (trim + primeira letra maiúscula) e usa o
  // resultado para agrupar quem respondeu a mesma coisa num único balão.
  function normalizeOpenText(text) {
    const trimmed = (text || '').trim();
    if (!trimmed) return '';
    return trimmed.charAt(0).toUpperCase() + trimmed.slice(1);
  }

  // As caixas de resposta aberta ficam visíveis desde o início (com base em
  // todo mundo que respondeu), igual às opções de múltipla escolha — só os
  // rostos dentro de cada caixa dependem de quem já foi revelado.
  $: openGroups =
    currentQuestion && currentQuestion.type === 'OPEN_TEXT'
      ? groupOpenAnswers(participants, currentQuestion, revealedIds)
      : [];

  function groupOpenAnswers(allParticipants, q, revealedSet) {
    const groups = new Map();
    for (const p of allParticipants) {
      const label = normalizeOpenText(answerFor(p, q)?.text) || '—';
      if (!groups.has(label)) groups.set(label, { label, participants: [] });
      if (revealedSet.has(p.id)) groups.get(label).participants.push(p);
    }
    return Array.from(groups.values());
  }

  $: groups = currentQuestion && currentQuestion.type === 'OPEN_TEXT' ? openGroups : choiceGroups;

  function reveal(p) {
    if (!currentQuestion) return;
    const questionId = currentQuestion.id;
    const set = revealed[questionId];
    if (set.has(p.id)) {
      set.delete(p.id);
      revealed = { ...revealed };
      api.events.live.unreveal(id, questionId, p.id).catch(() => {});
    } else {
      set.add(p.id);
      revealed = { ...revealed };
      api.events.live.reveal(id, questionId, p.id).catch(() => {});
    }
  }

  function revealAll() {
    if (!currentQuestion) return;
    const set = revealed[currentQuestion.id];
    const questionId = currentQuestion.id;
    pending.forEach((p, i) => {
      setTimeout(() => {
        set.add(p.id);
        revealed = { ...revealed };
        api.events.live.reveal(id, questionId, p.id).catch(() => {});
      }, i * 150);
    });
  }

  function resetReveal() {
    if (!currentQuestion) return;
    revealed = { ...revealed, [currentQuestion.id]: new Set() };
    api.events.live.reset(id, currentQuestion.id).catch(() => {});
  }

  function goPrev() {
    if (currentIndex > 0) {
      currentIndex -= 1;
      syncQuestion(questions[currentIndex].id);
    }
  }

  function goNext() {
    if (currentIndex < questions.length - 1) {
      currentIndex += 1;
      syncQuestion(questions[currentIndex].id);
    }
  }

  function jumpToQuestion(i) {
    if (i === currentIndex) return;
    currentIndex = i;
    syncQuestion(questions[i].id);
  }

  function questionKind(q) {
    if (q.type === 'OPEN_TEXT') return 'Resposta aberta';
    if (q.type === 'GROUP') return 'Grupo';
    return 'Individual';
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<ReactionBurstLayer />

<main class="stage-page" class:presentation-mode={presentationMode} class:stage-center={loading || error}>
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if error}
    <p class="form-error">{error}</p>
  {:else}
    {#if !presentationMode}
      <div class="stage-topbar">
        <a
          class="stage-back-link"
          href="/events/{id}"
          aria-label="Voltar para o evento"
          on:click|preventDefault={back}
        >
          <img class="stage-topbar-logo" src="/img/arandu-logo.png" alt="Arandu" />
        </a>
        <span class="stage-topbar-title">{event.title}</span>
        {#if eventStatusInfo}
          <span
            class="stage-status-badge"
            style="background:{eventStatusInfo.tint};color:{eventStatusInfo.color}"
          >
            ● {eventStatusInfo.label}
          </span>
        {/if}
        <span class="stage-topbar-spacer"></span>
        <span class="stage-pin-chip">#{event.pinCode.toUpperCase()}</span>
        <span class="stage-qa-badge">Q&amp;A {qaInbox.length}</span>
        <Button variant="secondary" size="sm" on:click={togglePresentationMode}>
          Modo apresentação
        </Button>
      </div>
    {/if}

    {#if questions.length === 0}
      <p class="text-muted stage-empty-msg">Este evento ainda não tem perguntas.</p>
    {:else}
      <div class="stage-layout">
        <div class="stage-main">
          <div class="stage-main-body">
            <h2 class="stage-question-title">{currentQuestion.title}</h2>

            <div class="stage-zones-grid">
              {#each groups as group (group.label)}
                <div class="zone stage-zone">
                  <div class="stage-zone-head">
                    <span>{group.label}</span>
                    <span class="stage-zone-count">{group.participants.length}</span>
                  </div>
                  <div class="stage-zone-faces">
                    {#each group.participants as p (p.id)}
                      <button
                        type="button"
                        class="stage-face"
                        title={p.name || p.email}
                        aria-label={`Desrevelar resposta de ${p.name || p.email}`}
                        on:click={() => reveal(p)}
                        animate:flip={{ duration: 350 }}
                        in:fly={{ y: -30, duration: 350 }}
                      >
                        {#if p.photo}
                          <img src={p.photo} alt="" />
                        {:else}
                          <span class="stage-face-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
                        {/if}
                      </button>
                    {/each}
                  </div>
                </div>
              {/each}
            </div>
          </div>

          {#if !presentationMode}
            <div class="stage-dock">
              <Button size="sm" on:click={revealAll} disabled={pending.length === 0}>
                Revelar tudo
              </Button>
              <Button variant="secondary" size="sm" on:click={resetReveal} disabled={revealedIds.size === 0}>
                Reiniciar
              </Button>
              <span class="stage-dock-spacer"></span>
              <div class="stage-dock-nav">
                <button
                  type="button"
                  class="stage-nav-btn prev"
                  aria-label="Pergunta anterior"
                  disabled={currentIndex === 0}
                  on:click={goPrev}
                >←</button>
                <span class="stage-dock-counter">{currentIndex + 1} / {questions.length}</span>
                <button
                  type="button"
                  class="stage-nav-btn next"
                  aria-label="Próxima pergunta"
                  disabled={currentIndex === questions.length - 1}
                  on:click={goNext}
                >→</button>
              </div>
              <span class="stage-dock-spacer"></span>
            </div>
          {/if}
        </div>

        {#if !presentationMode}
          <div class="stage-rail">
            <div class="stage-rail-questions">
              <div class="stage-rail-section-head">
                <span>Perguntas</span>
                <span class="text-muted">{questions.length}</span>
              </div>
              <div class="stage-rail-list">
                {#each questions as q, i (q.id)}
                  <button
                    type="button"
                    class="stage-question-row"
                    class:active={i === currentIndex}
                    on:click={() => jumpToQuestion(i)}
                  >
                    <span class="stage-question-n">{i + 1}</span>
                    <span class="stage-question-info">
                      <span class="stage-question-title-text">{q.title}</span>
                      <span class="stage-question-kind text-muted">{questionKind(q)}</span>
                    </span>
                  </button>
                {/each}
              </div>
            </div>

            <div class="stage-rail-switches">
              <div class="control-row">
                <Switch checked={blanked} on:change={toggleBlanked} />
                <span>Tela em branco</span>
              </div>
              <div class="control-row">
                <Switch checked={answersHidden} on:change={toggleAnswersHidden} />
                <span>Esconder respostas da plateia</span>
              </div>
              <div class="control-row">
                <Switch checked={interactionsEnabled} on:change={toggleInteractions} />
                <span>Interações dos participantes</span>
              </div>
            </div>

            <div class="stage-rail-qa">
              <p class="qa-inbox-title">Perguntas dos participantes</p>
              {#if qaInbox.length === 0}
                <p class="text-muted">Nenhuma mensagem ainda.</p>
              {:else}
                {#each qaInbox as m (m.id)}
                  <div class="qa-item">
                    <div class="qa-item-body">
                      <span class="qa-item-email">{m.name || m.email || 'Convidado'}</span>
                      <span class="qa-item-text">{m.text}</span>
                    </div>
                    <Button variant="secondary" size="sm" on:click={() => dismissQA(m.id)}>Dispensar</Button>
                  </div>
                {/each}
              {/if}
            </div>

            <div class="stage-rail-message">
              <Input
                label="Aviso pra tela dos participantes"
                bind:value={messageDraft}
                placeholder="Ex: Voltamos em 5 minutos"
              />
              <div class="stage-rail-message-actions">
                <Button size="sm" on:click={sendMessage} disabled={messageDraft.trim() === message}>
                  Enviar aviso
                </Button>
                <Button variant="secondary" size="sm" on:click={clearMessage} disabled={!message}>
                  Limpar
                </Button>
              </div>
            </div>

            <div class="stage-rail-footer">
              <Button variant="secondary" block on:click={openAudienceScreen}>
                Abrir tela de apresentação
              </Button>
            </div>
          </div>
        {/if}
      </div>

      {#if presentationMode}
        <p class="text-muted stage-presentation-hint">
          Esc para sair do modo apresentação · ← → para navegar · R revela todos
        </p>
      {/if}
    {/if}
  {/if}
</main>

<style>
  .stage-page {
    max-width: none;
    width: 100%;
    min-height: 100vh;
    min-height: 100dvh;
    box-sizing: border-box;
    padding: 0;
    display: flex;
    flex-direction: column;
  }

  .stage-center {
    align-items: center;
    justify-content: center;
  }

  .stage-topbar {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 14px;
    height: 56px;
    padding: 0 20px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
  }

  .stage-back-link {
    display: flex;
    line-height: 0;
    opacity: 1;
    transition: opacity 0.15s ease;
  }

  .stage-back-link:hover {
    opacity: 0.8;
  }

  .stage-topbar-logo {
    height: 24px;
    width: auto;
  }

  .stage-topbar-title {
    font-size: 0.95rem;
    font-weight: 700;
  }

  .stage-status-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 3px 10px;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .stage-topbar-spacer {
    flex: 1;
  }

  .stage-pin-chip {
    padding: 5px 12px;
    border-radius: 999px;
    background: var(--bg-input);
    font-size: 0.85rem;
    font-weight: 700;
    letter-spacing: 0.12em;
  }

  .stage-qa-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    border-radius: 999px;
    background: var(--tint-orange);
    color: var(--orange);
    font-size: 0.8rem;
    font-weight: 700;
  }

  .stage-empty-msg {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0;
  }

  .stage-layout {
    flex: 1;
    min-height: 0;
    display: flex;
  }

  .stage-main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .stage-main-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 24px 26px 22px;
    overflow-y: auto;
  }

  .stage-question-title {
    margin: 0;
    font-size: 2.2rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.1;
  }

  .presentation-mode .stage-question-title {
    margin-top: 8px;
  }

  .stage-pending-strip {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    min-height: 58px;
    padding: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 12px;
  }

  .stage-pending-empty {
    margin: 0;
  }

  .stage-pending-face {
    width: 48px;
    height: 48px;
    flex-shrink: 0;
    border-radius: 50%;
    border: 2px dashed var(--border-strong);
    padding: 0;
    overflow: hidden;
    cursor: pointer;
    background: var(--bg-input);
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-weight: 700;
    transition: border-color 0.15s ease, color 0.15s ease, transform 0.1s ease;
  }

  .stage-pending-face:hover {
    border-color: var(--accent);
    color: var(--accent);
    transform: scale(1.06);
  }

  .stage-pending-face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .stage-zones-grid {
    flex: 1;
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-auto-rows: 1fr;
    gap: 16px;
  }

  @media (max-width: 640px) {
    .stage-zones-grid {
      grid-template-columns: 1fr;
    }
  }

  .stage-zone {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px 18px;
    border-radius: 12px;
    border: 1px solid var(--border);
    background: var(--bg-input);
  }

  .stage-zone-head {
    display: flex;
    align-items: baseline;
    gap: 10px;
  }

  .stage-zone-head span:first-child {
    flex: 1;
    font-size: 1.05rem;
    font-weight: 700;
  }

  .stage-zone-count {
    font-size: 1.5rem;
    font-weight: 800;
    line-height: 1;
    color: var(--accent);
  }

  .stage-zone-faces {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    min-height: 52px;
  }

  .stage-face {
    width: 52px;
    height: 52px;
    flex-shrink: 0;
    border-radius: 50%;
    border: none;
    padding: 0;
    overflow: hidden;
    cursor: pointer;
    background: var(--accent);
    transition: transform 0.1s ease;
  }

  .stage-face:hover {
    transform: scale(1.06);
  }

  .stage-face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .stage-face-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--accent);
    color: #fff;
    font-weight: 700;
  }

  .stage-dock {
    position: relative;
    height: 64px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 0 24px;
    background: var(--bg-elev);
    border-top: 1px solid var(--border);
  }

  .stage-dock-spacer {
    flex: 1;
  }

  .stage-dock-nav {
    position: absolute;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .stage-nav-btn {
    width: 34px;
    height: 34px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    cursor: pointer;
  }

  .stage-nav-btn.prev {
    border: 1px solid var(--border-strong);
    background: transparent;
    color: var(--text-muted);
  }

  .stage-nav-btn.next {
    border: none;
    background: var(--accent);
    color: #fff;
  }

  .stage-nav-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .stage-dock-counter {
    min-width: 64px;
    text-align: center;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-muted);
  }

  .stage-presentation-hint {
    margin: 4px 0 0;
    text-align: center;
    font-size: 0.75rem;
    opacity: 0.5;
  }

  .stage-rail {
    width: 280px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    background: var(--bg-elev);
    border-left: 1px solid var(--border);
    overflow-y: auto;
  }

  @media (max-width: 900px) {
    .stage-layout {
      flex-direction: column;
    }

    .stage-rail {
      width: 100%;
      max-height: 50vh;
      border-left: none;
      border-top: 1px solid var(--border);
    }
  }

  .stage-rail-questions {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    max-height: 260px;
    padding: 16px 16px 10px;
  }

  .stage-rail-section-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 8px;
    font-size: 0.85rem;
    font-weight: 700;
  }

  .stage-rail-list {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    overflow-y: auto;
  }

  .stage-question-row {
    display: flex;
    gap: 10px;
    padding: 9px 10px;
    border-radius: 10px;
    cursor: pointer;
    background: transparent;
    border: none;
    border-left: 3px solid transparent;
    text-align: left;
    font-family: var(--font-ui);
  }

  .stage-question-row.active {
    background: var(--tint-purple);
    border-left-color: var(--accent);
  }

  .stage-question-n {
    flex-shrink: 0;
    font-size: 0.78rem;
    font-weight: 800;
    color: var(--text-muted);
  }

  .stage-question-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .stage-question-title-text {
    font-size: 0.82rem;
    line-height: 1.3;
    color: var(--text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .stage-question-row.active .stage-question-title-text {
    font-weight: 700;
    color: var(--accent);
  }

  .stage-question-kind {
    font-size: 0.7rem;
  }

  .stage-rail-switches {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 16px;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }

  .control-row {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
  }

  .stage-rail-qa {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px 16px;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }

  .qa-inbox-title {
    margin: 0;
    font-weight: 600;
    font-size: 0.85rem;
  }

  .qa-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 10px 12px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 10px;
  }

  .qa-item-body {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .qa-item-email {
    font-size: 0.78rem;
    color: var(--text-muted);
  }

  .qa-item-text {
    font-size: 0.85rem;
    word-break: break-word;
  }

  .stage-rail-message {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 16px;
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }

  .stage-rail-message-actions {
    display: flex;
    gap: 8px;
  }

  .stage-rail-footer {
    padding: 12px 16px 16px;
    flex-shrink: 0;
  }
</style>
