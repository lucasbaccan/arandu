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

  const KEYS = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];

  const PRIVACY_ITEMS = [
    {
      title: 'O que pedimos',
      body: 'Seu nome, seu e-mail e, se você quiser, uma foto de rosto. Nada além disso.'
    },
    {
      title: 'O que aparece no telão',
      body: 'Seu nome, sua foto e a resposta escolhida — e só no momento em que o organizador revelar você. Antes disso, o servidor não envia sua resposta para a tela pública.'
    },
    {
      title: 'O que ninguém vê',
      body: 'Seu e-mail nunca aparece na apresentação. Ele fica visível apenas para o organizador do evento, para identificar sua participação.'
    },
    {
      title: 'A foto é opcional',
      body: 'Sem foto, você aparece com a inicial do seu nome. A dinâmica funciona do mesmo jeito.'
    },
    {
      title: 'As respostas ficam no evento',
      body: 'Elas existem para serem mostradas na apresentação deste evento e nada mais: não viram relatório, não são cruzadas com outros eventos nem compartilhadas fora dele.'
    },
    {
      title: 'Editar depois',
      body: 'Você recebe um link privado para corrigir respostas ou trocar a foto sem precisar criar conta.'
    }
  ];

  let loading = true;
  let notFound = false;
  let error = '';

  let event = null;
  let questions = [];
  let answers = {};

  let step = 'identify'; // identify | privacy | question | review | done | closed

  let name = '';
  let email = '';
  let photo = '';
  let nameError = '';
  let emailError = '';
  let photoWarning = false;

  let currentIndex = 0;
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
      const initial = {};
      for (const q of qs) {
        initial[q.id] = { optionId: '', text: '' };
      }
      answers = initial;

      if (!ev.answersOpen) {
        step = 'closed';
        return;
      }

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
      name = participant.name || '';
      photo = participant.photo;
      editToken = token;
      isEditMode = true;
      const next = { ...answers };
      for (const a of participant.answers) {
        next[a.questionId] = { optionId: a.optionId, text: a.text };
      }
      answers = next;
      step = 'question';
      currentIndex = 0;
    } catch (e) {
      editLinkNotice = 'Link de edição inválido ou expirado. Você pode responder normalmente abaixo.';
    }
  }

  function validateName() {
    if (!name.trim()) return 'Informe seu nome.';
    return '';
  }

  function validateEmail() {
    const trimmed = email.trim();
    if (!trimmed) return 'Informe seu e-mail.';
    if (!emailRe.test(trimmed)) return 'Informe um e-mail válido.';
    return '';
  }

  function startFlow() {
    nameError = validateName();
    emailError = validateEmail();
    if (nameError || emailError) return;
    if (!photo && !photoWarning) {
      photoWarning = true;
      return;
    }
    step = 'question';
    currentIndex = 0;
  }

  function onPhotoChange(e) {
    photo = e.detail;
    if (photo) photoWarning = false;
  }

  $: currentQuestion = questions[currentIndex];
  $: isLast = currentIndex === questions.length - 1;
  $: answeredCount = questions.filter((q) => isAnswered(q, answers)).length;
  $: progressPct = questions.length ? Math.round((answeredCount / questions.length) * 100) : 0;
  $: initial = (name.trim() || email.trim() || '?').charAt(0).toUpperCase();
  $: firstName = (name.trim() || 'você').split(' ')[0];

  function isAnswered(q, ans) {
    const a = ans[q.id];
    if (!a) return false;
    return q.type === 'OPEN_TEXT' ? a.text.trim() !== '' : a.optionId !== '';
  }

  function answerPreview(q, ans) {
    const a = ans[q.id];
    if (!a) return '';
    if (q.type === 'OPEN_TEXT') return a.text.trim();
    const opt = q.options.find((o) => o.id === a.optionId);
    return opt ? opt.text : '';
  }

  function goBack() {
    if (currentIndex === 0) {
      step = 'identify';
    } else {
      currentIndex -= 1;
    }
  }

  function goNext() {
    if (isLast) {
      step = 'review';
    } else {
      currentIndex += 1;
    }
  }

  function jumpTo(i) {
    currentIndex = i;
    step = 'question';
  }

  function backToLast() {
    step = 'question';
    currentIndex = questions.length - 1;
  }

  function handleKeydown(e) {
    if (step !== 'question' || e.key !== 'Enter') return;
    if (document.activeElement && document.activeElement.tagName === 'TEXTAREA') return;
    e.preventDefault();
    goNext();
  }

  async function finish() {
    const allAnswered = questions.every((q) => isAnswered(q, answers));
    if (!allAnswered) {
      submitError = 'Responda todas as perguntas antes de enviar.';
      return;
    }

    submitting = true;
    submitError = '';
    try {
      const { editToken: returnedToken } = await api.public.events.submit(id, {
        email: email.trim(),
        name: name.trim(),
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
      } else if (e.status === 400 && /nome/i.test(e.message)) {
        nameError = e.message;
        step = 'identify';
      } else {
        submitError = e.message;
      }
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<main class="answer-root" class:centered={loading || notFound || error || step === 'closed' || (questions.length === 0 && !loading)}>
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
  {:else if questions.length === 0}
    <Card title={event.title}>
      <p class="subtitle">Este evento ainda não tem perguntas.</p>
    </Card>
  {:else if step === 'identify'}
    <div class="identify-shell">
      <div class="identify-info">
        <img class="identify-logo" src="/img/arandu-completo.png" alt="Arandu" />
        <h1 class="identify-title">{event.title}</h1>
        <p class="identify-desc">
          São {questions.length} perguntas rápidas. Sua foto e seu nome aparecem no telão quando o
          organizador revelar as respostas.
        </p>
        <div class="identify-meta">
          <span class="text-muted">≈ 2 minutos</span>
          <span class="text-muted">Dá para editar depois</span>
          <button type="button" class="link-btn" on:click={() => (step = 'privacy')}>
            Privacidade e uso dos dados
          </button>
        </div>
      </div>

      <form class="identify-card" novalidate on:submit|preventDefault={startFlow}>
        {#if editLinkNotice}
          <p class="form-warning">{editLinkNotice}</p>
        {/if}
        <AvatarCropper on:change={onPhotoChange} />
        <Input
          label="Como quer aparecer"
          bind:value={name}
          error={nameError}
          placeholder="Seu nome"
          autocomplete="name"
          required
        />
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
        <Button type="submit" block size="lg">
          {photoWarning ? 'Continuar mesmo assim' : 'Começar'}
        </Button>
      </form>
    </div>
  {:else if step === 'privacy'}
    <div class="privacy-shell">
      <div class="privacy-header">
        <button
          type="button"
          class="icon-btn"
          aria-label="Voltar"
          on:click={() => (step = 'identify')}
        >‹</button>
        <img class="privacy-logo" src="/img/arandu-logo.png" alt="Arandu" />
        <span class="privacy-header-title">Privacidade e uso dos dados</span>
      </div>
      <div class="privacy-body">
        <div class="privacy-intro">
          <h1>Seus dados servem para uma coisa só: a dinâmica</h1>
          <p class="text-muted">
            O Arandu existe para o grupo se conhecer. Nada do que você envia aqui vira anúncio,
            ranking ou lista de contatos.
          </p>
        </div>
        <div class="privacy-items">
          {#each PRIVACY_ITEMS as p (p.title)}
            <div class="privacy-item">
              <span class="privacy-item-title">{p.title}</span>
              <span class="text-muted">{p.body}</span>
            </div>
          {/each}
          <div class="privacy-item">
            <span class="privacy-item-title">Quanto tempo fica guardado</span>
            <span class="text-muted">
              Os dados ficam no evento enquanto ele existir. Quando o organizador apaga o evento,
              as respostas e fotos vão junto.
            </span>
          </div>
        </div>
        <p class="privacy-note text-muted">
          Dúvidas ou pedido de remoção: fale com quem organiza o evento. O Arandu roda na
          infraestrutura de quem hospeda a plataforma; não enviamos seus dados para serviços de
          terceiros.
        </p>
        <Button type="button" on:click={() => (step = 'identify')}>Entendi, continuar</Button>
      </div>
    </div>
  {:else if step === 'question'}
    <div class="question-shell">
      <aside class="q-sidebar">
        <div class="q-sidebar-head">
          <span class="q-avatar">
            {#if photo}
              <img src={photo} alt="" />
            {:else}
              {initial}
            {/if}
          </span>
          <div class="q-sidebar-id">
            <span class="q-sidebar-name">{name || 'Sem nome'}</span>
            <span class="q-sidebar-email text-muted">{email}</span>
          </div>
          <button
            type="button"
            class="icon-btn"
            title="Editar seus dados"
            aria-label="Editar seus dados"
            on:click={() => (step = 'identify')}
          >✎</button>
        </div>
        <nav class="q-sidebar-nav">
          {#each questions as q, i (q.id)}
            <button type="button" class="q-nav-item" class:current={i === currentIndex} on:click={() => jumpTo(i)}>
              <span class="q-nav-dot" class:done={isAnswered(q, answers)}>
                {isAnswered(q, answers) ? '✓' : i + 1}
              </span>
              <span class="q-nav-text">
                <span class="q-nav-title">{q.title}</span>
                <span class="q-nav-answer text-muted">{answerPreview(q, answers) || 'Sem resposta'}</span>
              </span>
            </button>
          {/each}
        </nav>
        <div class="q-sidebar-progress">
          <div class="q-progress-row">
            <span class="text-muted">{answeredCount} de {questions.length} respondidas</span>
            <strong>{progressPct}%</strong>
          </div>
          <div class="q-progress-rail">
            <div class="q-progress-fill" style="width: {progressPct}%;"></div>
          </div>
        </div>
      </aside>

      <div class="q-topbar">
        <span class="q-topbar-count text-muted">{currentIndex + 1} de {questions.length}</span>
        <button
          type="button"
          class="q-avatar q-topbar-avatar"
          aria-label="Editar seus dados"
          on:click={() => (step = 'identify')}
        >
          {#if photo}
            <img src={photo} alt="" />
          {:else}
            {initial}
          {/if}
        </button>
      </div>
      <div class="q-topbar-segments">
        {#each questions as q, i (q.id)}
          <span class="q-segment" class:current={i === currentIndex} class:done={isAnswered(q, answers)}></span>
        {/each}
      </div>

      <div class="q-main">
        <div class="q-main-scroll">
          {#if isEditMode}
            <p class="text-muted">Editando suas respostas anteriores</p>
          {/if}
          <span class="q-desktop-count text-muted">Pergunta {currentIndex + 1} de {questions.length}</span>
          <h1 class="q-title">{currentQuestion.title}</h1>

          {#if currentQuestion.type === 'OPEN_TEXT'}
            <div class="q-text-wrap">
              <textarea
                class="q-textarea"
                bind:value={answers[currentQuestion.id].text}
                placeholder="Escreva sua resposta"
                rows="5"
              ></textarea>
              <span class="q-text-count text-muted">
                {answers[currentQuestion.id].text.length} caracteres
              </span>
            </div>
          {:else}
            <div class="q-options">
              {#each currentQuestion.options as opt, i (opt.id)}
                <label class="q-option" class:selected={answers[currentQuestion.id].optionId === opt.id}>
                  <input
                    type="radio"
                    class="sr-only"
                    aria-label={opt.text}
                    bind:group={answers[currentQuestion.id].optionId}
                    value={opt.id}
                  />
                  <span class="q-option-key">{KEYS[i]}</span>
                  <span class="q-option-text">{opt.text}</span>
                </label>
              {/each}
            </div>
          {/if}
        </div>

        <div class="q-dock">
          <Button variant="secondary" type="button" on:click={goBack}>Voltar</Button>
          <span class="q-dock-spacer"></span>
          <span class="q-dock-hint text-muted">Enter avança</span>
          <Button type="button" on:click={goNext}>{isLast ? 'Revisar' : 'Próxima'}</Button>
        </div>
      </div>
    </div>
  {:else if step === 'review'}
    <div class="review-shell">
      <div class="review-scroll">
        <div class="review-header">
          <h1>Confira antes de enviar</h1>
          <p class="text-muted">
            Você pode alterar qualquer resposta agora — ou depois, pelo seu link privado.
          </p>
        </div>
        <button type="button" class="review-identity" on:click={() => (step = 'identify')}>
          <span class="q-avatar">
            {#if photo}
              <img src={photo} alt="" />
            {:else}
              {initial}
            {/if}
          </span>
          <span class="review-identity-text">
            <span class="review-identity-name">{name || 'Sem nome'}</span>
            <span class="text-muted">{email}</span>
          </span>
          <span class="review-edit-label">Editar</span>
        </button>
        <div class="review-list">
          {#each questions as q, i (q.id)}
            <button type="button" class="review-row" on:click={() => jumpTo(i)}>
              <span class="review-row-n text-muted">{i + 1}</span>
              <span class="review-row-text">
                <span class="text-muted review-row-q">{q.title}</span>
                <span class="review-row-a" class:unanswered={!isAnswered(q, answers)}>
                  {answerPreview(q, answers) || 'Sem resposta'}
                </span>
              </span>
              <span class="review-row-chevron text-muted">›</span>
            </button>
          {/each}
        </div>
      </div>
      {#if submitError}
        <p class="form-error review-error">{submitError}</p>
      {/if}
      <div class="review-dock">
        <Button variant="secondary" type="button" on:click={backToLast} disabled={submitting}>
          Voltar
        </Button>
        <span class="q-dock-spacer"></span>
        <span class="text-muted review-dock-count">{answeredCount} de {questions.length} respondidas</span>
        <Button type="button" on:click={finish} disabled={submitting || answeredCount < questions.length}>
          {submitting ? 'Enviando…' : 'Enviar respostas'}
        </Button>
      </div>
    </div>
  {:else if step === 'done'}
    <div class="done-shell">
      <div class="done-card">
        <img class="done-logo" src="/img/arandu-completo.png" alt="Arandu" />
        <span class="q-avatar done-avatar">
          {#if photo}
            <img src={photo} alt="" />
          {:else}
            {initial}
          {/if}
        </span>
        <h1 class="done-title">Prontinho, {firstName}!</h1>
        <p class="text-muted done-sub">
          Suas respostas ficam escondidas até o organizador revelar você no telão, no dia do
          evento.
        </p>
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
        <p class="text-muted done-hint">Pode fechar esta página. Até lá!</p>
      </div>
    </div>
  {/if}
</main>

<style>
  .answer-root {
    flex: 1;
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  .answer-root.centered {
    align-items: center;
    justify-content: center;
    padding: 24px;
    gap: 16px;
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

  .link-btn {
    padding: 0;
    border: none;
    background: transparent;
    color: var(--accent);
    font-family: var(--font-ui);
    font-size: 0.85rem;
    font-weight: 600;
    text-decoration: underline;
    cursor: pointer;
  }

  /* Avatar reutilizado na sidebar, no topbar mobile, na revisão e na tela final */
  .q-avatar {
    width: 44px;
    height: 44px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--accent);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    font-weight: 800;
    overflow: hidden;
    border: none;
    padding: 0;
    cursor: default;
  }

  .q-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  /* --- Identificação --- */

  .identify-shell {
    flex: 1;
    width: 100%;
    max-width: 1000px;
    margin: 0 auto;
    padding: 48px 32px;
    display: flex;
    align-items: center;
    gap: 56px;
  }

  .identify-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .identify-logo {
    height: 200px;
    width: auto;
    align-self: center;
  }

  .identify-title {
    margin: 0;
    font-size: 2rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.15;
  }

  .identify-desc {
    margin: 0;
    color: var(--text-muted);
    font-size: 1rem;
    line-height: 1.5;
  }

  .identify-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 20px;
    margin-top: 4px;
    font-size: 0.85rem;
  }

  .identify-card {
    width: 400px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 32px 28px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
  }

  /* --- Privacidade --- */

  .privacy-shell {
    flex: 1;
    width: 100%;
    display: flex;
    flex-direction: column;
  }

  .privacy-header {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 18px 32px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
  }

  .privacy-logo {
    height: 26px;
    width: auto;
  }

  .privacy-header-title {
    font-size: 0.95rem;
    font-weight: 700;
  }

  .privacy-body {
    flex: 1;
    display: flex;
    justify-content: center;
    padding: 40px 32px 56px;
  }

  .privacy-intro,
  .privacy-items,
  .privacy-note {
    width: 100%;
  }

  .privacy-body > * {
    max-width: 760px;
  }

  .privacy-body {
    flex-direction: column;
    align-items: center;
  }

  .privacy-intro {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 22px;
  }

  .privacy-intro h1 {
    margin: 0;
    font-size: 1.7rem;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .privacy-intro p {
    margin: 0;
    font-size: 1rem;
    line-height: 1.5;
  }

  .privacy-items {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 20px;
  }

  .privacy-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 16px 18px;
    border-radius: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
  }

  .privacy-item-title {
    font-size: 0.95rem;
    font-weight: 700;
  }

  .privacy-note {
    margin: 0 0 20px;
    padding: 14px 16px;
    border-radius: 12px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    font-size: 0.85rem;
    line-height: 1.5;
  }

  /* --- Perguntas --- */

  .question-shell {
    flex: 1;
    width: 100%;
    display: flex;
    min-height: 0;
  }

  .q-sidebar {
    width: 300px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    background: var(--bg-elev);
    border-right: 1px solid var(--border);
  }

  .q-sidebar-head {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 18px;
    border-bottom: 1px solid var(--border);
  }

  .q-sidebar-id {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  .q-sidebar-name {
    font-size: 0.9rem;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .q-sidebar-email {
    font-size: 0.72rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .q-sidebar-nav {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 12px;
    overflow-y: auto;
  }

  .q-nav-item {
    flex-shrink: 0;
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 10px;
    text-align: left;
    cursor: pointer;
    font-family: var(--font-ui);
    background: transparent;
    border: 1px solid transparent;
    color: var(--text-muted);
    transition: background 0.15s ease, border-color 0.15s ease;
  }

  .q-nav-item.current {
    background: rgba(43, 0, 187, 0.08);
    border-color: var(--accent);
    color: var(--text);
  }

  .q-nav-item:hover {
    background: var(--bg-input);
  }

  .q-nav-dot {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
    margin-top: 1px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.68rem;
    font-weight: 800;
    background: var(--bg-input);
    color: var(--text-muted);
  }

  .q-nav-dot.done {
    background: var(--accent);
    color: #fff;
  }

  .q-nav-text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .q-nav-title {
    font-size: 0.82rem;
    line-height: 1.3;
  }

  .q-nav-answer {
    font-size: 0.7rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .q-sidebar-progress {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 14px 16px;
    border-top: 1px solid var(--border);
  }

  .q-progress-row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    font-size: 0.78rem;
  }

  .q-progress-rail {
    height: 6px;
    border-radius: 999px;
    background: var(--bg-input);
    overflow: hidden;
  }

  .q-progress-fill {
    height: 100%;
    border-radius: 999px;
    background: var(--accent);
    transition: width 0.3s ease;
  }

  .q-topbar,
  .q-topbar-segments {
    display: none;
  }

  .q-main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .q-main-scroll {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding: 44px 56px 32px;
    overflow-y: auto;
  }

  .q-desktop-count {
    font-size: 0.85rem;
    font-weight: 600;
  }

  .q-title {
    margin: 0;
    font-size: 2.2rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.15;
  }

  .q-options {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
  }

  .q-option {
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 76px;
    padding: 16px 20px;
    border-radius: 12px;
    cursor: pointer;
    font-size: 1.05rem;
    font-weight: 600;
    background: var(--bg-elev);
    border: 2px solid var(--border);
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .q-option.selected {
    border-color: var(--accent);
    background: rgba(43, 0, 187, 0.08);
  }

  .q-option-key {
    width: 30px;
    height: 30px;
    flex-shrink: 0;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.8rem;
    font-weight: 800;
    background: var(--bg-input);
    color: var(--text-muted);
  }

  .q-option.selected .q-option-key {
    background: var(--accent);
    color: #fff;
  }

  .q-text-wrap {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .q-textarea {
    width: 100%;
    box-sizing: border-box;
    padding: 18px 20px;
    border-radius: 12px;
    border: 1px solid var(--border-strong);
    background: var(--bg-input);
    color: var(--text);
    font-family: inherit;
    font-size: 1.1rem;
    line-height: 1.5;
    resize: vertical;
    outline: none;
  }

  .q-textarea:focus {
    border-color: var(--accent);
  }

  .q-text-count {
    align-self: flex-end;
    font-size: 0.8rem;
  }

  .q-dock,
  .review-dock {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px 56px;
    background: var(--bg-elev);
    border-top: 1px solid var(--border);
  }

  .q-dock-spacer {
    flex: 1;
  }

  .q-dock-hint {
    font-size: 0.8rem;
  }

  /* --- Revisão --- */

  .review-shell {
    flex: 1;
    width: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .review-scroll {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
    padding: 44px 32px 32px;
    overflow-y: auto;
  }

  .review-scroll > * {
    width: 100%;
    max-width: 720px;
  }

  .review-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .review-header h1 {
    margin: 0;
    font-size: 1.9rem;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .review-header p {
    margin: 0;
    font-size: 0.95rem;
  }

  .review-identity {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 16px 18px;
    border-radius: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    cursor: pointer;
    text-align: left;
    font-family: var(--font-ui);
  }

  .review-identity-text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .review-identity-name {
    font-size: 0.95rem;
    font-weight: 700;
  }

  .review-edit-label {
    flex-shrink: 0;
    padding: 8px 14px;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    font-size: 0.82rem;
    font-weight: 600;
  }

  .review-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .review-row {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 14px 18px;
    border-radius: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    cursor: pointer;
    text-align: left;
    font-family: var(--font-ui);
  }

  .review-row-n {
    width: 24px;
    height: 24px;
    flex-shrink: 0;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.72rem;
    font-weight: 800;
    background: var(--bg-input);
  }

  .review-row-text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .review-row-q {
    font-size: 0.8rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .review-row-a {
    font-size: 0.95rem;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .review-row-a.unanswered {
    color: var(--danger);
  }

  .review-error {
    padding: 0 32px;
  }

  .review-dock-count {
    font-size: 0.85rem;
  }

  /* --- Concluído --- */

  .done-shell {
    flex: 1;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px 24px;
  }

  .done-card {
    width: 520px;
    max-width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    padding: 36px 34px 40px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    text-align: center;
  }

  .done-logo {
    height: 100px;
    width: auto;
  }

  .done-avatar {
    width: 72px;
    height: 72px;
    font-size: 1.5rem;
  }

  .done-title {
    margin: 0;
    font-size: 1.6rem;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .done-sub {
    margin: 0;
    font-size: 0.95rem;
    line-height: 1.45;
  }

  .done-hint {
    margin: 0;
    font-size: 0.85rem;
  }

  .edit-link-box {
    width: 100%;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 6px;
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

  /* --- Responsivo: empilha em telas estreitas --- */

  @media (max-width: 860px) {
    .identify-shell {
      flex-direction: column;
      align-items: stretch;
      gap: 28px;
      padding: 32px 20px;
    }

    .identify-logo {
      height: 72px;
      align-self: center;
    }

    .identify-title {
      font-size: 1.6rem;
      text-align: center;
    }

    .identify-desc {
      text-align: center;
    }

    .identify-meta {
      justify-content: center;
    }

    .identify-card {
      width: 100%;
      box-sizing: border-box;
      padding: 24px 20px;
    }

    .privacy-header {
      padding: 16px 20px;
    }

    .privacy-body {
      padding: 24px 20px 40px;
    }

    .q-sidebar {
      display: none;
    }

    .q-topbar {
      flex-shrink: 0;
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 16px 20px;
      background: var(--bg-elev);
      border-bottom: 1px solid var(--border);
    }

    .q-topbar-count {
      flex: 1;
      text-align: center;
      font-size: 0.8rem;
      font-weight: 600;
    }

    .q-topbar-avatar {
      cursor: pointer;
      width: 34px;
      height: 34px;
      font-size: 0.8rem;
    }

    .q-topbar-segments {
      display: flex;
      gap: 4px;
      padding: 10px 20px 0;
    }

    .q-segment {
      flex: 1;
      height: 4px;
      border-radius: 999px;
      background: var(--bg-input);
    }

    .q-segment.current {
      background: var(--accent);
    }

    .q-segment.done:not(.current) {
      background: var(--accent-hover);
    }

    .q-main-scroll {
      padding: 20px 20px 16px;
    }

    .q-title {
      font-size: 1.6rem;
    }

    .q-options {
      grid-template-columns: 1fr;
    }

    .q-dock,
    .review-dock {
      padding: 14px 20px 26px;
    }

    .q-dock-hint {
      display: none;
    }

    .review-scroll {
      padding: 20px 20px 16px;
    }

    .review-error {
      padding: 0 20px;
    }

    .review-dock-count {
      display: none;
    }
  }
</style>
