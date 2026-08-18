<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import PinChip from '../components/PinChip.svelte';
  import PresentationStage from '../components/PresentationStage.svelte';
  import PublicShell from '../components/PublicShell.svelte';
  import ReactionBar from '../components/ReactionBar.svelte';
  import ReactionBurstLayer from '../components/ReactionBurstLayer.svelte';
  import ThemeToggle from '../components/ThemeToggle.svelte';
  import { fireReaction } from '../lib/reactionStore.js';

  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]+$/;

  // A URL é a única fonte da verdade pra sessão: nada fica salvo no
  // navegador. Sem PIN na URL, não tem como ter entrado — volta pra Home.
  // Um F5 recarrega essa mesma checagem do zero, então só continua na tela
  // de identificação quem chegou com o PIN na barra de endereço (seja pela
  // Home, seja por um link compartilhado que já vem com ?pin=).
  const params = new URLSearchParams(window.location.search);
  const initialPin = (params.get('pin') || '').trim();
  if (!initialPin) {
    navigate('/');
  }

  /*
   * A mesma view serve dois papéis, com layouts diferentes:
   * - telão: escala de projeção, PIN sempre visível, sem interações.
   * - celular: rodapé fixo com reações e recado, ao alcance do polegar.
   * O padrão vem da largura da janela; `?view=telao|celular` força um deles, e
   * o botão no canto troca sem perder a sessão.
   */
  const requestedView = params.get('view');
  let bigScreen = window.innerWidth >= 1024;
  let viewMode = requestedView === 'telao' || requestedView === 'celular' ? requestedView : null;
  $: mode = viewMode || (bigScreen ? 'telao' : 'celular');

  function onResize() {
    bigScreen = window.innerWidth >= 1024;
  }

  let step = 'identify'; // identify | watching
  const pinCode = initialPin;
  let email = '';
  let identifyError = '';
  let joining = false;

  let token = '';
  let snapshot = null;
  let loadError = '';
  let connected = true;
  let eventSource = null;

  let qaDraft = '';
  let qaSent = false;
  let qaError = '';

  async function joinAs(asGuest) {
    identifyError = '';
    const trimmedEmail = asGuest ? '' : email.trim();
    if (!asGuest && !emailRe.test(trimmedEmail)) {
      identifyError = 'Informe um e-mail válido.';
      return;
    }
    joining = true;
    try {
      const result = await api.public.events.live.join(id, {
        pinCode,
        email: trimmedEmail
      });
      token = result.token;
      step = 'watching';
      startWatching();
    } catch (e) {
      identifyError = e.message;
    } finally {
      joining = false;
    }
  }

  async function startWatching() {
    loadError = '';
    try {
      snapshot = await api.public.events.live.state(id, token);
    } catch (e) {
      loadError = e.message;
      return;
    }
    connectStream();
  }

  function connectStream() {
    if (eventSource) eventSource.close();
    eventSource = new EventSource(api.public.events.live.streamUrl(id, token));
    eventSource.onmessage = (e) => {
      connected = true;
      snapshot = JSON.parse(e.data);
    };
    eventSource.onerror = () => {
      connected = false;
    };
    // A própria reação de quem manda só aparece via este eco do SSE (sem
    // disparo otimista no clique) — evita rajada dupla, já que quem manda
    // também está assinando o próprio stream.
    eventSource.addEventListener('reaction', (e) => {
      fireReaction(JSON.parse(e.data).emoji);
    });
  }

  function handleReact(e) {
    api.public.events.live.react(id, token, e.detail).catch(() => {});
  }

  async function submitQuestion() {
    const text = qaDraft.trim();
    if (!text) return;
    qaError = '';
    try {
      await api.public.events.live.submitQuestion(id, token, text);
      qaDraft = '';
      qaSent = true;
      setTimeout(() => {
        qaSent = false;
      }, 3000);
    } catch (e) {
      qaError = e.message;
    }
  }

  function backToJoin() {
    if (eventSource) eventSource.close();
    token = '';
    snapshot = null;
    loadError = '';
    step = 'identify';
  }

  $: currentQuestion =
    snapshot && snapshot.questions.find((q) => q.id === snapshot.currentQuestionId);
</script>

<svelte:window on:resize={onResize} />

<main class="audience">
  {#if !initialPin}
    <div class="audience-center"><p class="text-muted">Redirecionando…</p></div>
  {:else if step === 'identify'}
    <PublicShell>
      <div class="card join-card">
        <div class="card-head">
          <h1>Entrar na apresentação</h1>
          <p class="card-sub">Identifique-se com seu e-mail ou entre como convidado.</p>
        </div>
        <form class="form" novalidate on:submit|preventDefault={() => joinAs(false)}>
          <Input
            label="E-mail"
            type="email"
            bind:value={email}
            error={identifyError}
            placeholder="seu@melhor.email"
            autocomplete="email"
          />
          <Button type="submit" block disabled={joining}>
            {joining ? 'Entrando…' : 'Entrar'}
          </Button>
          <Button
            type="button"
            variant="ghost"
            block
            disabled={joining}
            on:click={() => joinAs(true)}
          >
            Entrar como convidado
          </Button>
        </form>
      </div>
    </PublicShell>
  {:else if loadError}
    <div class="audience-center">
      <div class="card">
        <h1>Não foi possível entrar</h1>
        <p class="form-error">{loadError}</p>
        <div class="error-actions">
          <Button type="button" on:click={startWatching}>Tentar novamente</Button>
          <Button type="button" variant="secondary" on:click={backToJoin}>Voltar</Button>
        </div>
      </div>
    </div>
  {:else if !snapshot}
    <div class="audience-center"><p class="text-muted">Carregando…</p></div>
  {:else if mode === 'telao'}
    <!-- Telão: escala de projeção, o PIN sempre visível para quem chega no
         meio da dinâmica, e nenhuma interação na tela. -->
    <div class="stage-screen">
      <div class="stage-screen-head">
        <img class="stage-logo" src="/img/arandu-logo.png" alt="Arandu" />
        <span class="stage-event">{snapshot.eventTitle}</span>
        {#if !connected}<span class="text-muted stage-reconnect">Reconectando…</span>{/if}
        <span class="stage-spacer"></span>
        <PinChip pin={pinCode} variant="stage" copyable={false} />
        <button type="button" class="mode-btn" on:click={() => (viewMode = 'celular')}>
          Interagir
        </button>
      </div>

      {#if snapshot.message}
        <div class="stage-overlay"><p>{snapshot.message}</p></div>
      {:else if snapshot.blanked}
        <div class="stage-overlay"><p class="text-muted">Aguarde, já voltamos…</p></div>
      {:else if snapshot.questions.length === 0}
        <div class="stage-overlay"><p class="text-muted">Este evento ainda não tem perguntas.</p></div>
      {:else if currentQuestion}
        <h1 class="stage-question">{currentQuestion.title}</h1>
        <div class="stage-zones">
          <PresentationStage
            layout="screen"
            pending={snapshot.pending || []}
            groups={snapshot.groups || []}
            hideZones={snapshot.answersHidden}
            showNames={!snapshot.namesHidden}
          />
        </div>
      {/if}
    </div>
  {:else}
    <!-- Celular de quem assiste: acompanha a dinâmica e interage. -->
    <div class="phone">
      <header class="phone-top">
        <img class="phone-logo" src="/img/arandu-logo.png" alt="Arandu" />
        <span class="phone-event">{snapshot.eventTitle}</span>
        {#if bigScreen}
          <button type="button" class="mode-btn" on:click={() => (viewMode = 'telao')}>
            Modo telão
          </button>
        {/if}
        <ThemeToggle />
      </header>

      <div class="phone-body">
        {#if !connected}
          <p class="text-muted phone-reconnect">Reconectando…</p>
        {/if}
        {#if snapshot.message}
          <div class="phone-overlay"><p>{snapshot.message}</p></div>
        {:else if snapshot.blanked}
          <div class="phone-overlay"><p class="text-muted">Aguarde, já voltamos…</p></div>
        {:else if snapshot.questions.length === 0}
          <p class="text-muted">Este evento ainda não tem perguntas.</p>
        {:else if currentQuestion}
          <h1 class="phone-question">{currentQuestion.title}</h1>
          <PresentationStage
            layout="compact"
            pending={snapshot.pending || []}
            groups={snapshot.groups || []}
            hideZones={snapshot.answersHidden}
            showNames={!snapshot.namesHidden}
          />
        {/if}
      </div>

      {#if snapshot.interactionsEnabled}
        <!-- Fixo no rodapé: reação e recado sempre alcançáveis com o polegar. -->
        <div class="phone-dock">
          <ReactionBar on:react={handleReact} />
          <form class="qa-form" novalidate on:submit|preventDefault={submitQuestion}>
            <input
              class="qa-input"
              bind:value={qaDraft}
              aria-label="Pergunta ou recado pro organizador"
              placeholder="Pergunta pro organizador…"
            />
            <Button type="submit" size="sm" disabled={!qaDraft.trim()}>Enviar</Button>
          </form>
          {#if qaError}<p class="form-error">{qaError}</p>{/if}
          {#if qaSent}<p class="text-muted qa-sent">Enviado!</p>{/if}
        </div>
      {/if}
    </div>
  {/if}

  <ReactionBurstLayer />
</main>

<style>
  .audience {
    flex: 1;
    display: flex;
    flex-direction: column;
    width: 100%;
    min-height: 100vh;
    min-height: 100dvh;
  }

  .audience-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .join-card {
    max-width: none;
  }

  .card-head {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .card-sub {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.875rem;
  }

  .error-actions {
    display: flex;
    gap: 10px;
  }

  .mode-btn {
    flex-shrink: 0;
    padding: 7px 12px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.75rem;
    font-weight: 700;
    cursor: pointer;
  }

  .mode-btn:hover {
    border-color: var(--accent);
    color: var(--text);
  }

  /* --- Telão --- */

  /*
   * Projeção: ninguém rola uma tela projetada no meio da dinâmica. height
   * fixo (não min-height) + overflow:hidden trava o quadro em 100dvh de
   * verdade, com o piso em 1366×768 (notebook/projetor mais comum) — numa
   * tela maior só sobra espaço, nunca aparece scroll. PresentationStage cuida
   * do teto de rostos visíveis por fileira (ver PENDING_CAP/zoneCap lá).
   */
  .stage-screen {
    /* Sem flex:1 de propósito: dentro do <main> em coluna, flex:1 vira
       flex-basis:0 e VENCE a altura explícita, deixando o elemento crescer
       pelo conteúdo em vez de travar em 100dvh — foi exatamente o que vazou
       a segunda fileira de zonas pra fora de 1366×768 num teste com muita
       gente revelada. height sozinho (flex-basis:auto) é respeitado. */
    height: 100vh;
    height: 100dvh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    padding: 3vh 3.5vw 2.5vh;
    box-sizing: border-box;
  }

  .stage-screen-head {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .stage-logo {
    height: 28px;
    width: auto;
  }

  .stage-event {
    font-size: 1.125rem;
    font-weight: 700;
    color: var(--text-muted);
  }

  .stage-reconnect {
    font-size: 0.875rem;
  }

  .stage-spacer {
    flex: 1;
  }

  .stage-question {
    margin: 2.5vh 0 0;
    flex-shrink: 0;
    font-size: clamp(1.75rem, 5.2vh, 3.25rem);
    font-weight: 800;
    letter-spacing: -0.025em;
    line-height: 1.15;
    max-width: 1000px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .stage-zones {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-top: 2.5vh;
  }

  .stage-overlay {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 28px;
    padding: 32px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 14px;
    text-align: center;
  }

  .stage-overlay p {
    margin: 0;
    font-size: 2rem;
    font-weight: 700;
  }

  @media (max-width: 1180px) {
    .stage-question {
      font-size: 2.25rem;
    }
  }

  /* --- Celular --- */

  .phone {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .phone-top {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    height: var(--topbar-h);
    padding: 0 16px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
  }

  .phone-logo {
    height: 22px;
    width: auto;
  }

  .phone-event {
    flex: 1;
    min-width: 0;
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .phone-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 20px 16px;
    overflow-y: auto;
  }

  .phone-reconnect {
    margin: 0;
    font-size: 0.8125rem;
  }

  .phone-question {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.2;
  }

  .phone-overlay {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 160px;
    padding: 24px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    text-align: center;
  }

  .phone-overlay p {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 700;
  }

  .phone-dock {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
    background: var(--bg-elev);
    border-top: 1px solid var(--border);
  }

  .qa-form {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 6px 6px 14px;
    border: 1.5px solid var(--border);
    border-radius: var(--radius-row);
    transition: border-color 0.15s ease;
  }

  .qa-form:focus-within {
    border-color: var(--accent);
  }

  .qa-input {
    flex: 1;
    min-width: 0;
    border: none;
    outline: none;
    background: transparent;
    font-family: var(--font-ui);
    font-size: 0.875rem;
    color: var(--text);
  }

  .qa-input::placeholder {
    color: var(--text-subtle);
  }

  .qa-sent {
    margin: 0;
    font-size: 0.8125rem;
  }
</style>
