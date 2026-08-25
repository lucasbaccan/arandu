<script>
  export let id = '';

  import { onDestroy, tick } from 'svelte';
  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import { questionKindInfo } from '../lib/eventStatus.js';
  import Button from '../components/Button.svelte';
  import Chip from '../components/Chip.svelte';
  import CrumbBar from '../components/CrumbBar.svelte';
  import Input from '../components/Input.svelte';
  import PinChip from '../components/PinChip.svelte';
  import PresentationStage from '../components/PresentationStage.svelte';
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
  // de apresentação (ou a tela pública /plateia) aberta.
  let modoDensidadeApresentacao = '';

  // Dicas ao passar o mouse no placar (painel do organizador): tooltip com
  // atraso de 0,6s sobre um rosto (mostra a resposta da pessoa) ou sobre o
  // rótulo de uma resposta (mostra quem está e quem AINDA vai cair ali).
  // Ligado/desligado pelo botão "💡 Dicas" no CrumbBar, ao lado do PIN.
  let hoverHints = true;

  function toggleHoverHints() {
    hoverHints = !hoverHints;
  }

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
    clearTimeout(resizeTimer);
  });

  // Modos de densidade do placar da apresentação (ver fitDensity em
  // PresentationStage.svelte) — o mesmo menu que antes era só a
  // engrenagem ⚙ dentro da janela de apresentação, agora como botões no
  // rodapé daqui. Cada clique manda o modo pro servidor (como
  // blanked/answersHidden) — quem já estiver com /apresentar ou /plateia
  // abertos vê a densidade trocar ao vivo, sem precisar recarregar nem
  // alternar pra tela do projetor. `mode` é o valor exato que o servidor
  // espera (ver validPresentDensityModes no backend).
  const MODOS_APRESENTACAO = [
    { mode: '', label: 'Automático', icon: 'A' },
    { mode: 'smart', label: 'Smart', icon: '★' },
    { mode: '1', label: '1 coluna', icon: '1' },
    { mode: '2', label: '2 colunas', icon: '2' },
    { mode: '3', label: '3 colunas', icon: '3' },
    { mode: '4', label: '4 colunas', icon: '4' }
  ];

  function selectPresentMode(mode) {
    modoDensidadeApresentacao = mode;
    api.eventos.aoVivo.definirModoDensidade(id, mode).catch(() => {});
    openPresentationWindow();
  }

  // Referência da janela de apresentação (somente leitura, sem clique, sem
  // controles de admin — só a pergunta, as opções e os participantes; feita
  // pra projetar ou compartilhar numa chamada sem expor o painel do
  // organizador). Guardar a referência deixa reaproveitar a MESMA janela em
  // vez de abrir uma nova a cada clique em "Modo apresentação" ou num botão
  // de modo — a densidade em si não depende mais da URL (ver
  // modoDensidadeApresentacao acima), então aqui só importa abrir/focar.
  let janelaApresentacao = null;

  function openPresentationWindow() {
    if (janelaApresentacao && !janelaApresentacao.closed) {
      janelaApresentacao.focus();
      return;
    }
    // Passar "features" (largura/altura) faz o navegador abrir uma janela
    // de verdade (sem abas, sem barra de endereço) em vez de só uma nova
    // aba. 1366×768: a resolução nativa mais comum de notebook/projetor —
    // o conteúdo em si é feito pra caber nesse piso sem rolagem (ver
    // StagePresentation.svelte); isso só evita abrir menor que isso por
    // padrão. Pra projetar de verdade, dá F11 na janela.
    janelaApresentacao = window.open(`/palco/${id}/apresentar`, `arandu-present-${id}`, 'width=1366,height=768');
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
        api.eventos.buscar(id),
        api.eventos.perguntas.listar(id),
        api.eventos.respostas.listar(id),
        api.eventos.aoVivo.estadoAdmin(id)
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
      modoDensidadeApresentacao = adminSnap.presentDensityMode || '';

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
  // /plateia/:id ver a mesma coisa em tempo real.
  function syncQuestion(questionId) {
    api.eventos.aoVivo.definirPergunta(id, questionId).catch(() => {});
  }

  // Conexão só de leitura: traz de volta blank/aviso/interações se a página
  // recarregar, e entrega as reações e mensagens de Q&A que chegam ao vivo.
  function connectAdminStream() {
    if (adminEventSource) adminEventSource.close();
    adminEventSource = new EventSource(api.eventos.aoVivo.urlFluxoAdmin(id));
    adminEventSource.onmessage = (e) => {
      const snap = JSON.parse(e.data);
      blanked = snap.blanked;
      answersHidden = snap.answersHidden;
      namesHidden = snap.namesHidden;
      interactionsEnabled = snap.interactionsEnabled;
      message = snap.message;
      messageDraft = snap.message;
      qaInbox = snap.qaInbox;
      modoDensidadeApresentacao = snap.presentDensityMode || '';
    };
    adminEventSource.addEventListener('reaction', (e) => {
      fireReaction(JSON.parse(e.data).emoji);
    });
  }

  function toggleBlanked() {
    blanked = !blanked;
    api.eventos.aoVivo.definirEmBranco(id, blanked).catch(() => {});
  }

  function toggleAnswersHidden() {
    answersHidden = !answersHidden;
    api.eventos.aoVivo.ocultarRespostas(id, answersHidden).catch(() => {});
  }

  function toggleInteractions() {
    interactionsEnabled = !interactionsEnabled;
    api.eventos.aoVivo.definirInteracoes(id, interactionsEnabled).catch(() => {});
  }

  // Estado do servidor (como os outros switches) — afeta a legenda de nome
  // sob cada rosto na janela de apresentação (/palco/:id/apresentar) e na tela
  // da plateia (/plateia/:id); esta tela (/palco) sempre mostra os nomes,
  // independente disso.
  function toggleNamesHidden() {
    namesHidden = !namesHidden;
    api.eventos.aoVivo.ocultarNomes(id, namesHidden).catch(() => {});
  }

  function sendMessage() {
    message = messageDraft.trim();
    api.eventos.aoVivo.definirMensagem(id, message).catch(() => {});
  }

  function clearMessage() {
    message = '';
    messageDraft = '';
    api.eventos.aoVivo.definirMensagem(id, '').catch(() => {});
  }

  function dismissQA(messageId) {
    qaInbox = qaInbox.filter((m) => m.id !== messageId);
    api.eventos.aoVivo.dispensarPergunta(id, messageId).catch(() => {});
  }

  function openAudienceScreen() {
    // A tela da plateia é única agora (sem modo telão/celular separado) —
    // basta o PIN pra entrar; quem abrir na janela de projeção vê a mesma
    // tela, maior.
    window.open(
      `/plateia/${id}?pin=${encodeURIComponent(event.pinCode.toUpperCase())}`,
      '_blank'
    );
  }

  $: currentQuestion = questions[currentIndex];
  $: revealedIds = currentQuestion ? revealed[currentQuestion.id] : new Set();

  // Título longo: encolhe a fonte até caber em ~TITLE_MAX_LINES (com folga),
  // pra pergunta gigante não empurrar o placar pra fora da tela nem criar
  // scroll. A fonte é multiplicada por --title-scale (ver
  // .palco-question-title no CSS); pergunta curta mantém escala 1. Reavalia
  // ao trocar de pergunta e ao redimensionar a janela.
  let titleEl = null;
  let titleScale = 1;
  const TITLE_MAX_LINES = 3;
  const TITLE_SCALE_MIN = 0.7;

  $: if (currentQuestion) fitQuestionTitle();

  function fitQuestionTitle() {
    tick().then(measureTitle);
  }

  function measureTitle() {
    if (!titleEl) return;
    const cs = getComputedStyle(titleEl);
    const lineH = parseFloat(cs.lineHeight) || 20;
    const lines = titleEl.offsetHeight / lineH;
    if (lines <= TITLE_MAX_LINES) {
      titleScale = 1;
      return;
    }
    // O nº de linhas cai ~proporcional à fonte: fator com folga de 5% pra
    // não oscilar na borda (linha que "quase cabe" não fica pulando entre
    // dois tamanhos a cada resize).
    titleScale = Math.max(TITLE_SCALE_MIN, (TITLE_MAX_LINES / lines) * titleScale * 0.95);
  }

  let resizeTimer;
  function handleWindowResize() {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(fitQuestionTitle, 150);
  }

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

  // Pro tooltip de zona (hoverHints): TODOS os participantes por resposta
  // (revelados + pendentes) — mostra quem já está na zona e quem ainda vai
  // aparecer nela quando for revelado. Mesmo agrupamento do placar, então os
  // rótulos batem 1:1 com as zonas (opção de múltipla escolha ou resposta
  // aberta normalizada).
  $: zoneAll = currentQuestion ? buildZoneAll(participants, currentQuestion) : new Map();

  function buildZoneAll(allParticipants, q) {
    const m = new Map();
    if (q.type === 'OPEN_TEXT') {
      for (const p of allParticipants) {
        const label = normalizeOpenText(answerFor(p, q)?.text) || '—';
        if (!m.has(label)) m.set(label, []);
        m.get(label).push(p);
      }
    } else {
      for (const opt of q.options) {
        m.set(
          opt.text,
          allParticipants.filter((p) => {
            const a = answerFor(p, q);
            return a && a.optionId === opt.id;
          })
        );
      }
    }
    return m;
  }

  // Texto legível da resposta de p pra pergunta atual (tooltip de rosto).
  function answerTextFor(p) {
    if (!currentQuestion) return null;
    const a = answerFor(p, currentQuestion);
    if (!a) return null;
    if (currentQuestion.type === 'OPEN_TEXT') {
      return normalizeOpenText(a.text) || '—';
    }
    return a.optionText || '—';
  }

  function reveal(p) {
    if (!currentQuestion) return;
    const questionId = currentQuestion.id;
    const set = revealed[questionId];
    if (set.has(p.id)) {
      set.delete(p.id);
      revealed = { ...revealed };
      api.eventos.aoVivo.ocultar(id, questionId, p.id).catch(() => {});
    } else {
      set.add(p.id);
      revealed = { ...revealed };
      api.eventos.aoVivo.revelar(id, questionId, p.id).catch(() => {});
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
        api.eventos.aoVivo.revelar(id, questionId, p.id).catch(() => {});
      }, i * 150);
    });
  }

  // ⚡ do popup de hover: revela só os pendentes DAQUELA resposta (zona),
  // não a pergunta inteira — o botão do rodapé "Revelar tudo" continua sendo
  // o caminho pra revelar tudo de uma vez.
  function revealZone(label) {
    if (!currentQuestion) return;
    const set = revealed[currentQuestion.id];
    const questionId = currentQuestion.id;
    const members = zoneAll.get(label) || [];
    members
      .filter((p) => !set.has(p.id))
      .forEach((p, i) => {
        setTimeout(() => {
          set.add(p.id);
          revealed = { ...revealed };
          api.eventos.aoVivo.revelar(id, questionId, p.id).catch(() => {});
        }, i * 150);
      });
  }

  function resetReveal() {
    if (!currentQuestion) return;
    revealed = { ...revealed, [currentQuestion.id]: new Set() };
    api.eventos.aoVivo.reiniciar(id, currentQuestion.id).catch(() => {});
  }

  // Reseta a revelação de TODAS as perguntas de uma vez (diferente do
  // "Reiniciar pergunta" do rodapé, que só afeta a pergunta atual) — pra
  // recomeçar a apresentação inteira do zero.
  function resetAllReveals() {
    revealed = Object.fromEntries(questions.map((q) => [q.id, new Set()]));
    api.eventos.aoVivo.reiniciarTudo(id).catch(() => {});
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

<svelte:window on:keydown={handleKeydown} on:resize={handleWindowResize} />

<ReactionBurstLayer />

<main class="shell">
  {#if loading}
    <div class="palco-center"><p class="text-muted">Carregando…</p></div>
  {:else if error}
    <div class="palco-center"><p class="form-error">{error}</p></div>
  {:else}
    <TopBar area="Organizador" />

    <CrumbBar
      crumbs={[
        { label: 'Eventos', href: '/painel' },
        { label: event.title, href: `/evento/${id}` },
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
        <button
          type="button"
          class="hint-toggle"
          class:on={hoverHints}
          aria-pressed={hoverHints}
          title="Dicas ao passar o mouse: mostra a resposta de cada pessoa e quem escolheu cada resposta (aparecem após 0,6 s)"
          on:click={toggleHoverHints}
        >
          <span class="hint-toggle-icon" aria-hidden="true">💡</span>
          <span>Dicas</span>
        </button>
        <Button variant="secondary" size="sm" on:click={resetAllReveals} disabled={!anyRevealed}>
          Reiniciar tudo
        </Button>
        <Button size="sm" on:click={openPresentationWindow}>Modo apresentação</Button>
      </svelte:fragment>
    </CrumbBar>

    {#if questions.length === 0}
      <div class="palco-center">
        <p class="text-muted">Este evento ainda não tem perguntas.</p>
      </div>
    {:else}
      <div class="palco-layout">
        <div class="palco-main">
          <div class="palco-main-body">
            <h2
              class="palco-question-title"
              bind:this={titleEl}
              style="--title-scale:{titleScale}"
            >{currentQuestion.title}</h2>

            <!-- Placar de respostas: a MESMA densidade automática do telão
                 (PresentationStage layout="screen", ver fitDensity) — mede o
                 espaço real, usa 1–4 colunas e encolhe as pílulas até TODAS
                 as respostas caberem na tela sem barra de rolagem.
                 scrollFallback: se mesmo na escala mínima não couber (evento
                 gigante), mostra tudo e deixa o placar rolar verticalmente
                 em vez de cortar opções com "+N". pendingScroll mantém a
                 lista de participantes numa linha com scroll horizontal
                 (nenhum pendente escondido atrás de "+N"), sem afetar o
                 placar. -->
            <div class="pending-head">
              <span class="overline">Pendentes · clique para revelar</span>
              <span class="overline">{pending.length}</span>
            </div>

            <PresentationStage
              layout="screen"
              {pending}
              {groups}
              onFaceClick={reveal}
              showNames
              pendingScroll
              scrollFallback
              {hoverHints}
              {answerTextFor}
              {zoneAll}
              onRevealZone={revealZone}
            />
          </div>

          <!-- Ações de revelação à esquerda, navegação de pergunta ao centro,
               modos da janela de apresentação à direita. -->
          <div class="palco-dock">
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
              <div class="modos-apresentacao">
                {#each MODOS_APRESENTACAO as mode (mode.mode)}
                  <button
                    type="button"
                    class="mode-btn"
                    class:active={modoDensidadeApresentacao === mode.mode}
                    title={`Modo apresentação — ${mode.label}`}
                    aria-label={`Modo apresentação — ${mode.label}`}
                    aria-pressed={modoDensidadeApresentacao === mode.mode}
                    on:click={() => selectPresentMode(mode.mode)}
                  >{mode.icon}</button>
                {/each}
              </div>
            </div>
          </div>
        </div>

        <div class="palco-rail">
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

  .palco-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .palco-layout {
    flex: 1;
    min-height: 0;
    display: flex;
  }

  .palco-main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    /* Isola o palco num contexto de empilhamento próprio (z-0): os avatares
       do pending-row e as transições flip/fly das zonas ganham camadas
       compostas (transform) durante a animação e, sem isso, pintam POR CIMA
       de elementos estáticos — dava pra ver rostos "vazando" sobre o trilho
       por um flash quando a fila recalculava o tamanho. */
    position: relative;
    z-index: 0;
  }

  .palco-main-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 18px 20px 16px;
    /* O placar abaixo (PresentationStage layout="screen") se mede e se
       encolhe pra caber inteiro — as respostas nunca precisam de scroll.
       overflow fica só como rede de segurança pra casos extremos (ex: título
       absurdamente longo), onde rolar é melhor que cortar. */
    overflow-y: auto;
    /* Isola todo o conteúdo animado do corpo (rostos da fila, pílulas das
       zonas) num único contexto de empilhamento — se qualquer camada ainda
       escapar do recorte, ela fica presa aqui dentro e não pinta sobre o
       dock, o trilho ou o resto da tela. (Não afeta o tooltip de hover, que
       é position:fixed — isolation não cria containing block pra ele.) */
    isolation: isolate;
  }

  .palco-question-title {
    margin: 0;
    /* Escala com a tela (clamp) e, quando a pergunta é longa demais (mais de
       TITLE_MAX_LINES), encolhe mais via --title-scale (ver fitQuestionTitle
       no script) pra caber em ~3 linhas e sobrar altura pro placar. Nunca
       corta: quebra em quantas linhas precisar, só que menor. */
    font-size: calc(clamp(1.25rem, 1.5vw + 1rem, 1.75rem) * var(--title-scale, 1));
    font-weight: 800;
    letter-spacing: -0.02em;
    line-height: 1.2;
  }

  .pending-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }

  .palco-dock {
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

  .modos-apresentacao {
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

  /* Botão "💡 Dicas" do CrumbBar (ao lado do PIN): alterna os tooltips de
     hover do placar. Mesmo tamanho dos botões sm, visual de ghost que vira
     accent quando ligado — igual aos botões de modo do rodapé. */
  .hint-toggle {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 32px;
    padding: 0 12px;
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 700;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;
  }

  .hint-toggle:hover {
    background: var(--surface-muted);
    color: var(--text);
    border-color: var(--accent);
  }

  .hint-toggle.on {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
  }

  .hint-toggle-icon {
    font-size: 0.9375rem;
    line-height: 1;
  }

  /* --- Trilho --- */

  .palco-rail {
    width: 330px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    background: var(--surface-muted);
    border-left: 1px solid var(--border);
    overflow: hidden;
    /* Trilho sempre por cima do palco (ver comentário em .palco-main): nada
       animado no palco — rostos do pending-row, tooltips, zonas — consegue
       pintar sobre a lista de perguntas. */
    position: relative;
    z-index: 1;
  }

  @media (max-width: 900px) {
    .palco-layout {
      flex-direction: column;
    }

    .palco-rail {
      width: 100%;
      max-height: 60vh;
      border-left: none;
      border-top: 1px solid var(--border);
    }

    .palco-dock {
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
