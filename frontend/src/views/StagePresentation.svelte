<script>
  export let id = '';

  import { onDestroy } from 'svelte';
  import { api } from '../lib/api.js';
  import PresentationStage from '../components/PresentationStage.svelte';
  import ReactionBurstLayer from '../components/ReactionBurstLayer.svelte';
  import { fireReaction } from '../lib/reactionStore.js';

  // Janela aberta pelo botão "Modo apresentação" no /stage/ — só o
  // organizador autenticado (dono do evento) consegue abrir, sem PIN nem
  // token de visitante. Nada aqui é clicável (sem reveal, sem Q&A) — mas as
  // reações da plateia aparecem, já que são só exibição, não interação desta
  // tela.
  let loading = true;
  let error = '';
  let snapshot = null;
  let eventSource = null;

  onDestroy(() => {
    if (eventSource) eventSource.close();
  });

  load();

  async function load() {
    try {
      snapshot = await api.events.live.presentationState(id);
      connectStream();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function connectStream() {
    eventSource = new EventSource(api.events.live.presentationStreamUrl(id));
    eventSource.onmessage = (e) => {
      snapshot = JSON.parse(e.data);
    };
    eventSource.addEventListener('reaction', (e) => {
      fireReaction(JSON.parse(e.data).emoji);
    });
  }

  $: currentQuestion =
    snapshot && snapshot.questions.find((q) => q.id === snapshot.currentQuestionId);
</script>

<main class="present-page" class:present-page-center={loading || error || !snapshot}>
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if error}
    <p class="form-error">{error}</p>
  {:else if snapshot.message}
    <div class="live-message"><p>{snapshot.message}</p></div>
  {:else if snapshot.blanked}
    <div class="live-blank"><p class="text-muted">Aguarde, já voltamos…</p></div>
  {:else if snapshot.questions.length === 0}
    <p class="text-muted">Este evento ainda não tem perguntas.</p>
  {:else if currentQuestion}
    <h2 class="present-question">{currentQuestion.title}</h2>
    <PresentationStage
      pending={snapshot.pending || []}
      groups={snapshot.groups || []}
      hideZones={snapshot.answersHidden}
      showNames={!snapshot.namesHidden}
    />
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

  .present-question {
    margin: 0 0 4px;
    font-size: 1.8rem;
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
</style>
