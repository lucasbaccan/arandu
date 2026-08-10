<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Input from '../components/Input.svelte';
  import AvatarCropper from '../components/AvatarCropper.svelte';
  import CopyButton from '../components/CopyButton.svelte';

  // Mesma regra usada pelo backend (isValidEmail em server.go) — precisa ficar
  // idêntica para o erro aparecer aqui, na identificação, e não só depois de
  // responder tudo e tentar finalizar.
  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]+$/;

  const SLIDE_HEIGHT = 300;
  const SLIDE_GAP = 14;
  const SLIDE_STEP = SLIDE_HEIGHT + SLIDE_GAP;
  const PEEK = 64;
  const VIEWPORT_HEIGHT = SLIDE_HEIGHT + PEEK;

  let loading = true;
  let notFound = false;
  let error = '';

  let event = null;
  let questions = [];
  let answers = {};

  let step = 'identify'; // identify | questions | done | closed

  let email = '';
  let photo = '';
  let emailError = '';
  let photoWarning = false;

  let currentIndex = 0;
  let questionError = '';
  let submitting = false;
  let submitError = '';

  const requestedEditToken = new URLSearchParams(window.location.search).get('edit') || '';
  let isEditMode = false;
  let editToken = '';
  let editLink = '';
  let editLinkNotice = '';

  load();

  async function load() {
    try {
      const { event: ev, questions: qs } = await api.public.events.get(id);
      event = ev;
      questions = qs;
      if (!ev.answersOpen) {
        step = 'closed';
        return;
      }
      const initial = {};
      for (const q of qs) {
        initial[q.id] = { optionId: '', text: '' };
      }
      answers = initial;

      if (requestedEditToken) {
        await loadForEdit(requestedEditToken);
      }
    } catch (e) {
      if (e.status === 404) {
        notFound = true;
      } else {
        error = e.message;
      }
    } finally {
      loading = false;
    }
  }

  async function loadForEdit(token) {
    try {
      const { participant } = await api.public.events.getParticipant(id, token);
      email = participant.email;
      photo = participant.photo;
      editToken = token;
      isEditMode = true;
      const next = { ...answers };
      for (const a of participant.answers) {
        next[a.questionId] = { optionId: a.optionId, text: a.text };
      }
      answers = next;
      step = 'questions';
      currentIndex = 0;
    } catch (e) {
      editLinkNotice = 'Link de edição inválido ou expirado. Você pode responder normalmente abaixo.';
    }
  }

  function validateEmail() {
    const trimmed = email.trim();
    if (!trimmed) return 'Informe seu e-mail.';
    if (!emailRe.test(trimmed)) return 'Informe um e-mail válido.';
    return '';
  }

  function startQuestions() {
    emailError = validateEmail();
    if (emailError) return;
    if (!photo && !photoWarning) {
      photoWarning = true;
      return;
    }
    step = 'questions';
    currentIndex = 0;
  }

  function onPhotoChange(e) {
    photo = e.detail;
    if (photo) photoWarning = false;
  }

  $: currentQuestion = questions[currentIndex];
  $: isLast = currentIndex === questions.length - 1;

  function isAnswered(q) {
    const a = answers[q.id];
    return q.type === 'OPEN_TEXT' ? a.text.trim() !== '' : a.optionId !== '';
  }

  function validateCurrent() {
    if (!isAnswered(currentQuestion)) {
      return currentQuestion.type === 'OPEN_TEXT'
        ? 'Escreva uma resposta para continuar.'
        : 'Selecione uma opção para continuar.';
    }
    return '';
  }

  function goBack() {
    questionError = '';
    if (currentIndex === 0) {
      step = 'identify';
    } else {
      currentIndex -= 1;
    }
  }

  function goNext() {
    questionError = validateCurrent();
    if (questionError) return;
    currentIndex += 1;
  }

  async function finish() {
    questionError = validateCurrent();
    if (questionError) return;

    const allAnswered = questions.every((q) => isAnswered(q));
    if (!allAnswered) {
      submitError = 'Responda todas as perguntas antes de enviar.';
      return;
    }

    submitting = true;
    submitError = '';
    try {
      const { editToken: returnedToken } = await api.public.events.submit(id, {
        email: email.trim(),
        photo,
        editToken,
        answers: questions.map((q) => ({
          questionId: q.id,
          optionId: answers[q.id].optionId,
          text: answers[q.id].text.trim()
        }))
      });
      editToken = returnedToken || editToken;
      if (editToken) {
        editLink = `${window.location.origin}/answer/${id}?edit=${editToken}`;
      }
      step = 'done';
    } catch (e) {
      if (e.status === 400 && /e-mail/i.test(e.message)) {
        emailError = e.message;
        step = 'identify';
      } else {
        submitError = e.message;
      }
    } finally {
      submitting = false;
    }
  }
</script>

<main class="page">
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if notFound}
    <Card title="Evento não encontrado">
      <p class="subtitle">Verifique o link e tente novamente.</p>
    </Card>
  {:else if error}
    <Card title="Algo deu errado">
      <p class="form-error">{error}</p>
    </Card>
  {:else if step === 'closed'}
    <Card title={event.title}>
      <p class="subtitle">As respostas não estão abertas para este evento no momento.</p>
    </Card>
  {:else if step === 'done'}
    <Card title="Respostas enviadas!" wide>
      <p class="subtitle">Obrigado por participar, {email}.</p>
      {#if editLink}
        <div class="edit-link-box">
          <p class="edit-link-label">
            Guarde este link para editar ou atualizar suas respostas depois:
          </p>
          <div class="edit-link-row">
            <input class="edit-link-input" type="text" readonly value={editLink} />
            <CopyButton text={editLink} label="Copiar link de edição" />
          </div>
          <p class="text-muted edit-link-hint">
            Perdeu o link? Você pode solicitá-lo ao responsável pelo evento.
          </p>
        </div>
      {/if}
    </Card>
  {:else if questions.length === 0}
    <Card title={event.title}>
      <p class="subtitle">Este evento ainda não tem perguntas.</p>
    </Card>
  {:else if step === 'identify'}
    <Card wide>
      <img class="answer-logo" src="/img/arandu-logo.png" alt="Arandu" />
      <form class="form" novalidate on:submit|preventDefault={startQuestions}>
        {#if editLinkNotice}
          <p class="form-warning">{editLinkNotice}</p>
        {/if}
        <p class="invite-text">
          Seja bem-vindo(a), você foi convidado(a) para responder o evento<br />
          <strong>{event.title}</strong><br />
          Vamos coletar algumas informações suas para deixar a dinâmica mais legal.
        </p>
        <AvatarCropper on:change={onPhotoChange} />
        <Input
          label="E-mail"
          type="email"
          bind:value={email}
          error={emailError}
          placeholder="seu@melhor.email"
          autocomplete="email"
          required
        />
        {#if photoWarning}
          <p class="form-warning">A dinâmica não será a mesma sem sua foto.</p>
        {/if}
        <Button type="submit" block>
          {photoWarning ? 'Continuar mesmo assim' : 'Começar'}
        </Button>
      </form>
    </Card>
  {:else if step === 'questions'}
    <Card title={event.title} wide>
      {#if isEditMode}
        <p class="text-muted progress-label">Editando suas respostas anteriores</p>
      {/if}
      <p class="text-muted progress-label">Pergunta {currentIndex + 1} de {questions.length}</p>

      <div class="stepper">
        <div class="carousel-viewport" style="height: {VIEWPORT_HEIGHT}px;">
          <div
            class="carousel-track"
            style="transform: translateY(-{currentIndex * SLIDE_STEP}px);"
          >
            {#each questions as q, i (q.id)}
              <div
                class="slide"
                class:current={i === currentIndex}
                style="height: {SLIDE_HEIGHT}px;"
                aria-hidden={i !== currentIndex}
              >
                <div class="card-face">
                  <h2 class="question-title">{q.title}</h2>

                  {#if q.type === 'OPEN_TEXT'}
                    <textarea
                      class="answer-text"
                      bind:value={answers[q.id].text}
                      placeholder="Escreva sua resposta"
                      rows="4"
                      tabindex={i === currentIndex ? 0 : -1}
                    ></textarea>
                  {:else}
                    <div class="options">
                      {#each q.options as opt (opt.id)}
                        <label class="option" class:selected={answers[q.id].optionId === opt.id}>
                          <input
                            type="radio"
                            class="sr-only"
                            bind:group={answers[q.id].optionId}
                            value={opt.id}
                            tabindex={i === currentIndex ? 0 : -1}
                          />
                          <span>{opt.text}</span>
                        </label>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
          <div class="fade-overlay" aria-hidden="true"></div>
        </div>

        <div class="progress-rail" style="height: {VIEWPORT_HEIGHT}px;">
          <div
            class="progress-fill"
            style="height: {((currentIndex + 1) / questions.length) * 100}%;"
          ></div>
        </div>
      </div>

      {#if questionError}
        <p class="form-error">{questionError}</p>
      {/if}
      {#if submitError}
        <p class="form-error">{submitError}</p>
      {/if}

      <div class="wizard-actions">
        <Button variant="secondary" type="button" on:click={goBack} disabled={submitting}>
          Voltar
        </Button>
        {#if isLast}
          <Button type="button" on:click={finish} disabled={submitting}>
            {submitting ? 'Enviando…' : 'Finalizar'}
          </Button>
        {:else}
          <Button type="button" on:click={goNext}>Próxima</Button>
        {/if}
      </div>
    </Card>
  {/if}
</main>

<style>
  .answer-logo {
    display: block;
    height: 56px;
    width: auto;
    margin: 0 auto;
  }

  .invite-text {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.9rem;
    line-height: 1.5;
    text-align: center;
  }

  .invite-text strong {
    color: var(--text);
    font-size: 1rem;
  }

  .edit-link-box {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 4px;
    padding: 14px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 10px;
  }

  .edit-link-label {
    margin: 0;
    font-size: 0.85rem;
  }

  .edit-link-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .edit-link-input {
    flex: 1;
    min-width: 0;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text);
    padding: 8px 10px;
    font-size: 0.85rem;
  }

  .edit-link-hint {
    margin: 4px 0 0;
    font-size: 0.8rem;
  }

  .progress-label {
    margin: 0 0 10px;
  }

  .stepper {
    display: flex;
    align-items: stretch;
    gap: 12px;
  }

  .carousel-viewport {
    position: relative;
    flex: 1;
    min-width: 0;
    overflow: hidden;
    border-radius: 10px;
  }

  .carousel-track {
    transition: transform 0.35s ease;
  }

  .slide {
    box-sizing: border-box;
    margin-bottom: 14px;
    opacity: 0.32;
    filter: brightness(0.55) saturate(0.7);
    transition: opacity 0.35s ease, filter 0.35s ease;
  }

  .card-face {
    height: 100%;
    box-sizing: border-box;
    overflow-y: auto;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 16px;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.28);
  }

  .slide.current {
    opacity: 1;
    filter: none;
  }

  .slide.current .card-face {
    border-color: var(--accent);
  }

  .fade-overlay {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 72px;
    background: linear-gradient(to bottom, transparent, var(--bg-input));
    pointer-events: none;
  }

  .progress-rail {
    position: relative;
    width: 8px;
    flex-shrink: 0;
    border-radius: 999px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    overflow: hidden;
  }

  .progress-fill {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    background: var(--accent);
    border-radius: 999px;
    transition: height 0.3s ease;
  }

  .question-title {
    font-size: 1.1rem;
    margin: 0 0 14px;
  }

  .answer-text {
    width: 100%;
    box-sizing: border-box;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text);
    padding: 10px 12px;
    font-size: 1rem;
    font-family: inherit;
    resize: vertical;
    outline: none;
  }

  .answer-text:focus {
    border-color: var(--accent);
  }

  .options {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .option {
    display: flex;
    align-items: center;
    padding: 12px 14px;
    border: 1px solid var(--border);
    border-radius: 8px;
    cursor: pointer;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .option.selected {
    border-color: var(--accent);
    background: rgba(43, 0, 187, 0.1);
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .wizard-actions {
    display: flex;
    gap: 10px;
    margin-top: 16px;
  }
</style>
