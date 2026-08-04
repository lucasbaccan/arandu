<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import PresentationStage from '../components/PresentationStage.svelte';

  let loading = true;
  let error = '';

  let event = null;
  let questions = [];
  let participants = [];
  let currentIndex = 0;

  // Modo apresentação: some com os controles e números de admin (contagens,
  // pergunta X de Y, botões) pra o organizador poder compartilhar a tela
  // (ex: Google Meet) mostrando só os rostos e as respostas. Navegação e
  // ações continuam disponíveis pelo teclado.
  let presentationMode = false;

  function togglePresentationMode() {
    presentationMode = !presentationMode;
  }

  function handleKeydown(e) {
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
  // /live/:id ver a mesma coisa em tempo real.
  function syncQuestion(questionId) {
    api.events.live.setQuestion(id, questionId).catch(() => {});
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
    const set = revealed[currentQuestion.id];
    if (set.has(p.id)) return;
    set.add(p.id);
    revealed = { ...revealed };
    api.events.live.reveal(id, currentQuestion.id, p.id).catch(() => {});
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
</script>

<svelte:window on:keydown={handleKeydown} />

<main class="present-page" class:presentation-mode={presentationMode}>
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if error}
    <p class="form-error">{error}</p>
  {:else}
    {#if !presentationMode}
      <div class="present-bar">
        <Button variant="secondary" on:click={back}>Sair do preview</Button>
        <h1 class="present-title">{event.title}</h1>
        <span class="text-muted">
          {participants.length}
          {participants.length === 1 ? 'pessoa respondeu' : 'pessoas responderam'}
        </span>
      </div>
    {/if}

    {#if questions.length === 0}
      <p class="text-muted present-empty">Este evento ainda não tem perguntas.</p>
    {:else}
      {#if !presentationMode}
        <p class="text-muted present-counter">Pergunta {currentIndex + 1} de {questions.length}</p>
      {/if}
      <h2 class="present-question">{currentQuestion.title}</h2>

      {#if !presentationMode}
        <div class="present-actions">
          <Button variant="secondary" on:click={revealAll} disabled={pending.length === 0}>
            Revelar todos
          </Button>
          <Button variant="secondary" on:click={resetReveal} disabled={revealedIds.size === 0}>
            Reiniciar revelação
          </Button>
          <Button variant="secondary" on:click={togglePresentationMode}>
            Modo apresentação
          </Button>
        </div>
      {/if}

      <PresentationStage {pending} {groups} onFaceClick={reveal} />

      {#if !presentationMode}
        <div class="present-nav">
          <Button variant="secondary" on:click={goPrev} disabled={currentIndex === 0}>
            Anterior
          </Button>
          <Button on:click={goNext} disabled={currentIndex === questions.length - 1}>
            Próxima
          </Button>
        </div>
      {:else}
        <p class="text-muted presentation-hint">
          Esc para sair do modo apresentação · ← → para navegar · R revela todos
        </p>
      {/if}
    {/if}
  {/if}
</main>

<style>
  .present-page {
    max-width: none;
    width: 100%;
    min-height: 100vh;
    box-sizing: border-box;
    padding: 24px 32px 32px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .present-bar {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .present-title {
    flex: 1;
    margin: 0;
    font-size: 1.2rem;
  }

  .present-counter {
    margin: 0;
  }

  .present-question {
    margin: 0 0 4px;
    font-size: 1.8rem;
  }

  .presentation-mode .present-question {
    margin-top: 8px;
  }

  .presentation-hint {
    margin: 4px 0 0;
    text-align: center;
    font-size: 0.75rem;
    opacity: 0.5;
  }

  .present-empty {
    margin: 0;
  }

  .present-actions {
    display: flex;
    gap: 10px;
  }

  .present-nav {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }
</style>
