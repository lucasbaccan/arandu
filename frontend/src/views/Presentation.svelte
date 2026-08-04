<script>
  export let id = '';

  import { fly, fade } from 'svelte/transition';
  import { flip } from 'svelte/animate';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';

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
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function back() {
    navigate(`/events/${id}`);
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
          option: opt,
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
      const text = normalizeOpenText(answerFor(p, q)?.text) || '—';
      if (!groups.has(text)) groups.set(text, { text, participants: [] });
      if (revealedSet.has(p.id)) groups.get(text).participants.push(p);
    }
    return Array.from(groups.values());
  }

  function reveal(p) {
    if (!currentQuestion) return;
    const set = revealed[currentQuestion.id];
    if (set.has(p.id)) return;
    set.add(p.id);
    revealed = { ...revealed };
  }

  function revealAll() {
    if (!currentQuestion) return;
    const set = revealed[currentQuestion.id];
    pending.forEach((p, i) => {
      setTimeout(() => {
        set.add(p.id);
        revealed = { ...revealed };
      }, i * 150);
    });
  }

  function resetReveal() {
    if (!currentQuestion) return;
    revealed = { ...revealed, [currentQuestion.id]: new Set() };
  }

  function goPrev() {
    if (currentIndex > 0) currentIndex -= 1;
  }

  function goNext() {
    if (currentIndex < questions.length - 1) currentIndex += 1;
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

      <div class="pending-row">
        {#each pending as p (p.id)}
          <button
            type="button"
            class="face"
            title={p.email}
            aria-label={`Revelar resposta de ${p.email}`}
            on:click={() => reveal(p)}
            animate:flip={{ duration: 350 }}
            out:fade={{ duration: 150 }}
          >
            {#if p.photo}
              <img src={p.photo} alt="" />
            {:else}
              <span class="face-placeholder">{p.email[0].toUpperCase()}</span>
            {/if}
          </button>
        {/each}
        {#if pending.length === 0}
          <p class="text-muted present-empty">
            {participants.length === 0 ? 'Ninguém respondeu ainda.' : 'Todas as respostas foram reveladas.'}
          </p>
        {/if}
      </div>

      {#if currentQuestion.type === 'OPEN_TEXT'}
        <div class="zones">
          {#if openGroups.length === 0}
            <p class="text-muted present-empty">Ninguém respondeu ainda.</p>
          {/if}
          {#each openGroups as group (group.text)}
            <div class="zone">
              <div class="zone-label">
                <span>{group.text}</span>
                <span class="zone-count">{group.participants.length}</span>
              </div>
              <div class="zone-faces">
                {#each group.participants as p (p.id)}
                  <span
                    class="face static"
                    title={p.email}
                    animate:flip={{ duration: 350 }}
                    in:fly={{ y: -30, duration: 350 }}
                  >
                    {#if p.photo}
                      <img src={p.photo} alt="" />
                    {:else}
                      <span class="face-placeholder">{p.email[0].toUpperCase()}</span>
                    {/if}
                  </span>
                {/each}
              </div>
            </div>
          {/each}
        </div>
      {:else}
        <div class="zones">
          {#each choiceGroups as g (g.option.id)}
            <div class="zone">
              <div class="zone-label">
                <span>{g.option.text}</span>
                <span class="zone-count">{g.participants.length}</span>
              </div>
              <div class="zone-faces">
                {#each g.participants as p (p.id)}
                  <span
                    class="face static"
                    title={p.email}
                    animate:flip={{ duration: 350 }}
                    in:fly={{ y: -30, duration: 350 }}
                  >
                    {#if p.photo}
                      <img src={p.photo} alt="" />
                    {:else}
                      <span class="face-placeholder">{p.email[0].toUpperCase()}</span>
                    {/if}
                  </span>
                {/each}
              </div>
            </div>
          {/each}
        </div>
      {/if}

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

  .pending-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    min-height: 64px;
    padding: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 12px;
  }

  .face {
    width: 52px;
    height: 52px;
    flex-shrink: 0;
    border-radius: 50%;
    border: 2px solid var(--border);
    padding: 0;
    overflow: hidden;
    cursor: pointer;
    background: var(--bg-input);
    transition: border-color 0.15s ease, transform 0.1s ease;
  }

  .face:hover {
    border-color: var(--accent);
    transform: scale(1.05);
  }

  .face.static {
    cursor: default;
    display: block;
  }

  .face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .face-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--accent);
    color: #fff;
    font-weight: 600;
  }

  .zones {
    flex: 1;
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
  }

  .zone {
    flex: 1;
    min-width: 220px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 14px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 12px;
  }

  .zone-label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-weight: 600;
  }

  .zone-count {
    color: var(--text-muted);
    font-weight: 400;
  }

  .zone-faces {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    min-height: 52px;
  }

  .present-nav {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }
</style>
