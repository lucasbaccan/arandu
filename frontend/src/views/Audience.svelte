<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Input from '../components/Input.svelte';
  import PresentationStage from '../components/PresentationStage.svelte';
  import ReactionBar from '../components/ReactionBar.svelte';
  import ReactionBurstLayer from '../components/ReactionBurstLayer.svelte';
  import { fireReaction } from '../lib/reactionStore.js';

  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]+$/;

  // A URL é a única fonte da verdade pra sessão: nada fica salvo no
  // navegador. Sem PIN na URL, não tem como ter entrado — volta pra Home.
  // Um F5 recarrega essa mesma checagem do zero, então só continua na tela
  // de identificação quem chegou com o PIN na barra de endereço (seja pela
  // Home, seja por um link compartilhado que já vem com ?pin=).
  const initialPin = (new URLSearchParams(window.location.search).get('pin') || '').trim();
  if (!initialPin) {
    navigate('/');
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

<main
  class="present-page"
  class:present-page-center={!initialPin || step === 'identify' || loadError || !snapshot}
>
  {#if !initialPin}
    <p class="text-muted">Redirecionando…</p>
  {:else if step === 'identify'}
    <Card wide>
      <img class="live-logo" src="/img/arandu-logo.png" alt="Arandu" />
      <form class="form" novalidate on:submit|preventDefault={() => joinAs(false)}>
        <p class="invite-text">Identifique-se com seu e-mail ou entre como convidado.</p>
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
        <Button type="button" variant="secondary" block disabled={joining} on:click={() => joinAs(true)}>
          Entrar como convidado
        </Button>
      </form>
    </Card>
  {:else if loadError}
    <Card title="Não foi possível entrar">
      <p class="form-error">{loadError}</p>
      <div class="live-error-actions">
        <Button type="button" on:click={startWatching}>Tentar novamente</Button>
        <Button type="button" variant="secondary" on:click={backToJoin}>Voltar</Button>
      </div>
    </Card>
  {:else if !snapshot}
    <p class="text-muted">Carregando…</p>
  {:else}
    <h1 class="present-title">{snapshot.eventTitle}</h1>
    {#if !connected}
      <p class="text-muted presentation-hint">Reconectando…</p>
    {/if}

    {#if snapshot.message}
      <div class="live-message"><p>{snapshot.message}</p></div>
    {:else if snapshot.blanked}
      <div class="live-blank"><p class="text-muted">Aguarde, já voltamos…</p></div>
    {:else if snapshot.questions.length === 0}
      <p class="text-muted present-empty">Este evento ainda não tem perguntas.</p>
    {:else if currentQuestion}
      <h2 class="present-question">{currentQuestion.title}</h2>
      <PresentationStage
        pending={snapshot.pending || []}
        groups={snapshot.groups || []}
        hideZones={snapshot.answersHidden}
      />
    {/if}

    {#if snapshot.interactionsEnabled}
      <ReactionBar on:react={handleReact} />
      <form class="qa-form" novalidate on:submit|preventDefault={submitQuestion}>
        <Input
          label="Pergunta ou recado pro organizador"
          bind:value={qaDraft}
          error={qaError}
          placeholder="Escreva aqui…"
        />
        <Button type="submit" disabled={!qaDraft.trim()}>Enviar</Button>
        {#if qaSent}
          <p class="text-muted">Enviado!</p>
        {/if}
      </form>
    {/if}
  {/if}

  <ReactionBurstLayer />
</main>

<style>
  .present-page {
    max-width: none;
    width: 100%;
    min-height: 100vh;
    min-height: 100dvh;
    box-sizing: border-box;
    padding: 24px 32px 32px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .present-page-center {
    align-items: center;
    justify-content: center;
  }

  .live-logo {
    display: block;
    height: 88px;
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

  .live-error-actions {
    display: flex;
    gap: 10px;
  }

  .present-title {
    margin: 0;
    font-size: 1.2rem;
  }

  .present-question {
    margin: 0 0 4px;
    font-size: 1.8rem;
  }

  .present-empty {
    margin: 0;
  }

  .presentation-hint {
    margin: 0;
    font-size: 0.85rem;
    opacity: 0.7;
  }

  .live-message,
  .live-blank {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 160px;
    padding: 24px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 12px;
    text-align: center;
  }

  .live-message p {
    margin: 0;
    font-size: 1.3rem;
  }

  .qa-form {
    display: flex;
    align-items: flex-end;
    gap: 10px;
  }

  .qa-form :global(.field) {
    flex: 1;
  }
</style>
