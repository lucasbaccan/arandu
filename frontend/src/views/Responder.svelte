<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Input from '../components/Input.svelte';
  import AvatarCropper from '../components/AvatarCropper.svelte';
  import CopyButton from '../components/CopyButton.svelte';
  import PublicShell from '../components/PublicShell.svelte';
  import TopBar from '../components/TopBar.svelte';

  // Mesma regra usada pelo backend (isValidEmail em server.go) — precisa ficar
  // idêntica para o erro aparecer aqui, na identificação, e não só depois de
  // responder tudo e tentar finalizar.
  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]+$/;

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

  let step = 'identify'; // identify | privacidade | question | review | done | closed

  let name = '';
  let email = '';
  let photo = '';
  // Em modo de edição, o servidor devolve a foto como URL de arquivo
  // (/api/fotos/{id}), não como base64. photoChanged distingue "mantive a
  // foto" (envia vazio pra conservar a do banco) de "troquei" (envia o novo
  // data URL do AvatarCropper).
  let photoChanged = false;
  let nameError = '';
  let emailError = '';
  let photoWarning = false;

  let currentIndex = 0;
  let submitting = false;
  let submitError = '';

  const requestedEditToken = new URLSearchParams(window.location.search).get('edit') || '';
  let isEditMode = false;
  let editToken = '';
  let linkEdicao = '';
  let avisoLinkEdicao = '';
  let closedMessage = 'As respostas não estão abertas para este evento no momento.';

  load();

  async function load() {
    try {
      const { event: ev, questions: qs } = await api.publico.eventos.buscar(id);
      event = ev;
      questions = qs;
      const initialAnswers = {};
      for (const q of qs) {
        initialAnswers[q.id] = { optionId: '', text: '' };
      }
      answers = initialAnswers;

      if (!ev.answersOpen) {
        step = 'closed';
        return;
      }

      if (requestedEditToken) {
        if (!ev.allowEdit) {
          closedMessage = 'A edição de respostas está desabilitada para este evento. Fale com o organizador se precisar corrigir algo.';
          step = 'closed';
          return;
        }
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
      const { participant } = await api.publico.eventos.buscarParticipante(id, token);
      email = participant.email;
      name = participant.name || '';
      photo = participant.photo;
      photoChanged = false;
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
      avisoLinkEdicao = 'Link de edição inválido ou expirado. Você pode responder normalmente abaixo.';
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
    photoChanged = true;
    if (photo) photoWarning = false;
  }

  $: currentQuestion = questions[currentIndex];
  $: isLast = currentIndex === questions.length - 1;
  $: answeredCount = questions.filter((q) => isAnswered(q, answers)).length;
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
      const { editToken: returnedToken } = await api.publico.eventos.enviar(id, {
        email: email.trim(),
        name: name.trim(),
        photo: photoChanged ? photo : '',
        editToken,
        answers: questions.map((q) => ({
          questionId: q.id,
          optionId: answers[q.id].optionId,
          text: answers[q.id].text.trim()
        }))
      });
      editToken = returnedToken || editToken;
      if (editToken) {
        linkEdicao = `${window.location.origin}/responder/${id}?edit=${editToken}`;
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

<main class="responder-root">
  {#if loading}
    <div class="responder-center"><p class="text-muted">Carregando…</p></div>
  {:else if notFound}
    <div class="responder-center">
      <Card title="Evento não encontrado">
        <p class="subtitle">Verifique o link e tente novamente.</p>
      </Card>
    </div>
  {:else if error}
    <div class="responder-center">
      <Card title="Algo deu errado">
        <p class="form-error">{error}</p>
      </Card>
    </div>
  {:else if step === 'closed'}
    <div class="responder-center">
      <Card title={event.title}>
        <p class="subtitle">{closedMessage}</p>
      </Card>
    </div>
  {:else if questions.length === 0}
    <div class="responder-center">
      <Card title={event.title}>
        <p class="subtitle">Este evento ainda não tem perguntas.</p>
      </Card>
    </div>
  {:else if step === 'identify'}
    <PublicShell>
      <div class="identify-head">
        <span class="chip event-chip">{event.title}</span>
        <h1 class="entry-title">
          {questions.length} pergunta{questions.length === 1 ? '' : 's'} rápida{questions.length === 1 ? '' : 's'}
        </h1>
        <p class="entry-sub">
          Sua foto e seu nome só aparecem no telão quando o organizador revelar as respostas.
        </p>
      </div>

      <form class="card identify-card" novalidate on:submit|preventDefault={startFlow}>
        {#if avisoLinkEdicao}
          <p class="form-warning">{avisoLinkEdicao}</p>
        {/if}
        <AvatarCropper compact on:change={onPhotoChange} />
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
        <Button type="submit" block>
          {photoWarning ? 'Continuar mesmo assim' : 'Começar'}
        </Button>
      </form>

      <div class="identify-meta">
        <span>≈ 2 minutos</span>
        <span class="meta-dot" aria-hidden="true"></span>
        <span>Dá para editar depois</span>
        <span class="meta-dot" aria-hidden="true"></span>
        <button type="button" class="link-btn" on:click={() => (step = 'privacidade')}>
          Privacidade e uso dos dados
        </button>
      </div>
    </PublicShell>
  {:else if step === 'privacidade'}
    <div class="shell">
      <TopBar area={event.title} showAccount={false} />
      <div class="crumb-simple">
        <button type="button" class="link-btn" on:click={() => (step = 'identify')}>‹ Voltar</button>
        <span class="crumb-simple-title">Privacidade e uso dos dados</span>
      </div>
      <div class="privacidade-body">
        <div class="privacidade-col">
          <div class="privacidade-intro">
            <h1 class="section-title">Seus dados servem para uma coisa só: a dinâmica</h1>
            <p class="section-sub">
              O Arandu existe para o grupo se conhecer. Nada do que você envia aqui vira anúncio,
              ranking ou lista de contatos.
            </p>
          </div>
          <div class="privacidade-items">
            {#each PRIVACY_ITEMS as p (p.title)}
              <div class="privacidade-item">
                <span class="privacidade-item-title">{p.title}</span>
                <span class="text-muted">{p.body}</span>
              </div>
            {/each}
            <div class="privacidade-item">
              <span class="privacidade-item-title">Quanto tempo fica guardado</span>
              <span class="text-muted">
                Os dados ficam no evento enquanto ele existir. Quando o organizador apaga o evento,
                as respostas e fotos vão junto.
              </span>
            </div>
          </div>
          <p class="privacidade-note text-muted">
            Dúvidas ou pedido de remoção: fale com quem organiza o evento. O Arandu roda na
            infraestrutura de quem hospeda a plataforma; não enviamos seus dados para serviços de
            terceiros.
          </p>
          <Button type="button" on:click={() => (step = 'identify')}>Entendi, continuar</Button>
        </div>
      </div>
    </div>
  {:else if step === 'question'}
    <div class="shell">
      <TopBar area={event.title} showAccount={false}>
        <button
          slot="account"
          type="button"
          class="me-pill"
          aria-label="Editar seus dados"
          on:click={() => (step = 'identify')}
        >
          <span class="me-avatar">
            {#if photo}<img src={photo} alt="" />{:else}{initial}{/if}
          </span>
          <span class="me-name">{(name.trim() || email).split(' ')[0]}</span>
        </button>
      </TopBar>

      <div class="progress-bar">
        <span class="progress-count">
          Pergunta {currentIndex + 1} <span class="text-subtle">de {questions.length}</span>
        </span>
        <div class="segments">
          {#each questions as q, i (q.id)}
            <button
              type="button"
              class="segment"
              class:current={i === currentIndex}
              class:done={isAnswered(q, answers)}
              aria-label={`Ir para a pergunta ${i + 1}`}
              on:click={() => jumpTo(i)}
            ></button>
          {/each}
        </div>
        <span class="progress-done text-muted">{answeredCount} respondidas</span>
      </div>

      <div class="q-main">
        <div class="q-col">
          {#if isEditMode}
            <p class="text-muted q-edit-note">Editando suas respostas anteriores</p>
          {/if}
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
              {#each currentQuestion.options as opt (opt.id)}
                <label class="q-option" class:selected={answers[currentQuestion.id].optionId === opt.id}>
                  <input
                    type="radio"
                    class="sr-only"
                    aria-label={opt.text}
                    bind:group={answers[currentQuestion.id].optionId}
                    value={opt.id}
                  />
                  <span class="q-option-text">{opt.text}</span>
                  <span class="q-option-check" aria-hidden="true">
                    {answers[currentQuestion.id].optionId === opt.id ? '✓' : ''}
                  </span>
                </label>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <div class="dock">
        <div class="dock-col">
          <Button variant="secondary" type="button" on:click={goBack}>Voltar</Button>
          <span class="dock-spacer"></span>
          <span class="dock-hint text-subtle">Enter avança</span>
          <Button type="button" on:click={goNext}>{isLast ? 'Revisar' : 'Próxima'}</Button>
        </div>
      </div>
    </div>
  {:else if step === 'review'}
    <div class="shell">
      <TopBar area={event.title} showAccount={false}>
        <button
          slot="account"
          type="button"
          class="me-pill"
          aria-label="Editar seus dados"
          on:click={() => (step = 'identify')}
        >
          <span class="me-avatar">
            {#if photo}<img src={photo} alt="" />{:else}{initial}{/if}
          </span>
          <span class="me-name">{(name.trim() || email).split(' ')[0]}</span>
        </button>
      </TopBar>

      <div class="progress-bar">
        <span class="progress-count">Revisão</span>
        <div class="segments">
          {#each questions as q, i (q.id)}
            <span class="segment" class:done={isAnswered(q, answers)}></span>
          {/each}
        </div>
        <span class="progress-done text-muted">{answeredCount} de {questions.length} respondidas</span>
      </div>

      <div class="q-main">
        <div class="q-col">
          <div class="review-header">
            <h1 class="section-title">Confira antes de enviar</h1>
            <p class="section-sub">
              Você pode alterar qualquer resposta agora — ou depois, pelo seu link privado.
            </p>
          </div>
          <button type="button" class="review-identity" on:click={() => (step = 'identify')}>
            <span class="review-avatar">
              {#if photo}<img src={photo} alt="" />{:else}{initial}{/if}
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
                <span class="review-row-n">{i + 1}</span>
                <span class="review-row-text">
                  <span class="text-muted review-row-q">{q.title}</span>
                  <span class="review-row-a" class:unanswered={!isAnswered(q, answers)}>
                    {answerPreview(q, answers) || 'Sem resposta'}
                  </span>
                </span>
                <span class="review-row-chevron text-subtle">›</span>
              </button>
            {/each}
          </div>
          {#if submitError}
            <p class="form-error">{submitError}</p>
          {/if}
        </div>
      </div>

      <div class="dock">
        <div class="dock-col">
          <Button variant="secondary" type="button" on:click={backToLast} disabled={submitting}>
            Voltar
          </Button>
          <span class="dock-spacer"></span>
          <Button
            type="button"
            on:click={finish}
            disabled={submitting || answeredCount < questions.length}
          >
            {submitting ? 'Enviando…' : 'Enviar respostas'}
          </Button>
        </div>
      </div>
    </div>
  {:else if step === 'done'}
    <PublicShell>
      <div class="card done-card">
        <span class="done-avatar">
          {#if photo}<img src={photo} alt="" />{:else}{initial}{/if}
        </span>
        <h1 class="done-title">Prontinho, {firstName}!</h1>
        <p class="text-muted done-sub">
          Suas respostas ficam escondidas até o organizador revelar você no telão, no dia do evento.
        </p>
        {#if linkEdicao}
          <div class="edit-link-box">
            <p class="edit-link-label">
              Guarde este link para editar ou atualizar suas respostas depois:
            </p>
            <div class="edit-link-row">
              <input class="edit-link-input" type="text" readonly value={linkEdicao} />
              <CopyButton text={linkEdicao} label="Copiar link de edição" />
            </div>
            <p class="text-muted edit-link-hint">
              Perdeu o link? Você pode solicitá-lo ao responsável pelo evento.
            </p>
          </div>
        {/if}
        <p class="text-muted done-hint">Pode fechar esta página. Até lá!</p>
      </div>
    </PublicShell>
  {/if}
</main>

<style>
  .responder-root {
    flex: 1;
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  .responder-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .link-btn {
    padding: 0;
    border: none;
    background: transparent;
    color: var(--accent);
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 700;
    cursor: pointer;
  }

  .link-btn:hover {
    text-decoration: underline;
  }

  /* --- Identificação --- */

  .identify-head {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    text-align: center;
  }

  .event-chip {
    background: var(--tint-cyan);
    color: var(--cyan-hover);
    padding: 4px 12px;
  }

  .entry-title {
    margin: 4px 0 0;
    font-size: 2rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.15;
  }

  .entry-sub {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.9375rem;
    line-height: 1.5;
  }

  .identify-card {
    max-width: none;
    margin-top: 24px;
    gap: 16px;
    padding: 24px;
  }

  .identify-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: center;
    gap: 12px;
    margin-top: 20px;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .meta-dot {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: var(--border-strong);
  }

  /* --- Identidade na topbar --- */

  .me-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 10px 4px 4px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text);
    font-family: var(--font-ui);
    cursor: pointer;
  }

  .me-pill:hover {
    border-color: var(--accent);
  }

  .me-avatar {
    width: 26px;
    height: 26px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--accent);
    color: var(--on-accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 800;
    overflow: hidden;
  }

  .me-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .me-name {
    font-size: 0.8125rem;
    font-weight: 700;
    max-width: 110px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* --- Faixa de progresso (mesma altura do breadcrumb do organizador) --- */

  .progress-bar {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 12px;
    height: var(--crumbbar-h);
    padding: 0 20px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
  }

  .progress-count {
    font-size: 0.8125rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .segments {
    flex: 1;
    display: flex;
    gap: 4px;
  }

  .segment {
    flex: 1;
    height: 4px;
    padding: 0;
    border: none;
    border-radius: 999px;
    background: var(--border);
    cursor: pointer;
  }

  .segment.done {
    background: color-mix(in srgb, var(--accent) 55%, var(--bg-elev));
  }

  .segment.current {
    background: var(--accent);
  }

  .progress-done {
    font-size: 0.75rem;
    white-space: nowrap;
  }

  /* --- Pergunta --- */

  .q-main {
    flex: 1;
    min-height: 0;
    display: flex;
    justify-content: center;
    padding: 40px 32px 24px;
    overflow-y: auto;
  }

  .q-col {
    width: 640px;
    max-width: 100%;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .q-edit-note {
    margin: 0;
    font-size: 0.8125rem;
  }

  .q-title {
    margin: 0;
    font-size: 2rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.2;
  }

  .q-options {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  /*
   * Seleção marcada por borda + check, não por preenchimento sólido: fica
   * legível também no tema escuro.
   */
  .q-option {
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 60px;
    padding: 12px 18px;
    border-radius: var(--radius-row);
    background: var(--bg-elev);
    border: 1px solid var(--border);
    font-size: 1rem;
    line-height: 1.25;
    cursor: pointer;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
  }

  .q-option:hover {
    border-color: var(--accent);
    box-shadow: var(--shadow-hover);
  }

  .q-option.selected {
    border: 2px solid var(--accent);
    padding: 11px 17px;
  }

  .q-option-text {
    flex: 1;
  }

  .q-option-check {
    color: var(--accent);
    font-size: 1rem;
    font-weight: 800;
  }

  .q-text-wrap {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .q-textarea {
    width: 100%;
    padding: 16px 18px;
    border-radius: var(--radius-row);
    border: 1.5px solid var(--border-strong);
    background: var(--bg-elev);
    color: var(--text);
    font-family: inherit;
    font-size: 1rem;
    line-height: 1.5;
    resize: vertical;
    outline: none;
  }

  .q-textarea:focus {
    border-color: var(--accent);
  }

  .q-text-count {
    align-self: flex-end;
    font-size: 0.75rem;
  }

  /* --- Rodapé de ações --- */

  .dock {
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    padding: 16px 32px;
    background: var(--bg-elev);
    border-top: 1px solid var(--border);
  }

  .dock-col {
    width: 640px;
    max-width: 100%;
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .dock-spacer {
    flex: 1;
  }

  .dock-hint {
    font-size: 0.75rem;
  }

  /* --- Revisão --- */

  .review-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .review-identity,
  .review-row {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 14px 18px;
    border-radius: var(--radius-row);
    background: var(--bg-elev);
    border: 1px solid var(--border);
    color: var(--text);
    font-family: var(--font-ui);
    text-align: left;
    cursor: pointer;
    transition: border-color 0.15s ease;
  }

  .review-identity:hover,
  .review-row:hover {
    border-color: var(--accent);
  }

  .review-avatar {
    width: 44px;
    height: 44px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--accent);
    color: var(--on-accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    font-weight: 800;
    overflow: hidden;
  }

  .review-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .review-identity-text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-size: 0.8125rem;
  }

  .review-identity-name {
    font-size: 0.9375rem;
    font-weight: 700;
  }

  .review-edit-label {
    flex-shrink: 0;
    padding: 8px 14px;
    border-radius: var(--radius-control);
    border: 1px solid var(--border-strong);
    font-size: 0.8125rem;
    font-weight: 700;
  }

  .review-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .review-row-n {
    width: 22px;
    height: 22px;
    flex-shrink: 0;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 800;
    background: var(--surface-muted);
    color: var(--text-muted);
  }

  .review-row-text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .review-row-q {
    font-size: 0.75rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .review-row-a {
    font-size: 0.9375rem;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .review-row-a.unanswered {
    color: var(--danger);
  }

  /* --- Privacidade --- */

  .crumb-simple {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 12px;
    height: var(--crumbbar-h);
    padding: 0 20px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
  }

  .crumb-simple-title {
    font-size: 0.8125rem;
    font-weight: 700;
  }

  .privacidade-body {
    flex: 1;
    display: flex;
    justify-content: center;
    padding: 32px 24px 56px;
    overflow-y: auto;
  }

  .privacidade-col {
    width: 720px;
    max-width: 100%;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 20px;
  }

  .privacidade-intro {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .privacidade-items {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .privacidade-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 16px 18px;
    border-radius: var(--radius-row);
    background: var(--bg-elev);
    border: 1px solid var(--border);
    font-size: 0.875rem;
    line-height: 1.5;
  }

  .privacidade-item-title {
    font-size: 0.9375rem;
    font-weight: 700;
  }

  .privacidade-note {
    margin: 0;
    padding: 14px 16px;
    border-radius: var(--radius-row);
    background: var(--surface-muted);
    border: 1px solid var(--border);
    font-size: 0.8125rem;
    line-height: 1.5;
  }

  .privacidade-col :global(.btn) {
    align-self: flex-start;
  }

  /* --- Concluído --- */

  .done-card {
    max-width: none;
    align-items: center;
    text-align: center;
    gap: 16px;
  }

  .done-avatar {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    background: var(--accent);
    color: var(--on-accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.5rem;
    font-weight: 800;
    overflow: hidden;
  }

  .done-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .done-title {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 800;
    letter-spacing: -0.02em;
  }

  .done-sub {
    margin: 0;
    font-size: 0.9375rem;
    line-height: 1.45;
  }

  .done-hint {
    margin: 0;
    font-size: 0.8125rem;
  }

  .edit-link-box {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 14px;
    background: var(--surface-muted);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    text-align: left;
  }

  .edit-link-label {
    margin: 0;
    font-size: 0.8125rem;
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
    font-family: var(--font-ui);
    font-size: 0.8125rem;
  }

  .edit-link-hint {
    margin: 4px 0 0;
    font-size: 0.75rem;
  }

  @media (max-width: 720px) {
    .q-main {
      padding: 24px 20px 16px;
    }

    .q-title {
      font-size: 1.5rem;
    }

    .dock {
      padding: 14px 20px 24px;
    }

    .dock-hint {
      display: none;
    }

    .progress-done {
      display: none;
    }
  }
</style>
