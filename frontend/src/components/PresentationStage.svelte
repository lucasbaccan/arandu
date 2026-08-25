<script>
  import { onDestroy, tick } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { flip } from 'svelte/animate';

  // pending: participantes ainda não revelados pra pergunta atual (id, email, photo).
  export let pending = [];
  // groups: buckets já visíveis desde o início (opções da pergunta, ou
  // respostas abertas distintas) — { label, participants: [...] }.
  export let groups = [];
  // onFaceClick: se informado, os rostos pendentes viram botão clicável
  // (tela do admin). Se omitido, ficam só leitura (tela pública).
  export let onFaceClick = null;
  // hideZones: esconde as zonas de resposta (opções e quem foi revelado
  // nelas) — usado pela plateia quando o organizador ativa "esconder
  // respostas". A fila de pendentes continua visível independente disso.
  export let hideZones = false;
  // showNames: mostra o nome junto de cada rosto, pendente ou revelado —
  // controlado pelo switch "Ocultar nomes" (estado do servidor), refletido
  // tanto na janela de apresentação do organizador (StagePresentation)
  // quanto na tela da plateia (Audience). O painel do próprio organizador
  // (/stage) não usa este componente e sempre mostra os nomes.
  export let showNames = false;
  /*
   * Dois papéis, duas escalas:
   * - 'screen': projeção. Placar com densidade automática (abaixo).
   * - 'compact': celular de quem assiste. Linhas fixas — quem rola é a tela
   *   (.phone-body), então não precisa caber de uma vez.
   */
  export let layout = 'screen';
  /*
   * forceCols: força o nº de colunas do placar (1–4) — usado pelo menu ⚙ da
   * tela de apresentação (/present1..4) pra comparar. 0 = automático
   * (/present): fitDensity escolhe o menor nº de colunas que cabe sem
   * cortar ninguém. Em ambos os casos a escala das pílulas cresce ou
   * encolhe pra usar o máximo de espaço possível nesse nº de colunas (ver
   * comentário mais abaixo) — não é um tamanho fixo. O celular (.compact)
   * ignora.
   */
  export let forceCols = 0;
  /*
   * smart: modo "Smart" do menu ⚙ (/presentsmart). Em vez de ficar com o
   * menor nº de colunas que já cabe (comportamento padrão de forceCols=0),
   * testa cada nº de colunas (1–4) e fica com o que render a MAIOR escala
   * — às vezes isso pede mais colunas que o Auto escolheria, mas resulta em
   * pílulas maiores. Cai pro mesmo corte com "+N" do modo padrão se nem a
   * escala mínima couber em nenhum candidato.
   */
  export let smart = false;
  /*
   * pendingScroll: no painel do organizador (/stage) a fila de pendentes
   * vira UMA linha com scroll horizontal mostrando todos (sem teto
   * PENDING_CAP, sem chip "+N") — o organizador navega a lista rolando na
   * mão, mas o placar de respostas abaixo continua sem rolagem (fitDensity).
   * O telão de projeção mantém o padrão: uma linha só, teto de rostos
   * visíveis, resto em "+N", sem scroll.
   */
  export let pendingScroll = false;
  /*
   * scrollFallback: usado pelo painel do organizador (/stage). O placar
   * tenta encaixar tudo na densidade automática (1–4 colunas, escala até
   * MIN_SCALE); se mesmo assim não couber, em vez de cortar com "+N" como o
   * telão faz, mostra TODAS as pílulas na escala mínima e deixa o placar
   * rolar verticalmente (o organizador pode rolar; tela projetada não).
   */
  export let scrollFallback = false;
  /*
   * Dicas ao passar o mouse (painel do organizador /stage; desligado nas
   * telas públicas). Com hoverHints ligado, um tooltip com atraso de
   * HOVER_DELAY (0,6s) aparece:
   * - sobre um rosto (pendente ou revelado): a resposta daquela pessoa pra
   *   pergunta atual (via answerTextFor);
   * - sobre o rótulo de uma resposta: quem está nela e quem AINDA vai cair
   *   nela quando for revelado — zoneAll traz revelados + pendentes; os
   *   pendentes aparecem com borda tracejada no tooltip e a fila lá em cima
   *   ganha um anel de destaque, ligando os dois lados da tela numa olhada.
   * O atraso existe pra o mouse poder passear pelo placar sem o tooltip
   * abrir/fechar a cada pixel: só "decide" mostrar depois de parar 0,6s em
   * cima de algo (e fecha na hora quando sai).
   */
  export let hoverHints = false;
  // answerTextFor(p) → texto legível da resposta de p (ou null se não tem).
  export let answerTextFor = null;
  // zoneAll → Map rótulo da resposta → participantes (revelados E pendentes).
  export let zoneAll = null;
  // onRevealZone(label): mini botão ⚡ do popup de zona — revela TODOS os
  // pendentes DAQUELA resposta (os listados no popup), não a pergunta inteira.
  // O Stage implementa com o mesmo escalonamento do "Revelar tudo" do rodapé.
  export let onRevealZone = null;

  // Paleta da logo, na ordem das opções — a mesma cor identifica a zona no
  // telão e no celular.
  const ZONE_COLORS = ['var(--cyan)', 'var(--pink)', 'var(--yellow)', 'var(--orange)'];

  $: totalParticipants =
    pending.length + groups.reduce((sum, g) => sum + g.participants.length, 0);

  // Primeiro nome + inicial do sobrenome: "Lucas Elias Baccan" vira
  // "Lucas B." — mais compacto que o nome completo (a pílula não estoura de
  // largura), mas ainda identifica melhor que só o primeiro nome quando tem
  // gente com o mesmo primeiro nome na sala. Nomes com uma palavra só (ou o
  // fallback pro e-mail, sem nome cadastrado) mostram só essa palavra.
  function displayName(p) {
    const raw = (p.name || p.email || '').trim();
    const parts = raw.split(/\s+/).filter(Boolean);
    if (!p.name || parts.length < 2) return parts[0] || raw;
    const last = parts[parts.length - 1];
    return `${parts[0]} ${last[0].toUpperCase()}.`;
  }

  function zoneColor(i) {
    return ZONE_COLORS[i % ZONE_COLORS.length];
  }

  /*
   * O telão (layout="screen") é projetado — ninguém rola uma tela projetada
   * durante a dinâmica. A fila de pendentes tem um teto fixo de rostos
   * visíveis e o resto vira um chip "+N"; o teto garante caber em 1366×768
   * (o piso: notebook/projetor mais comum). Com muitas zonas (7+), a fila
   * inteira encolhe (.tight) pra devolver ~35px de altura ao placar. O
   * painel do organizador (/stage) passa pendingScroll e ignora este teto:
   * lá a fila rola na horizontal e mostra todos os pendentes.
   */
  const PENDING_CAP = 14;

  // O chip "+N" ocupa um slot na mesma fileira — quando vai sobrar gente, o
  // teto de itens visíveis cede um lugar pra ele.
  function capWithChipRoom(count, cap) {
    if (count <= cap) return { visible: count, hidden: 0 };
    return { visible: cap - 1, hidden: count - (cap - 1) };
  }

  $: pendingSlots =
    layout === 'screen' && !pendingScroll ? capWithChipRoom(pending.length, PENDING_CAP) : null;
  $: visiblePending = pendingSlots ? pending.slice(0, pendingSlots.visible) : pending;
  $: hiddenPendingCount = pendingSlots ? pendingSlots.hidden : 0;
  // Depende só do nº de zonas (input estático), nunca de medida de tela —
  // senão o encolher da fila mudaria a altura medida do placar, que mudaria
  // a densidade, que mudaria a fila… (loop de layout).
  $: pendingTight = (layout === 'screen' && groups.length >= 7) || pendingScroll;

  /*
   * ---------- Placar com densidade automática ----------
   *
   * Cada resposta vira uma zona (rótulo + contagem, e os respondentes como
   * pílulas de rosto + primeiro nome). Como o telão não rola, o componente
   * mede o espaço real (bind:clientWidth/Height em .zones) e ESTIMA (fator
   * 0.6 × fonte pra largura de texto) se um nº de colunas cabe sem cortar
   * ninguém — e, se coubesse, até que escala (sizesFor) dá pra crescer sem
   * estourar. Não existe um tamanho de pílula fixo: com pouca gente as
   * pílulas expandem até o teto (MAX_SCALE) pra ocupar o espaço sobrando;
   * com muita, encolhem até o piso (MIN_SCALE) que ainda cabe tudo — cada
   * combinação de colunas tem a sua própria escala máxima (maxFeasibleScale).
   * Auto e forçado (1–4) pegam o candidato certo e usam a escala máxima
   * dele; Smart testa todos e fica com a maior escala entre eles (ver
   * fitDensity). Só se nem a escala mínima coubesse em nenhum candidato
   * (evento gigante) é que volta o chip "+N", com teto por zona. O
   * overflow:hidden em .zone-people é a rede de segurança pra qualquer erro
   * de arredondamento.
   */
  const MIN_SCALE = 0.3;
  // Teto do modo "smart" — alinhado ao rosto de 64px da fila de pendentes
  // (avatar 46×MAX_SCALE ≈ 64px), pra não crescer além do resto do telão.
  const MAX_SCALE = 1.4;
  const MAX_COLS = 4;
  const COL_GAP = 16;
  // Título da resposta NÃO escala com a densidade: fixo e legível — quem
  // cresce/encolhe conforme o espaço são só as pílulas. Mesmo valor do
  // font-size de .zone-text/.zone-count no CSS; mudou um, mude o outro.
  const LABEL_FONT = 20;

  // Tamanhos derivados da escala — os valores em s=1 são os mesmos dos
  // fallbacks das custom properties no CSS abaixo; mudou um, mude o outro.
  // Os pisos (Math.max) travam o "menor possível" mesmo em escalas abaixo
  // de MIN_SCALE (não deveria acontecer, mas evita pílula ilegível se
  // algum arredondamento escorregar pra fora do intervalo esperado).
  function sizesFor(s) {
    return {
      avatar: Math.max(20, Math.round(46 * s)),
      nameFont: Math.max(12, Math.round(16 * s)),
      pillPadX: Math.max(5, Math.round(10 * s)),
      pillPadY: Math.max(3, Math.round(5 * s)),
      pillGap: Math.max(5, Math.round(10 * s)),
      zonePadV: Math.max(6, Math.round(15 * s)),
      zonePadH: Math.max(8, Math.round(20 * s)),
      rowGap: Math.max(6, Math.round(14 * s)),
    };
  }

  function pillWidth(nameLen, z, withName) {
    if (!withName) return z.avatar;
    return z.avatar + 2 * z.pillPadX + 8 + Math.ceil(nameLen * z.nameFont * 0.6);
  }

  function pillHeight(z, withName) {
    return withName ? z.avatar + 2 * z.pillPadY + 3 : z.avatar;
  }

  // Quantas linhas de pílulas o grupo ocupa numa área de largura areaW
  // (mesmo algoritmo guloso do flex-wrap).
  function pillLines(group, areaW, z, withName) {
    if (group.participants.length === 0) return 0;
    let lines = 1;
    let x = 0;
    for (const p of group.participants) {
      const w = pillWidth(displayName(p).length, z, withName);
      if (x > 0 && x + z.pillGap + w > areaW) {
        lines += 1;
        x = w;
      } else {
        x += (x > 0 ? z.pillGap : 0) + w;
      }
    }
    return lines;
  }

  function labelHeight(label, labelW, maxLines) {
    const len = (label || '').length + 4; // +4 ≈ espaço da contagem
    const lines = Math.min(maxLines, Math.max(1, Math.ceil((len * LABEL_FONT * 0.55) / labelW)));
    return lines * Math.round(LABEL_FONT * 1.25);
  }

  // Largura útil pras pílulas de uma zona em (colunas, escala).
  function pillAreaW(W, cols, z, labelW) {
    if (cols === 1) return Math.max(80, W - labelW - 2 * z.zonePadH - 16);
    const colW = (W - COL_GAP * (cols - 1)) / cols;
    return Math.max(80, colW - 2 * z.zonePadH);
  }

  // Altura estimada do placar inteiro em (escala, colunas). cols=1: rótulo ao
  // lado das pílulas; cols>1: cartões com rótulo em cima — o grid alinha as
  // linhas, então cada linha custa o max dos vizinhos.
  function estimateTotal(gs, W, withName, cols, z, labelW) {
    const areaW = pillAreaW(W, cols, z, labelW);
    const heights = gs.map((g) => {
      const lines = pillLines(g, areaW, z, withName);
      const ph = pillHeight(z, withName);
      const peopleH = lines === 0 ? 0 : lines * ph + (lines - 1) * z.pillGap;
      if (cols === 1) {
        return Math.max(labelHeight(g.label, labelW, 3), peopleH) + 2 * z.zonePadV + 2;
      }
      return labelHeight(g.label, areaW, 2) + (peopleH ? peopleH + 8 : 0) + 2 * z.zonePadV + 2;
    });
    let total = 0;
    let rows = 0;
    for (let i = 0; i < heights.length; i += cols) {
      total += Math.max(...heights.slice(i, i + cols));
      rows += 1;
    }
    return total + (rows - 1) * z.rowGap;
  }

  // Busca binária pela MAIOR escala que ainda cabe (sem cortar ninguém) num
  // nº de colunas fixo. estimateTotal só cresce com a escala — nunca
  // diminui — então existe um único ponto de corte entre "cabe" e "não
  // cabe", e a busca binária converge pra ele. Retorna null se nem a
  // escala mínima coube nessas colunas.
  function maxFeasibleScale(gs, W, H, withName, cols, labelW) {
    if (estimateTotal(gs, W, withName, cols, sizesFor(MIN_SCALE), labelW) > H) return null;
    if (estimateTotal(gs, W, withName, cols, sizesFor(MAX_SCALE), labelW) <= H) return MAX_SCALE;
    let lo = MIN_SCALE;
    let hi = MAX_SCALE;
    for (let i = 0; i < 18; i++) {
      const mid = (lo + hi) / 2;
      if (estimateTotal(gs, W, withName, cols, sizesFor(mid), labelW) <= H) {
        lo = mid;
      } else {
        hi = mid;
      }
    }
    return lo;
  }

  function fitDensity(gs, W, H, withName, forced, smart, scrollFallback) {
    if (!W || !H || gs.length === 0) return null; // sem medida (1º frame/testes): CSS usa os fallbacks
    const labelW = Math.min(Math.max(200, W * 0.28), 460);
    const candidates = forced
      ? [Math.min(forced, Math.max(1, gs.length))]
      : [1, 2, 3, 4].filter((c) => c <= MAX_COLS && c <= Math.max(1, gs.length));

    // Escala sempre cresce até o maior valor que ainda cabe (maxFeasibleScale)
    // — nunca fica travada no tamanho mínimo, seja qual for o modo. O que
    // muda de um modo pro outro é só QUAL(IS) nº de colunas entram na
    // disputa e como se escolhe entre eles:
    if (smart) {
      // Modo "Smart": testa todos os candidatos e fica com a combinação
      // (colunas × escala) que usa o MAIOR espaço possível, não importa
      // quantas colunas isso peça — as respostas crescem até o teto
      // MAX_SCALE, sempre repartidas em colunas de largura igual.
      let bestSmart = null;
      for (const cols of candidates) {
        const s = maxFeasibleScale(gs, W, H, withName, cols, labelW);
        if (s !== null && (!bestSmart || s > bestSmart.s)) bestSmart = { s, cols, labelW, caps: null };
      }
      if (bestSmart) return bestSmart;
      // Nem a escala mínima coube em nenhum nº de colunas — cai pro mesmo
      // corte com "+N" do modo padrão, abaixo.
    } else {
      // Auto (0 colunas forçadas) prefere o menor nº de colunas que já
      // couber — testa 1, depois 2, etc., e fica no primeiro que tiver
      // alguma escala viável. Forçado (1–4) só tem esse candidato mesmo.
      // Em ambos os casos a escala desse nº de colunas é maximizada, não
      // fixa: com pouca gente ela expande pra ocupar o espaço sobrando; com
      // muita, encolhe até o ponto que ainda cabe tudo.
      for (const cols of candidates) {
        const s = maxFeasibleScale(gs, W, H, withName, cols, labelW);
        if (s !== null) return { s, cols, labelW, caps: null };
      }
    }

    // Nem na escala mínima coube. No painel do organizador (scrollFallback)
    // não se corta nada: mostra TODAS as pílulas na escala mínima, no nº de
    // colunas que estimar a MENOR altura total (menos rolagem), e o placar
    // rola verticalmente (ver .zones.scrollable) — quem rola é o
    // organizador, não a tela projetada.
    if (scrollFallback) {
      let best = null;
      const z0 = sizesFor(MIN_SCALE);
      for (const cols of candidates) {
        const total = estimateTotal(gs, W, withName, cols, z0, labelW);
        if (!best || total < best.total) best = { s: MIN_SCALE, cols, labelW, caps: null, total };
      }
      return best;
    }

    // Telão: pra cada nº de colunas candidato, raciona as linhas de pílulas
    // por zona e corta com "+N" — vence o arranjo que mostra mais gente no
    // total.
    let best = null;
    const z = sizesFor(MIN_SCALE);
    const lineH = pillHeight(z, withName) + z.pillGap;
    const labelH = Math.round(LABEL_FONT * 1.25);
    for (const cols of candidates) {
      const areaW = pillAreaW(W, cols, z, labelW);
      const rows = Math.ceil(gs.length / cols);
      const rowH = (H - (rows - 1) * z.rowGap) / rows;
      const inner = rowH - 2 * z.zonePadV - 2 - (cols > 1 ? labelH + 8 : 0);
      const lines = Math.max(1, Math.floor((inner + z.pillGap) / lineH));
      const caps = new Map();
      let shown = 0;
      for (const g of gs) {
        const avgW = g.participants.length
          ? g.participants.reduce((sum, p) => sum + pillWidth(displayName(p).length, z, withName), 0) /
            g.participants.length
          : areaW;
        const perLine = Math.max(1, Math.floor(areaW / (avgW + z.pillGap)));
        const cap = Math.max(2, lines * perLine);
        caps.set(g.label, cap);
        shown += Math.min(cap, g.participants.length);
      }
      if (!best || shown > best.shown) best = { s: MIN_SCALE, cols, labelW, caps, shown };
    }
    return best;
  }

  let zonesW = 0;
  let zonesH = 0;

  $: metrics = layout === 'screen'
    ? fitDensity(groups, zonesW, zonesH, showNames, forceCols, smart, scrollFallback)
    : null;
  $: cols = metrics ? metrics.cols : Math.max(1, forceCols || 1);
  $: zoneStyle = metrics
    ? (() => {
        const z = sizesFor(metrics.s);
        return (
          `--cols:${metrics.cols};--label-w:${Math.round(metrics.labelW)}px;` +
          `--avatar:${z.avatar}px;--name-font:${z.nameFont}px;` +
          `--pill-pad:${z.pillPadY}px ${z.pillPadX}px;--pill-gap:${z.pillGap}px;` +
          `--zone-pad:${z.zonePadV}px ${z.zonePadH}px;--row-gap:${z.rowGap}px;`
        );
      })()
    : '';

  $: visibleGroups = groups.map((g) => {
    const cap =
      layout === 'screen' && metrics && metrics.caps
        ? (metrics.caps.get(g.label) ?? Infinity)
        : Infinity;
    const slots =
      g.participants.length > cap
        ? capWithChipRoom(g.participants.length, cap)
        : { visible: g.participants.length, hidden: 0 };
    return { ...g, visible: g.participants.slice(0, slots.visible), hiddenCount: slots.hidden };
  });

  // ---- Tooltip das dicas de hover (ver comentário das props lá em cima) ----
  // O popup de ZONA é interativo: cada pessoa listada é um botão que chama
  // onFaceClick — o organizador revela/desrevela direto pelo hover da
  // pergunta, sem precisar caçar o rosto na fila (e o popup continua aberto
  // pra revelar várias seguidas). Por isso o fechamento ganhou um atraso
  // curto (HIDE_DELAY): ao mover o mouse do rótulo pro popup (ou do popup
  // de volta) o tooltip não some no meio do caminho.
  const HOVER_DELAY = 600;
  const HIDE_DELAY = 200;
  const TIP_PEOPLE_CAP = 12;
  let hoverTimer = null;
  let hideTimer = null;
  let hover = null; // { kind: 'person', participant, answerText } | { kind: 'zone', label, participants, anchor }
  let tipEl = null;
  let tipReady = false;
  let tipPos = { left: 0, top: 0 };

  function clearTimers() {
    if (hoverTimer) {
      clearTimeout(hoverTimer);
      hoverTimer = null;
    }
    if (hideTimer) {
      clearTimeout(hideTimer);
      hideTimer = null;
    }
  }

  function scheduleHover(kind, payload, e) {
    if (!hoverHints) return;
    clearHover();
    // O âncora é capturado na hora do evento: dentro do setTimeout o
    // currentTarget do evento já não vale mais (e o nó pode até ter saído).
    const anchor = e.currentTarget;
    hoverTimer = setTimeout(() => showHover(kind, payload, anchor), HOVER_DELAY);
  }

  function showHover(kind, payload, anchor) {
    hover = { kind, ...payload, anchor };
    positionTip(anchor);
  }

  function positionTip(anchor) {
    const rect = anchor.getBoundingClientRect();
    if (tipReady) {
      // Já visível (ex.: reancorou depois de revelar alguém pelo popup):
      // reposiciona direto, sem piscar.
      const tw = tipEl ? tipEl.offsetWidth : 0;
      const th = tipEl ? tipEl.offsetHeight : 0;
      tipPos = {
        left: Math.max(8, Math.min(rect.left + rect.width / 2 - tw / 2, window.innerWidth - tw - 8)),
        top: rect.top - th - 10 < 8 ? rect.bottom + 10 : rect.top - th - 10
      };
      return;
    }
    // Posição provisória (canto do âncora) antes de medir o tooltip — evita
    // piscar fora do lugar no frame em que ele monta.
    tipPos = { left: rect.left, top: rect.top };
    tick().then(() => {
      if (!tipEl || !hover) return;
      const tw = tipEl.offsetWidth;
      const th = tipEl.offsetHeight;
      tipPos = {
        left: Math.max(8, Math.min(rect.left + rect.width / 2 - tw / 2, window.innerWidth - tw - 8)),
        top: rect.top - th - 10 < 8 ? rect.bottom + 10 : rect.top - th - 10
      };
      tipReady = true;
    });
  }

  function clearHover() {
    clearTimers();
    hover = null;
    tipReady = false;
  }

  // Mouse saiu do âncora/popup: espera HIDE_DELAY antes de fechar, pra quem
  // está indo pro popup (ou voltando dele) não ver o tooltip sumir no meio.
  function scheduleHide() {
    clearTimers();
    hideTimer = setTimeout(() => {
      hideTimer = null;
      hover = null;
      tipReady = false;
    }, HIDE_DELAY);
  }

  // Mouse entrou no popup: cancela o fechamento agendado (fica aberto pra
  // dar tempo de clicar nas pessoas).
  function keepOpen() {
    if (hideTimer) {
      clearTimeout(hideTimer);
      hideTimer = null;
    }
  }

  function onFaceEnter(p, e) {
    scheduleHover(
      'person',
      { participant: p, answerText: answerTextFor ? answerTextFor(p) : null },
      e
    );
  }

  function onZoneLabelEnter(g, e) {
    const all = zoneAll ? zoneAll.get(g.label) : g.participants;
    scheduleHover('zone', { label: g.label, participants: all || [] }, e);
  }

  // Quem já está revelado nesta pergunta — o tooltip de zona usa isso pra
  // distinguir (com borda tracejada) quem ainda está pendente na fila e pra
  // rotular os botões de revelar/desrevelar do popup.
  $: revealedSet = new Set(groups.flatMap((g) => g.participants.map((p) => p.id)));

  // Contagem de pendentes reativa: revelar/desrevelar pelo popup atualiza o
  // rodapé sem precisar reabrir o tooltip.
  $: zoneHoverPendingCount =
    hover && hover.kind === 'zone'
      ? hover.participants.filter((p) => !revealedSet.has(p.id)).length
      : 0;

  // Rostos pendentes que vão cair na zona em hover: anel de destaque na fila
  // ligando o tooltip da resposta aos rostos que ainda estão lá em cima.
  $: hoveredZoneIds =
    hover && hover.kind === 'zone' ? new Set(hover.participants.map((p) => p.id)) : null;

  // O tooltip de PESSOA ancora num rosto que pode sumir/mudar quando o placar
  // muda (revelar/desrevelar, trocar de pergunta, reiniciar) — fecha na hora
  // em vez de ficar flutuando num ponto órfão da tela. O tooltip de ZONA fica
  // aberto (o rótulo da resposta não sai do lugar) e só se reancora. Compara
  // uma assinatura do placar (ids da fila + ids das zonas) em vez das
  // referências dos arrays, que o Svelte troca a cada re-render mesmo sem
  // mudança real.
  let boardSig = null;
  $: {
    const sig =
      pending.map((p) => p.id).join(',') + '|' +
      groups.map((g) => g.participants.map((p) => p.id).join(',')).join(';');
    if (boardSig !== null && sig !== boardSig) {
      if (hover && hover.kind === 'zone') {
        if (hover.anchor) positionTip(hover.anchor);
      } else if (hover && hover.kind === 'person') {
        clearHover();
      }
    }
    boardSig = sig;
  }

  // Trocar de pergunta troca o Map zoneAll por inteiro — o tooltip de zona
  // aponta pra lista antiga e ficaria órfão; nesse caso fecha em vez de
  // reancorar num rótulo que nem existe mais.
  $: if (
    hover && hover.kind === 'zone' && zoneAll && zoneAll.get(hover.label) !== hover.participants
  ) {
    clearHover();
  }

  // Desligar as dicas ("💡 Dicas" no CrumbBar) fecha qualquer tooltip aberto.
  $: if (!hoverHints) clearHover();

  onDestroy(() => clearHover());
</script>

<svelte:window on:resize={clearHover} />

{#if layout === 'screen' && pending.length === 0}
  <!-- Sem ninguém na fila a faixa vira uma linha de texto: no telão, aqueles
       76px de card vazio empurram a última zona para fora da tela. -->
  <p class="text-muted present-empty pending-note">
    {totalParticipants === 0 ? 'Ninguém respondeu ainda.' : 'Todas as respostas foram reveladas.'}
  </p>
{:else if layout === 'screen'}
  <div
    class="pending-row"
    class:tight={pendingTight}
    class:scroll={pendingScroll}
    on:scroll={clearHover}
  >
    {#each visiblePending as p (p.id)}
      <div
        class="face-wrap"
        role="group"
        class:zone-hinted={hoveredZoneIds && hoveredZoneIds.has(p.id)}
        animate:flip={{ duration: 350 }}
        out:fade={{ duration: 150 }}
        on:mouseenter={(e) => onFaceEnter(p, e)}
        on:mouseleave={scheduleHide}
      >
        <button
          type="button"
          class="face pending"
          class:static={!onFaceClick}
          disabled={!onFaceClick}
          title={hoverHints ? undefined : p.name || p.email}
          aria-label={onFaceClick ? "Revelar resposta de " + (p.name || p.email) : p.name || p.email}
          on:click={() => onFaceClick && onFaceClick(p)}
        >
          {#if p.photo}
            <img src={p.photo} alt="" />
          {:else}
            <span class="face-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
          {/if}
        </button>
        {#if showNames}
          <span class="face-name">{displayName(p)}</span>
        {/if}
      </div>
    {/each}
    {#if hiddenPendingCount > 0}
      <div class="face-wrap">
        <span class="face more-chip" title={`+${hiddenPendingCount} pendentes`}>+{hiddenPendingCount}</span>
      </div>
    {/if}
  </div>
{/if}

<div
  class="zones"
  class:compact={layout === 'compact'}
  class:multi={layout === 'screen' && cols > 1}
  class:scrollable={layout === 'screen' && scrollFallback}
  style={zoneStyle}
  bind:clientWidth={zonesW}
  bind:clientHeight={zonesH}
  on:scroll={clearHover}
>
  {#if hideZones}
    <p class="text-muted present-empty">O organizador escondeu as respostas por enquanto.</p>
  {:else}
    {#if groups.length === 0}
      <p class="text-muted present-empty">Ninguém respondeu ainda.</p>
    {/if}

    {#each visibleGroups as group, gi (group.label)}
      <div class="zone" style="--zone-color:{zoneColor(gi)}">
        <div
          class="zone-label"
          role="group"
          on:mouseenter={(e) => onZoneLabelEnter(group, e)}
          on:mouseleave={scheduleHide}
        >
          <span class="zone-text">{group.label}</span>
          <span class="zone-count">{group.participants.length}</span>
        </div>
        <div class="zone-people">
          {#each group.visible as p (p.id)}
            <div
              class="person-wrap"
              role="group"
              animate:flip={{ duration: 350 }}
              in:fly={{ y: -20, duration: 350 }}
              on:mouseenter={(e) => onFaceEnter(p, e)}
              on:mouseleave={scheduleHide}
            >
              <button
                type="button"
                class="person"
                class:face-only={!showNames}
                class:static={!onFaceClick}
                disabled={!onFaceClick}
                title={hoverHints ? undefined : p.name || p.email}
                aria-label={onFaceClick ? "Desrevelar resposta de " + (p.name || p.email) : p.name || p.email}
                on:click={() => onFaceClick && onFaceClick(p)}
              >
                <span class="person-face">
                  {#if p.photo}
                    <img src={p.photo} alt="" />
                  {:else}
                    <span class="face-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
                  {/if}
                </span>
                {#if showNames}
                  <span class="person-name">{displayName(p)}</span>
                {/if}
              </button>
            </div>
          {/each}
          {#if group.hiddenCount > 0}
            <span class="person more-pill" title={`+${group.hiddenCount} em "${group.label}"`}>
              +{group.hiddenCount}
            </span>
          {/if}
        </div>
      </div>
    {/each}
  {/if}
</div>

{#if hover}
  <div
    class="hover-tip"
    class:ready={tipReady}
    class:interactive={hover.kind === 'zone' && !!onFaceClick}
    role="tooltip"
    style="left:{tipPos.left}px;top:{tipPos.top}px"
    bind:this={tipEl}
    on:mouseenter={keepOpen}
    on:mouseleave={scheduleHide}
  >
    {#if hover.kind === 'person'}
      <div class="tip-person">
        {#if hover.participant.photo}
          <img class="tip-avatar" src={hover.participant.photo} alt="" />
        {:else}
          <span class="tip-avatar tip-avatar-placeholder">
            {(hover.participant.name || hover.participant.email)[0].toUpperCase()}
          </span>
        {/if}
        <div class="tip-person-info">
          <strong class="tip-name">{hover.participant.name || hover.participant.email}</strong>
          {#if hover.answerText}
            <span class="tip-answer">{hover.answerText}</span>
          {:else}
            <span class="tip-answer tip-answer-none">Sem resposta</span>
          {/if}
        </div>
      </div>
    {:else}
      <div class="tip-zone">
        <div class="tip-zone-head">
          <strong class="tip-zone-label">{hover.label}</strong>
          <span class="tip-zone-actions">
            <span class="tip-zone-count">{hover.participants.length}</span>
            <button
              type="button"
              class="tip-reveal-all"
              aria-label="Revelar todos desta resposta"
              title="Revelar todos desta resposta"
              disabled={!onRevealZone || zoneHoverPendingCount === 0}
              on:click={() => onRevealZone && onRevealZone(hover.label)}
            >⚡</button>
          </span>
        </div>
        {#if hover.participants.length === 0}
          <p class="text-muted tip-zone-empty">Ninguém respondeu isto ainda.</p>
        {:else}
          <div class="tip-people">
            {#each hover.participants.slice(0, TIP_PEOPLE_CAP) as p (p.id)}
              <button
                type="button"
                class="tip-person-pill"
                class:pending={!revealedSet.has(p.id)}
                class:static={!onFaceClick}
                disabled={!onFaceClick}
                aria-label={onFaceClick
                  ? (revealedSet.has(p.id)
                      ? 'Desrevelar resposta de ' + (p.name || p.email)
                      : 'Revelar resposta de ' + (p.name || p.email))
                  : undefined}
                title={onFaceClick
                  ? (revealedSet.has(p.id) ? 'Clique para desrevelar' : 'Clique para revelar')
                  : undefined}
                on:click={() => onFaceClick && onFaceClick(p)}
              >
                {#if p.photo}
                  <img class="tip-pill-avatar" src={p.photo} alt="" />
                {:else}
                  <span class="tip-pill-avatar tip-pill-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
                {/if}
                <span class="tip-pill-name">{displayName(p)}</span>
              </button>
            {/each}
            {#if hover.participants.length > TIP_PEOPLE_CAP}
              <span class="tip-more">+{hover.participants.length - TIP_PEOPLE_CAP}</span>
            {/if}
          </div>
          <span class="tip-zone-foot">
            {hover.participants.length}
            {hover.participants.length === 1 ? 'pessoa' : 'pessoas'} nesta resposta
            {#if zoneHoverPendingCount > 0}
              · {zoneHoverPendingCount} {zoneHoverPendingCount === 1 ? 'ainda pendente' : 'ainda pendentes'}
            {/if}
          </span>
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .present-empty {
    margin: 0;
    font-size: 0.9375rem;
  }

  .pending-note {
    flex-shrink: 0;
  }

  .pending-row {
    display: flex;
    /* nowrap de propósito no telão: é fila de projeção, não pode crescer em
       altura — o teto de PENDING_CAP mantém isso sempre cabendo numa linha
       só. O painel do organizador (pendingScroll) rola na horizontal. */
    flex-wrap: nowrap;
    overflow: hidden;
    gap: 12px;
    min-height: 76px;
    padding: 14px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    /* Nunca comprime a fila pra ceder espaço ao placar: quem encolhe é a
       área de zonas (flex:1 / min-height:0), não os pendentes. */
    flex-shrink: 0;
    /* Contém os avatares na própria fila: os rostos ganham camadas compostas
       (transform) durante o flip/fade e, em alguns navegadores, a camada
       escapa do recorte do overflow (especialmente com border-radius) —
       dá um flash deles sobre outros elementos da tela. will-change: transform
       composita a fila inteira, então o recorte é aplicado na própria camada
       do navegador (e não quebra o scroll horizontal). */
    will-change: transform;
  }

  /* Modo "scroll" (painel do organizador, /stage): uma linha só com scroll
     horizontal mostrando TODOS os pendentes (sem teto, sem "+N") — o placar
     abaixo (PresentationStage .zones, flex:1) mantém a altura estável e
     continua cabendo sem rolagem. */
  .pending-row.scroll {
    flex-wrap: nowrap;
    overflow-x: auto;
    overflow-y: hidden;
  }

  /* Fila de pendentes: rosto de 64px, nome a 15px — a fila é a "vitrine" de
     quem ainda vai ser revelado, então mantém escala grande; só encolhe
     (.tight) quando o placar tem 7+ zonas e cada pixel de altura conta. */
  .face-wrap {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    width: 76px;
  }

  .face-name {
    max-width: 76px;
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .face {
    width: 64px;
    height: 64px;
    flex-shrink: 0;
    border-radius: 50%;
    border: none;
    padding: 0;
    overflow: hidden;
    cursor: pointer;
    background: var(--accent);
    transition: transform 0.1s ease;
  }

  .pending-row.tight {
    min-height: 0;
    padding: 10px 12px;
    gap: 10px;
  }

  .pending-row.tight .face-wrap {
    width: 56px;
    gap: 4px;
  }

  .pending-row.tight .face {
    width: 44px;
    height: 44px;
  }

  .pending-row.tight .face-placeholder {
    font-size: 1.0625rem;
  }

  .pending-row.tight .face-name {
    max-width: 56px;
    font-size: 0.75rem;
  }

  .face:not(:disabled):hover {
    transform: scale(1.05);
  }

  .face.static {
    cursor: default;
    display: block;
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
    font-size: 1.5rem;
    font-weight: 800;
  }

  /* Pendente (ainda não revelado): só a borda tracejada marca a diferença —
     a "arte" do rosto (foto ou inicial) é a mesma de quando revelado, igual
     ao avatar do painel de respostas (ResponsesPanel), pra não ter dois
     estilos de avatar diferentes na mesma pessoa. */
  .face.pending {
    border: 2px dashed var(--border-strong);
  }

  .face.pending:not(:disabled):hover {
    border-color: var(--accent);
  }

  /* Chip "+N": mesmo tamanho do avatar, pra continuar a fileira sem quebrar o
     ritmo visual. */
  .more-chip {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--surface-muted);
    border: 1.5px dashed var(--border-strong);
    color: var(--text-muted);
    font-size: 1.0625rem;
    font-weight: 800;
    cursor: default;
  }

  /*
   * position:relative é OBRIGATÓRIO aqui: o bind:clientWidth/Height faz o
   * Svelte injetar um <iframe> position:absolute com height:100% dentro de
   * .zones. Sem um pai posicionado, o iframe ancora no ancestral posicionado
   * mais próximo (a página inteira) e pode estourar a altura do documento por
   * arredondamento — foi a origem de uma barra de rolagem fantasma no telão.
   * (O Svelte tentaria setar position:relative inline, mas o style= reativo
   * do componente sobrescreve o inline style a cada atualização.)
   */
  .zones {
    position: relative;
    flex: 1;
    min-height: 0;
    /* Mesma proteção da fila de pendentes: as pílulas de pessoa (fly/flip)
       ganham camadas compostas na transição e podem vazar do recorte —
       compositar o placar mantém tudo recortado na própria camada. */
    will-change: transform;
  }

  /* Os tamanhos vêm das custom properties setadas por fitDensity() — os
     fallbacks aqui são os valores da escala 1 (ver sizesFor), usados no 1º
     frame e nos testes. Telão nunca rola: overflow escondido é só rede de
     segurança pra erro de estimativa. */
  .zones:not(.compact) {
    display: flex;
    flex-direction: column;
    gap: var(--row-gap, 6px);
    overflow: hidden;
  }

  /* Painel do organizador (scrollFallback): se nem na escala mínima couber,
     o placar rola verticalmente em vez de cortar opções — overflow-x travado
     pra não aparecer scroll horizontal no placar. */
  .zones.scrollable {
    overflow-y: auto;
    overflow-x: hidden;
  }

  /* Anti-loop de barra de rolagem: o placar se mede por clientWidth/Height
     (bind abaixo) pra densidade automática caber sem rolagem — mas quando o
     conteúdo fica exatamente na borda de caber, a barra aparece, a largura
     medida encolhe (~15px), a densidade recalcula e troca o nº de colunas, a
     altura muda e a barra volta: ciclo sem fim (acontecia ao revelar gente,
     dependendo da resolução da tela e dos tamanhos das respostas).
     scrollbar-gutter: stable reserva o espaço da barra SEMPRE, então a
     largura não oscila com ela aparecer/sumir. Onde não há suporte (Safari
     antigo), overflow-y: scroll deixa a barra sempre presente — mesma
     estabilidade, ao custo de um trilho visível mesmo quando cabe tudo. */
  @supports (scrollbar-gutter: stable) {
    .zones.scrollable {
      scrollbar-gutter: stable;
    }
  }

  @supports not (scrollbar-gutter: stable) {
    .zones.scrollable {
      overflow-y: scroll;
    }
  }

  /* Com mais respostas, fitDensity (ou o menu ⚙) muda pra 2–4 colunas de
     cartões com rótulo em cima, em vez de encolher tudo numa coluna só. */
  .zones.multi:not(.compact) {
    display: grid;
    grid-template-columns: repeat(var(--cols, 2), 1fr);
    column-gap: 16px;
    row-gap: var(--row-gap, 6px);
    align-content: stretch;
  }

  .zone {
    display: flex;
    align-items: center;
    gap: 16px;
    min-width: 0;
    padding: var(--zone-pad, 6px 8px);
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-left: 5px solid var(--zone-color);
    border-radius: 12px;
  }

  .zone-label {
    flex: 0 0 var(--label-w, 30%);
    min-width: 0;
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 10px;
  }

  .zones.multi .zone {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .zones.multi .zone-label {
    flex: 0 0 auto;
  }

  /* Fixo de propósito (= LABEL_FONT no script): o título da resposta só
     precisa ser legível — não cresce em tela folgada nem encolhe na densa. */
  .zone-text {
    font-size: 20px;
    font-weight: 700;
    line-height: 1.25;
    /* Resposta aberta gigante corta em vez de estourar a linha. */
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .zones.multi .zone-text {
    -webkit-line-clamp: 2;
  }

  .zone-count {
    font-size: 20px;
    font-weight: 800;
    color: var(--zone-color);
  }

  .zone-people {
    flex: 1;
    min-width: 0;
    align-self: stretch;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    align-content: center;
    gap: var(--pill-gap, 5px);
    /* Rede de segurança pra erro de arredondamento na estimativa (ver
       comentário lá em cima) — nunca deveria disparar de verdade. Só
       overflow:hidden, sem máscara de desvanecer: como align-content:center
       centraliza as linhas de pílula na altura sobrando, um degradê fixo na
       borda inferior acabava desbotando a última linha inteira sempre que o
       conteúdo enchia a zona (sem overflow nenhum) — não só nesse caso raro. */
    overflow: hidden;
  }

  .zones.multi .zone-people {
    align-self: auto;
    align-content: flex-start;
  }

  .person-wrap {
    display: flex;
    min-width: 0;
  }

  /* Pílula de respondente: rosto + primeiro nome, lado a lado — bem mais
     densa que rosto com legenda embaixo, e a cor da zona a ancora na
     resposta certa. */
  .person {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: var(--pill-pad, 3px 5px);
    border-radius: 999px;
    background: color-mix(in srgb, var(--zone-color) 12%, var(--bg-elev));
    border: 1px solid color-mix(in srgb, var(--zone-color) 35%, transparent);
    cursor: pointer;
    transition: transform 0.1s ease;
  }

  .person:not(:disabled):hover {
    transform: scale(1.04);
  }

  .person.static {
    cursor: default;
  }

  /* "Ocultar nomes" ligado: sobra só o rosto redondo, sem moldura de pílula. */
  .person.face-only {
    padding: 0;
    border: none;
    background: none;
  }

  .person-face {
    width: var(--avatar, 20px);
    height: var(--avatar, 20px);
    flex-shrink: 0;
    border-radius: 50%;
    overflow: hidden;
    background: var(--accent);
  }

  .person-face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .person-face .face-placeholder {
    font-size: calc(var(--avatar, 20px) * 0.42);
  }

  .person-name {
    font-size: var(--name-font, 12px);
    font-weight: 600;
    color: var(--text);
    white-space: nowrap;
  }

  .person.more-pill {
    background: var(--surface-muted);
    border: 1.5px dashed var(--border-strong);
    color: var(--text-muted);
    font-size: var(--name-font, 12px);
    font-weight: 800;
    padding: var(--pill-pad, 3px 5px);
    cursor: default;
  }

  /* --- Celular de quem assiste: placar em linhas, tamanhos fixos. --- */

  .zones.compact {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .zones.compact .zone {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
    padding: 12px 14px;
    border-left-width: 3px;
    border-radius: var(--radius-row);
  }

  .zones.compact .zone-label {
    flex: 0 0 auto;
  }

  .zones.compact .zone-text,
  .zones.compact .zone-count {
    font-size: 0.9375rem;
  }

  .zones.compact .zone-text {
    font-weight: 600;
  }

  .zones.compact .zone-people {
    align-self: auto;
    align-content: flex-start;
    overflow: visible;
    gap: 6px;
  }

  .zones.compact .zone-people:empty {
    display: none;
  }

  .zones.compact .person {
    gap: 6px;
    padding: 3px 8px;
  }

  .zones.compact .person.face-only {
    padding: 0;
  }

  .zones.compact .person-face {
    width: 28px;
    height: 28px;
  }

  .zones.compact .person-face .face-placeholder {
    font-size: 0.75rem;
  }

  .zones.compact .person-name,
  .zones.compact .person.more-pill {
    font-size: 0.8125rem;
  }

  /* --- Tooltip das dicas de hover (painel do organizador, /stage) --- */

  /* position:fixed + pointer-events:none: segue a tela e nunca atrapalha o
     hover de quem está embaixo. Opacity 0 até posicionar (tipReady) pra não
     piscar no canto do âncora enquanto mede o próprio tamanho. */
  .hover-tip {
    position: fixed;
    z-index: 60;
    max-width: 320px;
    padding: 10px 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-control);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
    pointer-events: none;
    opacity: 0;
    transform: translateY(-4px);
    transition: opacity 0.12s ease, transform 0.12s ease;
  }

  .hover-tip.ready {
    opacity: 1;
    transform: translateY(0);
  }

  /* Só o popup de zona é interativo (dá pra clicar nas pessoas pra
     revelar/desrevelar); o de pessoa continua só leitura. */
  .hover-tip.interactive {
    pointer-events: auto;
  }

  /* Rosto → resposta da pessoa */
  .tip-person {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .tip-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;
    background: var(--accent);
  }

  .tip-avatar-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--on-accent);
    font-size: 0.875rem;
    font-weight: 800;
  }

  .tip-person-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .tip-name {
    font-size: 0.8125rem;
    line-height: 1.3;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tip-answer {
    font-size: 0.875rem;
    font-weight: 700;
    color: var(--accent);
    word-break: break-word;
  }

  /* O prefixo "Resposta:" vem via ::before pra o texto do tooltip continuar
     um nó de texto só (o matcher de texto dos testes bate direto nele). */
  .tip-answer:not(.tip-answer-none)::before {
    content: 'Resposta: ';
    color: var(--text-muted);
    font-weight: 600;
  }

  .tip-answer-none {
    color: var(--text-subtle);
    font-weight: 600;
  }

  /* Resposta → quem está (e quem vai cair) nela */
  .tip-zone-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 8px;
  }

  .tip-zone-label {
    font-size: 0.875rem;
    font-weight: 800;
    line-height: 1.25;
    word-break: break-word;
  }

  .tip-zone-count {
    font-size: 0.8125rem;
    font-weight: 800;
    color: var(--text-muted);
  }

  .tip-zone-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  /* Mini botão ⚡ do popup: revela todos os pendentes da pergunta de uma vez
     (ícone somente; o título/aria-label explicam). */
  .tip-reveal-all {
    width: 24px;
    height: 24px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-size: 0.8125rem;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    transition: border-color 0.15s ease, color 0.15s ease, background 0.15s ease;
  }

  .tip-reveal-all:not(:disabled):hover {
    border-color: var(--accent);
    color: var(--accent);
    background: var(--surface-muted);
  }

  .tip-reveal-all:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .tip-people {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  /* Pílula de pessoa no tooltip: pendente (ainda na fila) fica com borda
     tracejada, igual aos rostos pendentes do placar. É um botão: clicar
     revela/desrevela direto pelo popup (painel do organizador). */
  .tip-person-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 3px 8px 3px 3px;
    border-radius: 999px;
    background: var(--surface-muted);
    border: 1px solid var(--border);
    color: var(--text);
    font-family: var(--font-ui);
    font-size: 0.75rem;
    max-width: 170px;
    cursor: pointer;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .tip-person-pill:not(:disabled):hover {
    border-color: var(--accent);
    background: var(--bg-elev);
  }

  .tip-person-pill.static,
  .tip-person-pill:disabled {
    cursor: default;
  }

  .tip-person-pill.pending {
    border-style: dashed;
    border-color: var(--border-strong);
  }

  /* Além da borda tracejada: o nome de quem ainda não foi revelado fica
     esmaecido (ainda não apareceu no placar); o já revelado mantém a cor
     cheia herdada de .tip-person-pill. */
  .tip-person-pill.pending .tip-pill-name {
    color: var(--text-subtle);
  }

  .tip-pill-avatar {
    width: 22px;
    height: 22px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;
    background: var(--accent);
  }

  .tip-pill-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--on-accent);
    font-size: 0.6875rem;
    font-weight: 800;
  }

  .tip-pill-name {
    font-size: 0.75rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tip-more {
    display: inline-flex;
    align-items: center;
    padding: 3px 8px;
    border-radius: 999px;
    border: 1.5px dashed var(--border-strong);
    color: var(--text-muted);
    font-size: 0.75rem;
    font-weight: 800;
  }

  .tip-zone-foot {
    display: block;
    margin-top: 8px;
    font-size: 0.6875rem;
    color: var(--text-subtle);
  }

  .tip-zone-empty {
    margin: 0;
    font-size: 0.8125rem;
  }

  /* Rostos pendentes que vão cair na zona em hover ganham um anel, ligando
     o tooltip da resposta à fila lá em cima. */
  .face-wrap.zone-hinted .face {
    box-shadow: 0 0 0 3px var(--accent);
  }

</style>
