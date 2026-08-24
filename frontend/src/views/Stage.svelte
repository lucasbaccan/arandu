<script>
  export let id = '';

  import { onDestroy, tick } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { flip } from 'svelte/animate';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { questionKindInfo } from '../lib/eventStatus.js';
  import Button from '../components/Button.svelte';
  import Chip from '../components/Chip.svelte';
  import CrumbBar from '../components/CrumbBar.svelte';
  import Input from '../components/Input.svelte';
  import PinChip from '../components/PinChip.svelte';
  import Switch from '../components/Switch.svelte';
  import Tabs from '../components/Tabs.svelte';
  import TopBar from '../components/TopBar.svelte';
  import ReactionBurstLayer from '../components/ReactionBurstLayer.svelte';
  import { fireReaction } from '../lib/reactionStore.js';

  let loading = true;
  let error = '';

  let event = null;
  let questions = [];
  let participants = [];
  let currentIndex = 0;

  let blanked = false;
  let answersHidden = false;
  let namesHidden = false;
  let interactionsEnabled = true;
  let message = '';
  let messageDraft = '';
  let qaInbox = [];
  let adminEventSource = null;
  // '' = automático, 'smart', ou '1'..'4' — ver PresentationStage.svelte
  // (forceCols/smart). Estado do servidor, igual blanked/answersHidden: os
  // botões de modo no rodapé mudam isso ao vivo pra quem já tiver a janela
  // de apresentação (ou a tela pública /audience) aberta.
  let presentDensityMode = '';

  // Trilho com abas no lugar de quatro blocos empilhados.
  let railTab = 'questions'; // questions | qa | notice

  // Navegar pelas setas (goPrev/goNext) muda currentIndex sem clicar na
  // lista — rola o painel sozinho pra pergunta ativa nunca ficar fora da
  // vista, sem exigir scroll manual do organizador.
  let railPanelEl;
  $: if (railTab === 'questions' && currentIndex >= 0) scrollActiveQuestionIntoView();

  async function scrollActiveQuestionIntoView() {
    await tick();
    if (!railPanelEl || typeof railPanelEl.scrollTo !== 'function') return;
    const activeEl = railPanelEl.querySelector('.question-row.active');
    if (!activeEl) return;

    // Inclui até 2 perguntas antes/depois na área visível (não só a ativa),
    // pra dar uma prévia do que vem antes/depois sem precisar rolar na mão.
    const sibling = (el, prop, steps) => {
      for (let i = 0; i < steps && el[prop]; i += 1) el = el[prop];
      return el;
    };
    const prevEl = sibling(activeEl, 'previousElementSibling', 2);
    const nextEl = sibling(activeEl, 'nextElementSibling', 2);
    // offsetTop é relativo ao offsetParent posicionado mais próximo, não ao
    // scroll container — por isso .rail-panel precisa de position:relative
    // (senão os cálculos abaixo ficam num sistema de coordenadas errado e a
    // rolagem pra cima nunca dispara).
    const rangeTop = prevEl.offsetTop;
    const rangeBottom = nextEl.offsetTop + nextEl.offsetHeight;
    const { scrollTop, clientHeight } = railPanelEl;

    if (rangeTop < scrollTop) {
      railPanelEl.scrollTo({ top: rangeTop, behavior: 'smooth' });
    } else if (rangeBottom > scrollTop + clientHeight) {
      railPanelEl.scrollTo({ top: rangeBottom - clientHeight, behavior: 'smooth' });
    }
  }

  onDestroy(() => {
    if (adminEventSource) adminEventSource.close();
  });

  // Modos de densidade do placar da apresentação (ver fitDensity em
  // PresentationStage.svelte) — o mesmo menu que antes era só a
  // engrenagem ⚙ dentro da janela de apresentação, agora como botões no
  // rodapé daqui. Cada clique manda o modo pro servidor (como
  // blanked/answersHidden) — quem já estiver com /present ou /audience
  // abertos vê a densidade trocar ao vivo, sem precisar recarregar nem
  // alternar pra tela do projetor. `mode` é o valor exato que o servidor
  // espera (ver validPresentDensityModes no backend).
  const PRESENT_MODES = [
    { mode: '', label: 'Automático', icon: 'A' },
    { mode: 'smart', label: 'Smart', icon: '★' },
    { mode: '1', label: '1 coluna', icon: '1' },
    { mode: '2', label: '2 colunas', icon: '2' },
    { mode: '3', label: '3 colunas', icon: '3' },
    { mode: '4', label: '4 colunas', icon: '4' }
  ];

  function selectPresentMode(mode) {
    presentDensityMode = mode;
    api.events.live.setDensityMode(id, mode).catch(() => {});
    openPresentationWindow();
  }

  // Referência da janela de apresentação (somente leitura, sem clique, sem
  // controles de admin — só a pergunta, as opções e os participantes; feita
  // pra projetar ou compartilhar numa chamada sem expor o painel do
  // organizador). Guardar a referência deixa reaproveitar a MESMA janela em
  // vez de abrir uma nova a cada clique em "Modo apresentação" ou num botão
  // de modo — a densidade em si não depende mais da URL (ver
  // presentDensityMode acima), então aqui só importa abrir/focar.
  let presentWindow = null;

  function openPresentationWindow() {
    if (presentWindow && !presentWindow.closed) {
      presentWindow.focus();
      return;
    }
    // Passar "features" (largura/altura) faz o navegador abrir uma janela
    // de verdade (sem abas, sem barra de endereço) em vez de só uma nova
    // aba. 1366×768: a resolução nativa mais comum de notebook/projetor —
    // o conteúdo em si é feito pra caber nesse piso sem rolagem (ver
    // StagePresentation.svelte); isso só evita abrir menor que isso por
    // padrão. Pra projetar de verdade, dá F11 na janela.
    presentWindow = window.open(`/stage/${id}/present`, `arandu-present-${id}`, 'width=1366,height=768');
  }

  function handleKeydown(e) {
    // Atalhos são globais (svelte:window), mas não podem competir com
    // digitação normal em campos de texto — ex: o aviso pra tela dos
    // participantes, onde espaço/R precisam virar caracteres, não ações.
    const tag = e.target.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA' || e.target.isContentEditable) {
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
    }
  }

  // revealed[questionId] = Set com os ids dos participantes já revelados nessa pergunta
  let revealed = {};

  load();

  async function load() {
    try {
      const [{ event: ev }, { questions: qs }, { participants: ps }, adminSnap] = await Promise.all([
        api.events.get(id),
        api.events.questions.list(id),
        api.events.responses.list(id),
        api.events.live.adminState(id)
      ]);
      event = ev;
      questions = qs;
      participants = ps;

      // Retoma de onde a apresentação parou: pergunta atual e revelação já
      // feita vêm do servidor (live.Manager), não começam sempre do zero —
      // um F5 ou reabrir /stage não deveria voltar pra pergunta 1 com tudo
      // pendente de novo.
      const revealedMap = adminSnap.revealed || {};
      revealed = Object.fromEntries(qs.map((q) => [q.id, new Set(revealedMap[q.id] || [])]));
      blanked = adminSnap.blanked;
      answersHidden = adminSnap.answersHidden;
      namesHidden = adminSnap.namesHidden;
      interactionsEnabled = adminSnap.interactionsEnabled;
      message = adminSnap.message;
      messageDraft = adminSnap.message;
      qaInbox = adminSnap.qaInbox;
      presentDensityMode = adminSnap.presentDensityMode || '';

      if (qs.length > 0) {
        const resumeIndex = qs.findIndex((q) => q.id === adminSnap.currentQuestionId);
        currentIndex = resumeIndex >= 0 ? resumeIndex : 0;
        // Só força a pergunta 1 no servidor se a apresentação nunca tinha
        // sido iniciada (evento novo) — senão preserva onde já estava.
        if (resumeIndex < 0) syncQuestion(qs[0].id);
      }

      connectAdminStream();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  // Espelha o estado ao vivo pro servidor (best-effort — não bloqueia nem
  // quebra a UI local se a rede falhar), pra quem está assistindo em
  // /audience/:id ver a mesma coisa em tempo real.
  function syncQuestion(questionId) {
    api.events.live.setQuestion(id, questionId).catch(() => {});
  }

  // Conexão só de leitura: traz de volta blank/aviso/interações se a página
  // recarregar, e entrega as reações e mensagens de Q&A que chegam ao vivo.
  function connectAdminStream() {
    if (adminEventSource) adminEventSource.close();
    adminEventSource = new EventSource(api.events.live.adminStreamUrl(id));
    adminEventSource.onmessage = (e) => {
      const snap = JSON.parse(e.data);
      blanked = snap.blanked;
      answersHidden = snap.answersHidden;
      namesHidden = snap.namesHidden;
      interactionsEnabled = snap.interactionsEnabled;
      message = snap.message;
      messageDraft = snap.message;
      qaInbox = snap.qaInbox;
      presentDensityMode = snap.presentDensityMode || '';
    };
    adminEventSource.addEventListener('reaction', (e) => {
      fireReaction(JSON.parse(e.data).emoji);
    });
  }

  function toggleBlanked() {
    blanked = !blanked;
    api.events.live.setBlanked(id, blanked).catch(() => {});
  }

  function toggleAnswersHidden() {
    answersHidden = !answersHidden;
    api.events.live.setAnswersHidden(id, answersHidden).catch(() => {});
  }

  function toggleInteractions() {
    interactionsEnabled = !interactionsEnabled;
    api.events.live.setInteractionsEnabled(id, interactionsEnabled).catch(() => {});
  }

  // Estado do servidor (como os outros switches) — afeta a legenda de nome
  // sob cada rosto na janela de apresentação (/stage/:id/present) e na tela
  // da plateia (/audience/:id); esta tela (/stage) sempre mostra os nomes,
  // independente disso.
  function toggleNamesHidden() {
    namesHidden = !namesHidden;
    api.events.live.setNamesHidden(id, namesHidden).catch(() => {});
  }

  function firstName(p) {
    return (p.name || p.email).split(' ')[0];
  }

  function sendMessage() {
    message = messageDraft.trim();
    api.events.live.setMessage(id, message).catch(() => {});
  }

  function clearMessage() {
    message = '';
    messageDraft = '';
    api.events.live.setMessage(id, '').catch(() => {});
  }

  function dismissQA(messageId) {
    qaInbox = qaInbox.filter((m) => m.id !== messageId);
    api.events.live.dismissQA(id, messageId).catch(() => {});
  }

  function openAudienceScreen() {
    // view=telao: essa janela é a projeção, não o celular de quem assiste.
    window.open(
      `/audience/${id}?pin=${encodeURIComponent(event.pinCode.toUpperCase())}&view=telao`,
      '_blank'
    );
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
          label: opt.text,
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
      const label = normalizeOpenText(answerFor(p, q)?.text) || '—';
      if (!groups.has(label)) groups.set(label, { label, participants: [] });
      if (revealedSet.has(p.id)) groups.get(label).participants.push(p);
    }
    return Array.from(groups.values());
  }

  $: groups = currentQuestion && currentQuestion.type === 'OPEN_TEXT' ? openGroups : choiceGroups;

  function reveal(p) {
    if (!currentQuestion) return;
    const questionId = currentQuestion.id;
    const set = revealed[questionId];
    if (set.has(p.id)) {
      set.delete(p.id);
      revealed = { ...revealed };
      api.events.live.unreveal(id, questionId, p.id).catch(() => {});
    } else {
      set.add(p.id);
      revealed = { ...revealed };
      api.events.live.reveal(id, questionId, p.id).catch(() => {});
    }
  }

  function revealAll() {
    if (!currentQuestion) return;
    const set = revealed[currentQuestion.id];
    const questionId = currentQuestion.id;
    pending.forEach((p, i) => {
      setTimeout(() => {
        set.add(p.id);
        revealed = { ...revealed };
        api.events.live.reveal(id, questionId, p.id).catch(() => {});
      }, i * 150);
    });
  }

  function resetReveal() {
    if (!currentQuestion) return;
    revealed = { ...revealed, [currentQuestion.id]: new Set() };
    api.events.live.reset(id, currentQuestion.id).catch(() => {});
  }

  // Reseta a revelação de TODAS as perguntas de uma vez (diferente do
  // "Reiniciar pergunta" do rodapé, que só afeta a pergunta atual) — pra
  // recomeçar a apresentação inteira do zero.
  function resetAllReveals() {
    revealed = Object.fromEntries(questions.map((q) => [q.id, new Set()]));
    api.events.live.resetAll(id).catch(() => {});
  }

  $: anyRevealed = Object.values(revealed).some((set) => set.size > 0);

  function goPrev() {
    if (currentIndex > 0) {
      currentIndex -= 1;
      syncQuestion(questions[currentIndex].id);
    }
  }

  function goNext() {
    if (currentIndex < questions.length - 1) {
      currentIndex += 1;
      syncQuestion(questions[currentIndex].id);
    }
  }

  function jumpToQuestion(i) {
    if (i === currentIndex) return;
    currentIndex = i;
    syncQuestion(questions[i].id);
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<ReactionBurstLayer />

<main class="shell">
  {#if loading}
    <div class="stage-center"><p class="text-muted">Carregando…</p></div>
  {:else if error}
    <div class="stage-center"><p class="form-error">{error}</p></div>
  {:else}
    <TopBar area="Organizador" />

    <CrumbBar
      crumbs={[
        { label: 'Eventos', href: '/dashboard' },
        { label: event.title, href: `/events/${id}` },
        { label: 'Ao vivo' }
      ]}
    >
      <Chip
        slot="status"
        dot
        label={`${participants.length} na sala`}
        tint="var(--tint-cyan)"
        color="var(--cyan-hover)"
      />
      <svelte:fragment slot="actions">
        <PinChip pin={event.pinCode} variant="boxed" />
        <Button variant="secondary" size="sm" on:click={resetAllReveals} disabled={!anyRevealed}>
          Reiniciar tudo
        </Button>
        <Button size="sm" on:click={openPresentationWindow}>Modo apresentação</Button>
      </svelte:fragment>
    </CrumbBar>

    {#if questions.length === 0}
      <div class="stage-center">
        <p class="text-muted">Este evento ainda não tem perguntas.</p>
      </div>
    {:else}
      <div class="stage-layout">
        <div class="stage-main">
          <div class="stage-main-body">
            <h2 class="stage-question-title">{currentQuestion.title}</h2>

            <div class="pending-block">
              <div class="pending-head">
                <span class="overline">Pendentes · clique para revelar</span>
                <span class="overline">{pending.length}</span>
              </div>
              <div class="pending-strip">
                {#each pending as p (p.id)}
                  <div class="face-wrap" animate:flip={{ duration: 350 }} out:fade={{ duration: 150 }}>
                    <button
                      type="button"
                      class="face"
                      title={p.name || p.email}
                      aria-label={`Revelar resposta de ${p.name || p.email}`}
                      on:click={() => reveal(p)}
                    >
                      {#if p.photo}
                        <img src={p.photo} alt="" />
                      {:else}
                        <span class="face-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
                      {/if}
                    </button>
                    <span class="face-name">{firstName(p)}</span>
                  </div>
                {/each}
                {#if pending.length === 0}
                  <p class="text-muted pending-empty">Todo mundo já foi revelado.</p>
                {/if}
              </div>
            </div>

            <div class="zones-grid">
              {#each groups as group (group.label)}
                <div class="zone">
                  <div class="zone-head">
                    <span class="zone-label">{group.label}</span>
                    <span class="zone-count">{group.participants.length}</span>
                  </div>
                  <div class="zone-faces">
                    {#each group.participants as p (p.id)}
                      <div class="face-wrap" animate:flip={{ duration: 350 }} in:fly={{ y: -30, duration: 350 }}>
                        <button
                          type="button"
                          class="face"
                          title={p.name || p.email}
                          aria-label={`Desrevelar resposta de ${p.name || p.email}`}
                          on:click={() => reveal(p)}
                        >
                          {#if p.photo}
                            <img src={p.photo} alt="" />
                          {:else}
                            <span class="face-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
                          {/if}
                        </button>
                        <span class="face-name">{firstName(p)}</span>
                      </div>
                    {/each}
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <!-- Ações de revelação à esquerda, navegação de pergunta ao centro,
               modos da janela de apresentação à direita. -->
          <div class="stage-dock">
            <div class="dock-left">
              <Button size="sm" on:click={revealAll} disabled={pending.length === 0}>
                Revelar tudo
              </Button>
              <Button variant="secondary" size="sm" on:click={resetReveal} disabled={revealedIds.size === 0}>
                Reiniciar pergunta
              </Button>
            </div>

            <div class="dock-nav">
              <button
                type="button"
                class="nav-btn"
                aria-label="Pergunta anterior"
                disabled={currentIndex === 0}
                on:click={goPrev}
              >←</button>
              <span class="dock-counter">{currentIndex + 1} / {questions.length}</span>
              <button
                type="button"
                class="nav-btn"
                aria-label="Próxima pergunta"
                disabled={currentIndex === questions.length - 1}
                on:click={goNext}
              >→</button>
            </div>

            <div class="dock-right">
              <div class="present-modes">
                {#each PRESENT_MODES as mode (mode.mode)}
                  <button
                    type="button"
                    class="mode-btn"
                    class:active={presentDensityMode === mode.mode}
                    title={`Modo apresentação — ${mode.label}`}
                    aria-label={`Modo apresentação — ${mode.label}`}
                    aria-pressed={presentDensityMode === mode.mode}
                    on:click={() => selectPresentMode(mode.mode)}
                  >{mode.icon}</button>
                {/each}
              </div>
            </div>
          </div>
        </div>

        <div class="stage-rail">
          <Tabs
            compact
            bind:value={railTab}
            tabs={[
              { value: 'questions', label: 'Perguntas', count: questions.length },
              { value: 'qa', label: 'Q&A', count: qaInbox.length },
              { value: 'notice', label: 'Aviso' }
            ]}
          />

          <div class="rail-panel" bind:this={railPanelEl}>
            {#if railTab === 'questions'}
              {#each questions as q, i (q.id)}
                <button
                  type="button"
                  class="question-row"
                  class:active={i === currentIndex}
                  on:click={() => jumpToQuestion(i)}
                >
                  <span class="question-n">{i + 1}</span>
                  <span class="question-info">
                    <span class="question-title-text">{q.title}</span>
                    <Chip
                      shape="square"
                      label={questionKindInfo(q.type).label}
                      tint={questionKindInfo(q.type).tint}
                      color={questionKindInfo(q.type).color}
                    />
                  </span>
                </button>
              {/each}
            {:else if railTab === 'qa'}
              {#if qaInbox.length === 0}
                <p class="text-muted rail-empty">Nenhuma mensagem ainda.</p>
              {:else}
                {#each qaInbox as m (m.id)}
                  <div class="qa-item">
                    <span class="qa-item-author">{m.name || m.email || 'Convidado'}</span>
                    <span class="qa-item-text">{m.text}</span>
                    <button type="button" class="qa-dismiss" on:click={() => dismissQA(m.id)}>
                      Dispensar
                    </button>
                  </div>
                {/each}
              {/if}
            {:else}
              <div class="notice-form">
                <Input
                  label="Aviso pra tela dos participantes"
                  bind:value={messageDraft}
                  placeholder="Ex: Voltamos em 5 minutos"
                />
                <div class="notice-actions">
                  <Button size="sm" on:click={sendMessage} disabled={messageDraft.trim() === message}>
                    Enviar aviso
                  </Button>
                  <Button variant="secondary" size="sm" on:click={clearMessage} disabled={!message}>
                    Limpar
                  </Button>
                </div>
              </div>
            {/if}
          </div>

          <!-- Sempre visíveis: são o que a plateia vê agora. -->
          <div class="rail-switches">
            <span class="overline">Tela dos participantes</span>
            <div class="control-row">
              <Switch aria-label="Tela em branco" checked={blanked} on:change={toggleBlanked} />
              <span class:on={blanked}>Tela em branco</span>
            </div>
            <div class="control-row">
              <Switch aria-label="Ocultar respostas" checked={answersHidden} on:change={toggleAnswersHidden} />
              <span class:on={answersHidden}>Ocultar respostas</span>
            </div>
            <div class="control-row">
              <Switch aria-label="Ocultar nomes" checked={namesHidden} on:change={toggleNamesHidden} />
              <span class:on={namesHidden}>Ocultar nomes</span>
            </div>
            <div class="control-row">
              <Switch aria-label="Interações da plateia" checked={interactionsEnabled} on:change={toggleInteractions} />
              <span class:on={interactionsEnabled}>Interações da plateia</span>
            </div>
          </div>

          <div class="rail-footer">
            <Button variant="secondary" block on:click={openAudienceScreen}>
              Abrir tela da plateia
            </Button>
          </div>
        </div>
      </div>
    {/if}
  {/if}
</main>

<style>
  .shell {
    flex: none;
    height: 100vh;
    height: 100dvh;
  }

  .stage-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .stage-layout {
    flex: 1;
    min-height: 0;
    display: flex;
  }

  .stage-main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .stage-main-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 18px 20px 16px;
    overflow-y: auto;
  }

  .stage-question-title {
    margin: 0;
    /* Menor que antes (era 2.125rem fixo) — junto com o padding reduzido
       acima, dá mais chance da pergunta caber na tela sem precisar rolar.
       Ainda quebra em quantas linhas precisar (sem clamp/corte), então uma
       pergunta longa sempre fica legível por inteiro, só empurra o resto
       do painel pra baixo do scroll quando não cabe. */
    font-size: clamp(1.25rem, 1.5vw + 1rem, 1.75rem);
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.2;
  }

  .pending-block {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .pending-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }

  .pending-strip {
    display: flex;
    flex-wrap: nowrap;
    overflow-x: auto;
    gap: 10px;
    min-height: 76px;
    padding: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
  }

  .pending-empty {
    margin: 0;
    font-size: 0.875rem;
  }

  .zones-grid {
    flex: 1;
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-auto-rows: minmax(120px, auto);
    gap: 12px;
  }

  @media (max-width: 640px) {
    .zones-grid {
      grid-template-columns: 1fr;
    }
  }

  .zone {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
    border-radius: var(--radius-row);
    border: 1px solid var(--border);
    background: var(--bg-elev);
  }

  .zone-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 10px;
  }

  .zone-label {
    font-size: 1rem;
    font-weight: 700;
  }

  .zone-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 26px;
    padding: 2px 8px;
    border-radius: 999px;
    background: var(--surface-muted);
    font-size: 0.8125rem;
    font-weight: 800;
  }

  .zone-faces {
    display: flex;
    flex-wrap: wrap;
    align-content: flex-start;
    gap: 10px;
    min-height: 52px;
  }

  .face-wrap {
    display: flex;
    flex-direction: column;
    align-items: center;
    flex-shrink: 0;
    gap: 5px;
    width: 56px;
  }

  .face {
    width: 46px;
    height: 46px;
    flex-shrink: 0;
    border-radius: 50%;
    border: none;
    padding: 0;
    overflow: hidden;
    cursor: pointer;
    background: var(--accent);
    font-family: var(--font-ui);
    font-weight: 800;
    transition: transform 0.15s ease, box-shadow 0.15s ease;
  }

  .face:hover {
    transform: translateY(-3px);
    box-shadow: 0 6px 16px rgba(23, 21, 42, 0.18);
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
    color: var(--on-accent);
  }

  .face-name {
    max-width: 56px;
    font-size: 0.6875rem;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .stage-dock {
    flex-shrink: 0;
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    gap: 10px;
    padding: 14px 24px;
    background: var(--bg-elev);
    border-top: 1px solid var(--border);
  }

  .dock-left {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }

  .dock-nav {
    display: flex;
    align-items: center;
    gap: 10px;
    justify-self: center;
  }

  .dock-right {
    display: flex;
    justify-content: flex-end;
    min-width: 0;
  }

  /* Navegação ←/→ da pergunta: os dois botões iguais, no estilo secundário
     (mesmo visual do "Reiniciar pergunta" neste dock). */
  .nav-btn {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-control);
    border: 1px solid var(--border-strong);
    background: var(--bg-elev);
    color: var(--text);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    cursor: pointer;
    transition: background 0.15s ease, border-color 0.15s ease, opacity 0.15s ease;
  }

  .nav-btn:hover:not(:disabled) {
    border-color: var(--accent);
    background: var(--surface-muted);
  }

  .nav-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .dock-counter {
    min-width: 56px;
    text-align: center;
    font-size: 0.875rem;
    font-weight: 800;
    letter-spacing: 0.04em;
  }

  .present-modes {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .mode-btn {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 800;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;
  }

  .mode-btn:hover {
    background: var(--surface-muted);
    color: var(--text);
    border-color: var(--accent);
  }

  .mode-btn.active {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
  }

  /* --- Trilho --- */

  .stage-rail {
    width: 330px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    background: var(--surface-muted);
    border-left: 1px solid var(--border);
    overflow: hidden;
  }

  @media (max-width: 900px) {
    .stage-layout {
      flex-direction: column;
    }

    .stage-rail {
      width: 100%;
      max-height: 60vh;
      border-left: none;
      border-top: 1px solid var(--border);
    }

    .stage-dock {
      grid-template-columns: 1fr;
      justify-items: center;
    }

    .dock-right {
      justify-content: center;
    }
  }

  .rail-panel {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 14px 12px;
    overflow-y: auto;
    /* offsetParent do question-row: sem isso, offsetTop fica num sistema de
       coordenadas diferente do scrollTop e a rolagem automática quebra. */
    position: relative;
  }

  .rail-empty {
    margin: 0;
    font-size: 0.8125rem;
  }

  .question-row {
    display: flex;
    gap: 10px;
    padding: 10px 12px;
    border-radius: var(--radius-control);
    border: none;
    border-left: 2px solid transparent;
    background: transparent;
    text-align: left;
    font-family: var(--font-ui);
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .question-row:hover {
    background: var(--bg-elev);
  }

  .question-row.active {
    background: var(--bg-elev);
    border-left-color: var(--accent);
  }

  .question-n {
    flex-shrink: 0;
    font-size: 0.75rem;
    font-weight: 800;
    color: var(--text-subtle);
  }

  .question-info {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 5px;
    min-width: 0;
  }

  .question-title-text {
    font-size: 0.8125rem;
    line-height: 1.3;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
  }

  .question-row.active .question-title-text {
    color: var(--text);
  }

  .qa-item {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
    padding: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
  }

  .qa-item-author {
    font-size: 0.6875rem;
    color: var(--text-subtle);
  }

  .qa-item-text {
    font-size: 0.8125rem;
    line-height: 1.4;
    word-break: break-word;
  }

  .qa-dismiss {
    padding: 5px 10px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.6875rem;
    font-weight: 700;
    cursor: pointer;
  }

  .qa-dismiss:hover {
    border-color: var(--accent);
    color: var(--text);
  }

  .notice-form {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .notice-actions {
    display: flex;
    gap: 8px;
  }

  .rail-switches {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
    border-top: 1px solid var(--border);
  }

  .control-row {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 0.8125rem;
    color: var(--text-muted);
  }

  .control-row span.on {
    color: var(--text);
  }

  .rail-footer {
    flex-shrink: 0;
    padding: 14px 16px;
    border-top: 1px solid var(--border);
  }
</style>
