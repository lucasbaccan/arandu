<script>
  export let id = '';

  import { onDestroy, onMount, tick } from 'svelte';
  import { api } from '../lib/api.js';
  import PresentationStage from '../components/PresentationStage.svelte';
  import ReactionBurstLayer from '../components/ReactionBurstLayer.svelte';
  import { fireReaction } from '../lib/reactionStore.js';

  // Janela aberta pelo botão "Modo apresentação" no /palco/ — só o
  // organizador autenticado (dono do evento) consegue abrir, sem PIN nem
  // token de visitante. Nada aqui é clicável (sem reveal, sem Q&A) — mas as
  // reações da plateia aparecem, já que são só exibição, não interação desta
  // tela. Usa a mesma escala de projeção do telão da plateia.
  let loading = true;
  let error = '';
  let snapshot = null;
  let eventSource = null;

  // Em telas pequenas o placar usa o layout 'compact' do PresentationStage
  // (linhas legíveis) em vez da escala de projetor; o layout 'screen' segue
  // para o telão. A fila de pendentes em compact mostra todos os rostos —
  // a media query do componente (≤640px) os quebra em linhas.
  let compact = false;

  onMount(() => {
    // Cobre retrato de celular E landscape (largura >640px com altura baixa):
    // um celular deitado não pode cair no layout de projetor.
    const mq = window.matchMedia('(max-width: 640px), (max-height: 480px)');
    compact = mq.matches;
    const onChange = (e) => (compact = e.matches);
    mq.addEventListener('change', onChange);
    return () => mq.removeEventListener('change', onChange);
  });

  onDestroy(() => {
    if (eventSource) eventSource.close();
    window.removeEventListener('resize', measureQuestionOverflow);
  });

  window.addEventListener('resize', measureQuestionOverflow);

  load();

  async function load() {
    try {
      snapshot = await api.eventos.aoVivo.estadoApresentacao(id);
      connectStream();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function connectStream() {
    eventSource = new EventSource(api.eventos.aoVivo.urlFluxoApresentacao(id));
    eventSource.onmessage = (e) => {
      snapshot = JSON.parse(e.data);
    };
    eventSource.addEventListener('reaction', (e) => {
      fireReaction(JSON.parse(e.data).emoji);
    });
  }

  $: currentQuestion =
    snapshot && snapshot.questions.find((q) => q.id === snapshot.currentQuestionId);

  // Pergunta longa demais pro clamp de altura: em vez de cortar o texto,
  // rola verticalmente devagar (ninguém rola uma tela projetada na mão) até
  // dar pra ler tudo, e volta pro início — ver .apresentar-question-scrolling.
  let questionWrapEl;
  let questionInnerEl;
  let scrollDistance = 0;

  $: if (currentQuestion) measureQuestionOverflow();

  async function measureQuestionOverflow() {
    await tick();
    if (!questionWrapEl || !questionInnerEl) return;
    const overflow = questionInnerEl.scrollHeight - questionWrapEl.clientHeight;
    scrollDistance = overflow > 4 ? overflow : 0;
  }

  // Densidade do placar (colunas × escala das pílulas, ver fitDensity em
  // PresentationStage.svelte) — vem do snapshot ao vivo, não da URL: os
  // botões de modo no rodapé de Palco.svelte mudam isso no servidor, e essa
  // janela só reflete o que chega por SSE, em tempo real, sem navegar.
  // 'modo_amplo' = Amplo (padrão — ''/nunca escolhido também cai nele),
  // 'modo_compacto' = Compacto, ou '1'..'4' colunas forçadas.
  $: densityMode = (snapshot && snapshot.presentDensityMode) || '';
  $: effectiveMode = densityMode || 'modo_amplo';
  $: isAmplo = effectiveMode === 'modo_amplo';
  $: forceCols = isAmplo ? 0 : Number(effectiveMode) || 0;
</script>

<main class="apresentar-page" class:apresentar-center={loading || error || !snapshot}>
  {#if loading}
    <p class="text-muted">Carregando…</p>
  {:else if error}
    <p class="form-error">{error}</p>
  {:else}
    <div class="apresentar-head">
      <img class="apresentar-logo" src="/img/arandu-logo.png" alt="Arandu" />
      <span class="apresentar-event" title={snapshot.eventTitle || ''}>{snapshot.eventTitle || ''}</span>
    </div>

    {#if snapshot.message}
      <div class="apresentar-overlay"><p>{snapshot.message}</p></div>
    {:else if snapshot.blanked}
      <div class="apresentar-overlay"><p class="text-muted">Aguarde, já voltamos…</p></div>
    {:else if snapshot.questions.length === 0}
      <div class="apresentar-overlay"><p class="text-muted">Este evento ainda não tem perguntas.</p></div>
    {:else if currentQuestion}
      {#if compact}
        <p class="apresentar-hint">💡 Abra numa tela maior para uma experiencia melhor</p>
      {/if}
      <h1 class="apresentar-question" bind:this={questionWrapEl}>
        <span
          class="apresentar-question-inner"
          class:apresentar-question-scrolling={scrollDistance > 0}
          style={scrollDistance > 0 ? `--scroll-distance: -${scrollDistance}px` : ''}
          bind:this={questionInnerEl}
        >{currentQuestion.title}</span>
      </h1>
      <div class="apresentar-zones">
        <PresentationStage
          layout={compact ? 'compact' : 'screen'}
          {forceCols}
          amplo={isAmplo}
          pending={snapshot.pending || []}
          groups={snapshot.groups || []}
          hideZones={snapshot.answersHidden}
          showNames={!snapshot.namesHidden}
        />
      </div>
    {/if}
  {/if}

  <ReactionBurstLayer />
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
  .apresentar-page {
    width: 100%;
    height: 100vh;
    height: 100dvh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    padding: 2.5vh 3vw 1.25vh;
    box-sizing: border-box;
  }

  .apresentar-center {
    align-items: center;
    justify-content: center;
  }

  .apresentar-head {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 14px;
    min-width: 0;
  }

  .apresentar-logo {
    height: 28px;
    width: auto;
    flex-shrink: 0;
  }

  .apresentar-event {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 1.125rem;
    font-weight: 700;
    color: var(--text-muted);
  }

  .apresentar-question {
    margin: 1.25vh 0 0;
    flex-shrink: 0;
    /* Escala com a altura da tela em vez de um px fixo — menor que antes pra
       sobrar altura pro título inteiro (até 3 linhas) sem cortar. */
    font-size: clamp(1.15rem, 3.2vh, 2.1rem);
    font-weight: 800;
    letter-spacing: -0.025em;
    line-height: 1.2;
    /* Ocupa toda a largura útil da página em vez de travar num px fixo — num
       telão largo, um cap de 1000px deixava metade da tela vazia à direita
       do título. */
    width: 100%;
    /* Título absurdamente longo (raro): em vez de cortar, trava a altura e
       deixa o .apresentar-question-scrolling rolar o texto até dar pra ler tudo. */
    max-height: 22vh;
    overflow: hidden;
    position: relative;
  }

  .apresentar-question-inner {
    display: block;
  }

  .apresentar-question-scrolling {
    animation: apresentar-question-scroll 14s ease-in-out infinite;
  }

  @keyframes apresentar-question-scroll {
    0%,
    10% {
      transform: translateY(0);
    }
    50%,
    60% {
      transform: translateY(var(--scroll-distance));
    }
    100% {
      transform: translateY(0);
    }
  }

  .apresentar-zones {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-top: 1.5vh;
  }

  .apresentar-overlay {
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
    /* Mensagens longas nunca devem ficar cortadas sem acesso. */
    overflow-y: auto;
  }

  .apresentar-overlay p {
    margin: 0;
    font-size: 2rem;
    font-weight: 700;
  }

  .apresentar-hint {
    flex-shrink: 0;
    align-self: center;
    margin: 1.25vh 0 0;
    padding: 8px 14px;
    border-radius: var(--radius-control);
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
    font-size: 0.8125rem;
    color: var(--text-muted);
    text-align: center;
  }

  /* Mobile: fonte da mensagem menor (2rem em 375px dá ~13 caracteres/linha)
     e a página rola (o placar compacto mostra todos os participantes — cortar
     sem scroll deixaria zonas inacessíveis). */
  @media (max-width: 640px) {
    .apresentar-overlay p {
      font-size: 1.375rem;
    }

    .apresentar-page {
      overflow-y: auto;
      overflow-x: hidden;
    }
  }
</style>
