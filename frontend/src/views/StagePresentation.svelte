<script>
  export let id = '';
  // cols: nº de colunas do placar — 0 (padrão, /present) deixa o
  // PresentationStage escolher sozinho; 1..4 (/present1..4) forçam, pra
  // comparar pelo menu flutuante (PresentVariantMenu.svelte).
  export let cols = 0;

  import { onDestroy } from 'svelte';
  import { api } from '../lib/api.js';
  import PresentationStage from '../components/PresentationStage.svelte';
  import PresentVariantMenu from '../components/PresentVariantMenu.svelte';
  import ReactionBurstLayer from '../components/ReactionBurstLayer.svelte';
  import { fireReaction } from '../lib/reactionStore.js';

  // Janela aberta pelo botão "Modo apresentação" no /stage/ — só o
  // organizador autenticado (dono do evento) consegue abrir, sem PIN nem
  // token de visitante. Nada aqui é clicável (sem reveal, sem Q&A) — mas as
  // reações da plateia aparecem, já que são só exibição, não interação desta
  // tela. Usa a mesma escala de projeção do telão da plateia.
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

<main class="present-page" class:present-center={loading || error || !snapshot}>
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if error}
    <p class="form-error">{error}</p>
  {:else}
    <div class="present-head">
      <img class="present-logo" src="/img/arandu-logo.png" alt="Arandu" />
      <span class="present-event">{snapshot.eventTitle || ''}</span>
    </div>

    {#if snapshot.message}
      <div class="present-overlay"><p>{snapshot.message}</p></div>
    {:else if snapshot.blanked}
      <div class="present-overlay"><p class="text-muted">Aguarde, já voltamos…</p></div>
    {:else if snapshot.questions.length === 0}
      <div class="present-overlay"><p class="text-muted">Este evento ainda não tem perguntas.</p></div>
    {:else if currentQuestion}
      <h1 class="present-question">{currentQuestion.title}</h1>
      <div class="present-zones">
        <PresentationStage
          layout="screen"
          forceCols={cols}
          pending={snapshot.pending || []}
          groups={snapshot.groups || []}
          hideZones={snapshot.answersHidden}
          showNames={!snapshot.namesHidden}
        />
      </div>
    {/if}
  {/if}

  <ReactionBurstLayer />
  {#if !loading && !error}
    <PresentVariantMenu {id} active={cols} />
  {/if}
</main>

<style>
  /*
   * Telão de projeção: ninguém rola uma tela projetada durante a
   * apresentação. height (não min-height) + overflow:hidden trava o quadro
   * em 100dvh de verdade — o piso é 1366×768 (o notebook/projetor mais
   * comum); numa tela maior só sobra espaço, nunca aparece scroll. Título e
   * zonas usam clamp()/vh pra caber nesse piso; PresentationStage mede o
   * espaço restante e adapta a densidade do placar (ver fitDensity lá).
   */
  .present-page {
    width: 100%;
    height: 100vh;
    height: 100dvh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    padding: 3vh 3.5vw 2.5vh;
    box-sizing: border-box;
  }

  .present-center {
    align-items: center;
    justify-content: center;
  }

  .present-head {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .present-logo {
    height: 28px;
    width: auto;
  }

  .present-event {
    font-size: 1.125rem;
    font-weight: 700;
    color: var(--text-muted);
  }

  .present-question {
    margin: 2.5vh 0 0;
    flex-shrink: 0;
    /* Escala com a altura da tela em vez de um px fixo — em 768px de altura
       (piso: 1366×768) isso fica ~40px, dando espaço de sobra pra 2 linhas
       sem empurrar as zonas pra fora. */
    font-size: clamp(1.75rem, 5.2vh, 3.25rem);
    font-weight: 800;
    letter-spacing: -0.025em;
    line-height: 1.15;
    /* Ocupa toda a largura útil da página em vez de travar num px fixo — num
       telão largo, um cap de 1000px deixava metade da tela vazia à direita
       do título. */
    width: 100%;
    /* Título absurdamente longo (raro) corta em 2 linhas em vez de vazar. */
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .present-zones {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-top: 2.5vh;
  }

  .present-overlay {
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

  .present-overlay p {
    margin: 0;
    font-size: 2rem;
    font-weight: 700;
  }
</style>
