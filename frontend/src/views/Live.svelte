<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';
  import Input from '../components/Input.svelte';
  import PresentationStage from '../components/PresentationStage.svelte';

  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]+$/;
  const storageKey = `live-viewer-${id}`;

  let step = 'join'; // join | watching
  let pinCode = '';
  let email = '';
  let joinError = '';
  let joining = false;

  let token = '';
  let snapshot = null;
  let loadError = '';
  let connected = true;
  let eventSource = null;

  restoreSession();

  function restoreSession() {
    try {
      const raw = sessionStorage.getItem(storageKey);
      if (!raw) return;
      const saved = JSON.parse(raw);
      if (saved.token) {
        token = saved.token;
        step = 'watching';
        startWatching();
      }
    } catch {
      // sessionStorage indisponível ou dado corrompido — só ignora, pede pra entrar de novo
    }
  }

  async function submitJoin() {
    joinError = '';
    if (!pinCode.trim()) {
      joinError = 'Informe o PIN do evento.';
      return;
    }
    if (!emailRe.test(email.trim())) {
      joinError = 'Informe um e-mail válido.';
      return;
    }
    joining = true;
    try {
      const result = await api.public.events.live.join(id, {
        pinCode: pinCode.trim(),
        email: email.trim()
      });
      token = result.token;
      try {
        sessionStorage.setItem(storageKey, JSON.stringify({ token: result.token }));
      } catch {
        // sem sessionStorage, segue sem persistir (só perde ao recarregar a página)
      }
      step = 'watching';
      startWatching();
    } catch (e) {
      joinError = e.message;
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
  }

  function backToJoin() {
    try {
      sessionStorage.removeItem(storageKey);
    } catch {
      // ignora
    }
    if (eventSource) eventSource.close();
    token = '';
    snapshot = null;
    loadError = '';
    step = 'join';
  }

  $: currentQuestion =
    snapshot && snapshot.questions.find((q) => q.id === snapshot.currentQuestionId);
</script>

<main class="present-page">
  {#if step === 'join'}
    <Card wide>
      <img class="live-logo" src="/img/porandu-logo-sem-bg.png" alt="Porandu" />
      <form class="form" novalidate on:submit|preventDefault={submitJoin}>
        <p class="invite-text">
          Entre com o PIN do evento e seu e-mail pra acompanhar a apresentação ao vivo.
        </p>
        <Input
          label="PIN do evento"
          bind:value={pinCode}
          placeholder="Ex: DEV-TEAM"
          uppercase
          required
        />
        <Input
          label="E-mail"
          type="email"
          bind:value={email}
          placeholder="seu@melhor.email"
          autocomplete="email"
          required
        />
        {#if joinError}
          <p class="form-error">{joinError}</p>
        {/if}
        <Button type="submit" block disabled={joining}>
          {joining ? 'Entrando…' : 'Entrar'}
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

    {#if snapshot.questions.length === 0}
      <p class="text-muted present-empty">Este evento ainda não tem perguntas.</p>
    {:else if currentQuestion}
      <h2 class="present-question">{currentQuestion.title}</h2>
      <PresentationStage pending={snapshot.pending || []} groups={snapshot.groups || []} />
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

  .live-logo {
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
</style>
