<script>
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
   * (/present): fitDensity escolhe quantas colunas cabem sem cortar
   * ninguém, sempre no tamanho mínimo de pílula (ver comentário mais
   * abaixo). O celular (.compact) ignora.
   */
  export let forceCols = 0;

  // Paleta da logo, na ordem das opções — a mesma cor identifica a zona no
  // telão e no celular.
  const ZONE_COLORS = ['var(--cyan)', 'var(--pink)', 'var(--yellow)', 'var(--orange)'];

  $: totalParticipants =
    pending.length + groups.reduce((sum, g) => sum + g.participants.length, 0);

  function firstName(p) {
    return (p.name || p.email).split(' ')[0];
  }

  function zoneColor(i) {
    return ZONE_COLORS[i % ZONE_COLORS.length];
  }

  /*
   * O telão (layout="screen") é projetado — ninguém rola uma tela projetada
   * durante a dinâmica. A fila de pendentes tem um teto fixo de rostos
   * visíveis e o resto vira um chip "+N"; o teto garante caber em 1366×768
   * (o piso: notebook/projetor mais comum). Com muitas zonas (7+), a fila
   * inteira encolhe (.tight) pra devolver ~35px de altura ao placar.
   */
  const PENDING_CAP = 14;

  // O chip "+N" ocupa um slot na mesma fileira — quando vai sobrar gente, o
  // teto de itens visíveis cede um lugar pra ele.
  function capWithChipRoom(count, cap) {
    if (count <= cap) return { visible: count, hidden: 0 };
    return { visible: cap - 1, hidden: count - (cap - 1) };
  }

  $: pendingSlots = layout === 'screen' ? capWithChipRoom(pending.length, PENDING_CAP) : null;
  $: visiblePending = pendingSlots ? pending.slice(0, pendingSlots.visible) : pending;
  $: hiddenPendingCount = pendingSlots ? pendingSlots.hidden : 0;
  // Depende só do nº de zonas (input estático), nunca de medida de tela —
  // senão o encolher da fila mudaria a altura medida do placar, que mudaria
  // a densidade, que mudaria a fila… (loop de layout).
  $: pendingTight = layout === 'screen' && groups.length >= 7;

  /*
   * ---------- Placar com densidade automática ----------
   *
   * Cada resposta vira uma zona (rótulo + contagem, e os respondentes como
   * pílulas de rosto + primeiro nome). Por pedido explícito, as pílulas
   * ficam sempre na MENOR fonte/avatar definidos (sizesFor(MIN_SCALE)) — não
   * existe mais busca por uma escala maior que caiba; o único grau de
   * liberdade automático é quantas colunas (1–4) evitam cortar gente. Como o
   * telão não rola, o componente mede o espaço real
   * (bind:clientWidth/Height em .zones) e ESTIMA (fator 0.6 × fonte pra
   * largura de texto) se aquele nº de colunas, no tamanho mínimo, cabe sem
   * cortar ninguém. Só se nem assim coubesse (evento gigante) é que volta o
   * chip "+N", com teto por zona. O overflow:hidden + máscara em
   * .zone-people é a rede de segurança pra qualquer erro de arredondamento.
   */
  const MIN_SCALE = 0.3;
  const MAX_COLS = 4;
  const COL_GAP = 16;
  // Título da resposta NÃO escala com a densidade: fixo e legível — quem
  // cresce/encolhe conforme o espaço são só as pílulas. Mesmo valor do
  // font-size de .zone-text/.zone-count no CSS; mudou um, mude o outro.
  const LABEL_FONT = 20;

  // Tamanhos derivados da escala — os valores em s=1 são os mesmos dos
  // fallbacks das custom properties no CSS abaixo; mudou um, mude o outro.
  // Os pisos (Math.max) é que definem o "menor possível" de verdade, já que
  // hoje só se usa sizesFor(MIN_SCALE) — mudou um piso, ajuste o outro lado.
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
      const w = pillWidth(firstName(p).length, z, withName);
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

  function fitDensity(gs, W, H, withName, forced) {
    if (!W || !H || gs.length === 0) return null; // sem medida (1º frame/testes): CSS usa os fallbacks
    const labelW = Math.min(Math.max(200, W * 0.28), 460);
    const candidates = forced
      ? [Math.min(forced, Math.max(1, gs.length))]
      : [1, 2, 3, 4].filter((c) => c <= MAX_COLS && c <= Math.max(1, gs.length));

    // Fonte fixa no mínimo (ver comentário acima) — só varia quantas
    // colunas cabem sem cortar ninguém.
    const zFixed = sizesFor(MIN_SCALE);
    for (const cols of candidates) {
      if (estimateTotal(gs, W, withName, cols, zFixed, labelW) <= H) {
        return { s: MIN_SCALE, cols, labelW, caps: null };
      }
    }

    // Nem na escala mínima coube: pra cada nº de colunas candidato, raciona
    // as linhas de pílulas por zona e corta com "+N" — vence o arranjo que
    // mostra mais gente no total.
    let best = null;
    const z = zFixed;
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
          ? g.participants.reduce((sum, p) => sum + pillWidth(firstName(p).length, z, withName), 0) /
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

  $: metrics = layout === 'screen' ? fitDensity(groups, zonesW, zonesH, showNames, forceCols) : null;
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
</script>

{#if layout === 'screen' && pending.length === 0}
  <!-- Sem ninguém na fila a faixa vira uma linha de texto: no telão, aqueles
       76px de card vazio empurram a última zona para fora da tela. -->
  <p class="text-muted present-empty pending-note">
    {totalParticipants === 0 ? 'Ninguém respondeu ainda.' : 'Todas as respostas foram reveladas.'}
  </p>
{:else if layout === 'screen'}
  <div class="pending-row" class:tight={pendingTight}>
    {#each visiblePending as p (p.id)}
      <div class="face-wrap" animate:flip={{ duration: 350 }} out:fade={{ duration: 150 }}>
        <button
          type="button"
          class="face pending"
          class:static={!onFaceClick}
          disabled={!onFaceClick}
          title={p.name || p.email}
          aria-label={onFaceClick ? `Revelar resposta de ${p.name || p.email}` : p.name || p.email}
          on:click={() => onFaceClick && onFaceClick(p)}
        >
          {#if p.photo}
            <img src={p.photo} alt="" />
          {:else}
            <span class="face-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
          {/if}
        </button>
        {#if showNames}
          <span class="face-name">{firstName(p)}</span>
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
  style={zoneStyle}
  bind:clientWidth={zonesW}
  bind:clientHeight={zonesH}
>
  {#if hideZones}
    <p class="text-muted present-empty">O organizador escondeu as respostas por enquanto.</p>
  {:else}
    {#if groups.length === 0}
      <p class="text-muted present-empty">Ninguém respondeu ainda.</p>
    {/if}

    {#each visibleGroups as group, gi (group.label)}
      <div class="zone" style="--zone-color:{zoneColor(gi)}">
        <div class="zone-label">
          <span class="zone-text">{group.label}</span>
          <span class="zone-count">{group.participants.length}</span>
        </div>
        <div class="zone-people">
          {#each group.visible as p (p.id)}
            <div class="person-wrap" animate:flip={{ duration: 350 }} in:fly={{ y: -20, duration: 350 }}>
              <button
                type="button"
                class="person"
                class:face-only={!showNames}
                class:static={!onFaceClick}
                disabled={!onFaceClick}
                title={p.name || p.email}
                aria-label={onFaceClick ? `Desrevelar resposta de ${p.name || p.email}` : p.name || p.email}
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
                  <span class="person-name">{firstName(p)}</span>
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
    /* nowrap de propósito: é fila de projeção, não pode crescer em altura —
       o teto de PENDING_CAP mantém isso sempre cabendo numa linha só. */
    flex-wrap: nowrap;
    overflow: hidden;
    gap: 12px;
    min-height: 76px;
    padding: 14px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
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
    overflow: hidden;
    /* Se a estimativa errar por uma linha, desvanece a borda cortada em vez
       de um corte reto no meio de uma pílula. */
    mask-image: linear-gradient(to bottom, black 80%, transparent 100%);
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
    mask-image: none;
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
</style>
