<script>
  import { navigate } from '../lib/router.js';

  const options = [
    {
      id: 1,
      name: 'Atual (web)',
      desc: 'Mesmo glifo/tamanho da web, só centralizado no mobile.',
      glyph: '›',
      color: 'var(--text-subtle)',
      web: { size: '1.125rem', weight: 400 },
      mob: { size: '1.125rem', weight: 400 }
    },
    {
      id: 2,
      name: 'Mobile maior',
      desc: 'Igual à web, mas maior e mais pesado só no mobile.',
      glyph: '›',
      color: 'var(--text-muted)',
      web: { size: '1.125rem', weight: 400 },
      mob: { size: '1.5rem', weight: 700 }
    },
    {
      id: 3,
      name: 'Negrito nos dois',
      desc: 'Peso 700 nos dois breakpoints, tamanho levemente maior.',
      glyph: '›',
      color: 'var(--text-muted)',
      web: { size: '1.25rem', weight: 700 },
      mob: { size: '1.25rem', weight: 700 }
    },
    {
      id: 4,
      name: 'Glifo pesado',
      desc: 'Usa o caractere ❯ (Black Right-Pointing Angle), mais "cheio".',
      chosen: true,
      glyph: '❯',
      color: 'var(--text-muted)',
      web: { size: '1.125rem', weight: 400 },
      mob: { size: '1.125rem', weight: 400 }
    },
    {
      id: 5,
      name: 'Chevron duplo',
      desc: 'Seta dupla » — indica navegação de forma explícita.',
      glyph: '»',
      color: 'var(--text-muted)',
      web: { size: '1rem', weight: 400 },
      mob: { size: '1rem', weight: 400 }
    },
    {
      id: 6,
      name: 'Triângulo fino',
      desc: 'Glifo ▸ (triângulo aberto), visual mais geométrico.',
      glyph: '▸',
      color: 'var(--text-muted)',
      web: { size: '1.125rem', weight: 400 },
      mob: { size: '1.125rem', weight: 400 }
    },
    {
      id: 7,
      name: 'Triângulo sólido',
      desc: 'Glifo ▶ (triângulo cheio) — inegavelmente uma seta.',
      glyph: '▶',
      color: 'var(--text-muted)',
      web: { size: '0.875rem', weight: 400 },
      mob: { size: '0.875rem', weight: 400 }
    },
    {
      id: 8,
      name: 'Seta reta',
      desc: 'Glifo →, leitura clássica de "navegar para".',
      glyph: '→',
      color: 'var(--text-muted)',
      web: { size: '1.125rem', weight: 400 },
      mob: { size: '1.125rem', weight: 400 }
    },
    {
      id: 9,
      name: 'Seta com gancho',
      desc: 'Glifo ➜, seta arredondada com rabo — bem destacada.',
      glyph: '➜',
      color: 'var(--text-muted)',
      web: { size: '1.125rem', weight: 400 },
      mob: { size: '1.125rem', weight: 400 }
    },
    {
      id: 10,
      name: 'Botão circular',
      desc: 'Chevron SVG em círculo — afetivo e visível nos dois.',
      glyph: 'svg',
      color: 'var(--text-muted)',
      web: { size: '1.125rem', weight: 400 },
      mob: { size: '1.125rem', weight: 400 }
    }
  ];

  const railOptions = [
    {
      id: 1,
      name: 'Chevron simples',
      desc: 'chevron_right/chevron_left — o mesmo gesto do atual, agora em fonte de ícones.',
      open: 'chevron_right',
      collapsed: 'chevron_left',
      font: true
    },
    {
      id: 2,
      name: 'Chevron de teclado',
      desc: 'keyboard_arrow_right/left — mais fino, o clássico do Material.',
      open: 'keyboard_arrow_right',
      collapsed: 'keyboard_arrow_left',
      font: true
    },
    {
      id: 3,
      name: 'Chevron duplo',
      desc: 'keyboard_double_arrow_right/left — "dobra até a borda".',
      open: 'keyboard_double_arrow_right',
      collapsed: 'keyboard_double_arrow_left',
      font: true
    },
    {
      id: 4,
      name: 'Navegar',
      desc: 'navigate_next/navigate_before — seta de navegação de paginação.',
      open: 'navigate_next',
      collapsed: 'navigate_before',
      font: true
    },
    {
      id: 5,
      name: 'Menu lateral',
      desc: 'menu_open/menu_close — hambúrguer com setas, metáfora de painel.',
      open: 'menu_open',
      collapsed: 'menu_close',
      font: true
    },
    {
      id: 6,
      name: 'Painel',
      desc: 'right_panel_close/right_panel_open — o próprio trilho com seta.',
      chosen: true,
      open: 'right_panel_close',
      collapsed: 'right_panel_open',
      font: true
    },
    {
      id: 7,
      name: 'Seta preenchida',
      desc: 'arrow_forward/arrow_back — seta sólida, leitura imediata.',
      open: 'arrow_forward',
      collapsed: 'arrow_back',
      font: true
    },
    {
      id: 8,
      name: 'Seta de contorno',
      desc: 'arrow_right_alt/arrow_left_alt — seta aberta, visual leve.',
      open: 'arrow_right_alt',
      collapsed: 'arrow_left_alt',
      font: true
    },
    {
      id: 9,
      name: 'Seta longa',
      desc: 'east/west — seta alongada, direção clara.',
      open: 'east',
      collapsed: 'west',
      font: true
    },
    {
      id: 10,
      name: 'Última página',
      desc: 'last_page/first_page — chevron com barra, "vai até o fim".',
      open: 'last_page',
      collapsed: 'first_page',
      font: true
    }
  ];

  const qaOptions = [
    {
      id: 1,
      name: 'Atual (ghost)',
      desc: 'Sem borda, cor discreta; hover roxo (move/edit) e vermelho (delete). É o que está no ar.',
      variant: 'ghost'
    },
    {
      id: 2,
      name: 'Borda sutil',
      desc: 'Mesmo ghost, mas com borda 1px visível (--border) para os botões se destacarem da linha.',
      variant: 'outline'
    },
    {
      id: 3,
      name: 'Cor por função',
      desc: 'Move/editar em cor de texto normal; deletar em vermelho desde sempre (não só no hover).',
      variant: 'colored'
    },
    {
      id: 4,
      name: 'Círculo',
      desc: 'Botões circulares (raio 50%) com borda sutil — visual mais leve e arredondado.',
      variant: 'circle'
    },
    {
      id: 5,
      name: 'Preenchido suave',
      desc: 'Fundo tint por função: move/editar em --accent-soft, deletar em --tint-danger.',
      chosen: true,
      variant: 'filled-soft'
    },
    {
      id: 6,
      name: 'Preenchido sólido',
      desc: 'Fundo sólido por função: move/editar roxo, deletar vermelho, com ícone branco.',
      variant: 'filled-solid'
    },
    {
      id: 7,
      name: 'Groupado',
      desc: 'Os 4 botões dentro de uma "caixa" única (container com borda e raio), separados por divisores.',
      variant: 'grouped'
    },
    {
      id: 8,
      name: 'Pill group',
      desc: 'Grupo em formato de pílula (raio 999) com fundo --surface-muted e divisores sutis.',
      variant: 'pill-group'
    },
    {
      id: 9,
      name: 'Ícones Material',
      desc: 'Mesmo ghost atual, mas com glifos da fonte de ícones (arrow_upward/arrow_downward/edit/delete).',
      variant: 'material'
    },
    {
      id: 10,
      name: 'Touch 44px',
      desc: 'Alvos de 44px sempre (não só em touch): ghost com borda sutil e ícone maior.',
      variant: 'touch'
    }
  ];

  const editIconOptions = [
    {
      id: 1,
      name: '✎ Lápis',
      desc: 'Lápis inclinado — glifo atual, bem reconhecível.',
      glyph: '✎',
      font: false
    },
    {
      id: 2,
      name: '🖊 Caneta',
      desc: 'Caneta ponta fina — escrita direta.',
      glyph: '🖊',
      font: false
    },
    {
      id: 3,
      name: '✏️ Lápis preenchido',
      desc: 'Lápis sólido, traço mais grosso.',
      glyph: '✏️',
      font: false
    },
    {
      id: 4,
      name: 'edit (Material)',
      desc: 'Glifo oficial "edit" da fonte Material Symbols.',
      glyph: 'edit',
      font: true
    },
    {
      id: 5,
      name: 'edit_note (Material)',
      desc: 'Página com lápis — edição de um documento.',
      glyph: 'edit_note',
      font: true
    },
    {
      id: 6,
      name: 'edit_square (Material)',
      desc: 'Lápis dentro de um quadrado.',
      chosen: true,
      glyph: 'edit_square',
      font: true
    },
    {
      id: 7,
      name: 'pen_size_3 (Material)',
      desc: 'Lápis/pen com traço grosso.',
      glyph: 'pen_size_3',
      font: true
    },
    {
      id: 8,
      name: 'draw (Material)',
      desc: 'Lápis desenhando, mais orgânico.',
      glyph: 'draw',
      font: true
    },
    {
      id: 9,
      name: 'rate_review (Material)',
      desc: 'Rascunho com lápis (estrela de comentário).',
      glyph: 'rate_review',
      font: true
    },
    {
      id: 10,
      name: 'create (Material)',
      desc: 'Glifo clássico "create" (lápis inclinado).',
      glyph: 'create',
      font: true
    }
  ];

  const editColorOptions = [
    { id: 1, name: 'Roxo atual', css: 'var(--accent)', tint: 'var(--accent-soft)', text: 'var(--accent)' },
    { id: 2, name: 'Âmbar', css: '#d97706', tint: 'rgba(217, 119, 6, 0.14)', text: '#b45309' },
    { id: 3, name: 'Laranja', css: 'var(--orange)', tint: 'var(--tint-orange)', text: 'var(--orange-text)' },
    { id: 4, name: 'Amarelo', css: 'var(--yellow)', tint: 'var(--tint-yellow)', text: '#8a6a00' },
    { id: 5, name: 'Ciano', css: 'var(--cyan)', tint: 'var(--tint-cyan)', text: 'var(--cyan-text)', chosen: true },
    { id: 6, name: 'Verde', css: '#16a34a', tint: 'rgba(22, 163, 74, 0.14)', text: '#15803d' },
    { id: 7, name: 'Rosa', css: 'var(--pink)', tint: 'var(--tint-pink)', text: '#c2187d' },
    { id: 8, name: 'Azul', css: '#2563eb', tint: 'rgba(37, 99, 235, 0.14)', text: '#1d4ed8' },
    { id: 9, name: 'Vermelho-âmbar', css: '#ea580c', tint: 'rgba(234, 88, 12, 0.14)', text: '#c2410c' },
    { id: 10, name: 'Âmbar profundo', css: '#b45309', tint: 'rgba(180, 83, 9, 0.16)', text: '#92400e' }
  ];

  const moveIconOptions = [
    {
      id: 1,
      name: '↑↓ Atual',
      desc: 'Setas de texto ↑/↓ — o que está no ar hoje.',
      up: '↑',
      down: '↓',
      font: false
    },
    {
      id: 2,
      name: '⬆⬇',
      desc: 'Setas brancas grossas — bem visíveis, traço pesado.',
      up: '⬆',
      down: '⬇',
      font: false
    },
    {
      id: 3,
      name: '▲▼',
      desc: 'Triângulos sólidos — leitura imediata de "sobe/desce".',
      up: '▲',
      down: '▼',
      font: false
    },
    {
      id: 4,
      name: '⇧⇩',
      desc: 'Setas com barra na base — clássicas de reordenação.',
      up: '⇧',
      down: '⇩',
      font: false
    },
    {
      id: 5,
      name: 'arrow_upward/downward',
      desc: 'Glifos oficiais Material de mover.',
      chosen: true,
      up: 'arrow_upward',
      down: 'arrow_downward',
      font: true
    },
    {
      id: 6,
      name: 'keyboard_arrow_up/down',
      desc: 'Chevrons de teclado do Material — leves.',
      up: 'keyboard_arrow_up',
      down: 'keyboard_arrow_down',
      font: true
    },
    {
      id: 7,
      name: 'keyboard_double_arrow',
      desc: 'Setas duplas — reforça o salto entre posições.',
      up: 'keyboard_double_arrow_up',
      down: 'keyboard_double_arrow_down',
      font: true
    },
    {
      id: 8,
      name: 'north/south',
      desc: 'Setas de direção (bússola) do Material.',
      up: 'north',
      down: 'south',
      font: true
    },
    {
      id: 9,
      name: 'unfold_more',
      desc: 'Símbolo de "reordenar arrastando" — sugere drag.',
      up: 'unfold_less',
      down: 'unfold_more',
      font: true
    },
    {
      id: 10,
      name: 'swap_vert',
      desc: 'Troca vertical — metáfora de mover itens.',
      up: 'swap_vert',
      down: 'swap_vert',
      font: true,
      single: true
    }
  ];

  const deleteIconOptions = [
    {
      id: 1,
      name: '× Atual',
      desc: 'Sinal de vezes — o que está no ar hoje.',
      glyph: '×',
      font: false
    },
    {
      id: 2,
      name: '🗑 Lixeira emoji',
      desc: 'Lixeira colorida (emoji) — carrega cor própria.',
      glyph: '🗑',
      font: false
    },
    {
      id: 3,
      name: '✕',
      desc: 'X fino (U+2715) — leve, menos agressivo que o ×.',
      glyph: '✕',
      font: false
    },
    {
      id: 4,
      name: 'delete (Material)',
      desc: 'Lixeira oficial do Material — icônica.',
      chosen: true,
      glyph: 'delete',
      font: true
    },
    {
      id: 5,
      name: 'delete_outline (Material)',
      desc: 'Lixeira só no contorno — mais leve.',
      glyph: 'delete_outline',
      font: true
    },
    {
      id: 6,
      name: 'close (Material)',
      desc: 'X do Material — fecha/remove.',
      glyph: 'close',
      font: true
    },
    {
      id: 7,
      name: 'remove (Material)',
      desc: 'Sinal de menos — remove sem conotação de fechar.',
      glyph: 'remove',
      font: true
    },
    {
      id: 8,
      name: 'clear (Material)',
      desc: 'X clássico de limpar.',
      glyph: 'clear',
      font: true
    },
    {
      id: 9,
      name: 'delete_forever (Material)',
      desc: 'Lixeira com x — ênfase em destruição.',
      glyph: 'delete_forever',
      font: true
    },
    {
      id: 10,
      name: 'cancel (Material)',
      desc: 'Círculo com x — forte e inequívoco.',
      glyph: 'cancel',
      font: true
    }
  ];

  const dockRightNames = [
    { id: 1, name: 'mode-dock', desc: 'Dock dos modos de apresentação — direto ao ponto.' },
    { id: 2, name: 'present-modes', desc: 'Modos da apresentação, sem "dock".' },
    { id: 3, name: 'density-dock', desc: 'Densidade das colunas é o que esses botões controlam.' },
    { id: 4, name: 'display-dock', desc: 'Dock de exibição/visualização.' },
    { id: 5, name: 'stage-modes', desc: 'Modos do palco (stage).' },
    { id: 6, name: 'layout-switcher', desc: 'Troca o layout das colunas.' },
    { id: 7, name: 'view-modes', desc: 'Modos de visualização.', chosen: true },
    { id: 8, name: 'present-actions', desc: 'Ações de apresentação, alinhado com dock-left.' },
    { id: 9, name: 'mode-tray', desc: 'Bandeja de modos.' },
    { id: 10, name: 'projection-modes', desc: 'Modos de projeção.' }
  ];

  const presentModeSets = [
    {
      id: 1,
      name: 'Atual (texto)',
      desc: 'A · ★ · 1 · 2 · 3 · 4 — o que está no ar hoje.',
      modes: ['A', '★', '1', '2', '3', '4'],
      font: false
    },
    {
      id: 2,
      name: 'auto_awesome + looks',
      desc: 'Material: auto_awesome, star e looks_one..looks_4.',
      modes: ['auto_awesome', 'star', 'looks_one', 'looks_two', 'looks_3', 'looks_4'],
      font: true
    },
    {
      id: 3,
      name: 'dashboard + view_column',
      desc: 'Material: grid/dashboard com view_column nas densidades.',
      modes: ['dashboard_customize', 'auto_mode', 'view_column', 'view_column_2', 'view_module', 'grid_view'],
      font: true
    },
    {
      id: 4,
      name: 'widgets + view_agenda',
      desc: 'Material: widgets, auto_awesome e view_agenda/view_week.',
      modes: ['widgets', 'auto_awesome', 'view_agenda', 'view_day', 'view_week', 'view_module'],
      font: true
    },
    {
      id: 5,
      name: 'Auto/smart em letra',
      desc: 'A e ★ continuam, colunas viram ícones Material.',
      modes: ['A', '★', 'looks_one', 'looks_two', 'looks_3', 'looks_4'],
      font: true
    },
    {
      id: 6,
      name: 'view_module variantes',
      desc: 'Material: view_module para todas as colunas, auto/smart com ícones próprios.',
      modes: ['auto_awesome', 'smart_toy', 'view_module', 'view_week', 'view_quilt', 'grid_view'],
      font: true
    },
    {
      id: 7,
      name: 'Candlestick/grades',
      desc: 'Material: auto_awesome, plus_one e grades numéricas (grid).',
      modes: ['auto_awesome', 'plus_one', 'grid_3x3', 'grid_4x4', 'grid_goldenratio', 'grid_off'],
      font: true
    },
    {
      id: 8,
      name: 'Avançado',
      desc: 'Material: settings_suggest, tune e view_column com larguras.',
      modes: ['settings_suggest', 'tune', 'view_column', 'view_stream', 'view_day', 'view_agenda'],
      font: true
    },
    {
      id: 9,
      name: 'Ícones numéricos',
      desc: 'Material: pin/looks numeric — pin, star e looks_one..4.',
      modes: ['pin', 'star', 'looks_one', 'looks_two', 'looks_3', 'looks_4'],
      font: true
    },
    {
      id: 10,
      name: 'Símbolos',
      desc: 'Glifos: ∞ auto, ✦ smart, ▤▥▦▧ crescente de colunas.',
      modes: ['∞', '✦', '▤', '▥', '▦', '▧'],
      font: false
    }
  ];

  // Auto vs Smart pelo ASPECTO do comportamento, não pelo nome:
  // Automático = "encaixa no menor" (coluna única/empilhado/compacto);
  // Smart = "maior escala" (grade cheia/células iguais/expande).
  const autoSmartPairs = [
    {
      id: 1,
      name: 'Empilhado vs Grade',
      desc: 'view_agenda (coluna empilhada = encaixa no menor) vs grid_view (grade cheia = maior escala).',
      auto: 'view_agenda',
      smart: 'grid_view',
      chosen: true
    },
    {
      id: 2,
      name: 'Coluna vs Painel',
      desc: 'view_column (uma coluna magra) vs dashboard (células que preenchem o espaço).',
      auto: 'view_column',
      smart: 'dashboard'
    },
    {
      id: 3,
      name: 'Lista vs Mosaico',
      desc: 'format_list_bulleted (lista em coluna única) vs view_quilt (mosaico que cresce).',
      auto: 'format_list_bulleted',
      smart: 'view_quilt'
    },
    {
      id: 4,
      name: 'Linha vs Módulos',
      desc: 'view_day (uma faixa) vs view_module (módulos iguais ocupando a área).',
      auto: 'view_day',
      smart: 'view_module'
    },
    {
      id: 5,
      name: 'Comprimir vs Expandir',
      desc: 'compress (comprime em menos colunas) vs expand (expande até a maior escala).',
      auto: 'compress',
      smart: 'expand'
    },
    {
      id: 6,
      name: 'Densidade baixa vs alta',
      desc: 'low_density (espaçado, poucas colunas) vs high_density (grade densa, escala máxima).',
      auto: 'low_density',
      smart: 'high_density'
    },
    {
      id: 7,
      name: 'Pino vs Grade',
      desc: 'pin (foco único) vs grid_4x4 (grade máxima preenchida).',
      auto: 'pin',
      smart: 'grid_4x4'
    },
    {
      id: 8,
      name: 'Notas vs Widgets',
      desc: 'notes (anotações em coluna) vs widgets (blocos que tomam todo o espaço).',
      auto: 'notes',
      smart: 'widgets'
    },
    {
      id: 9,
      name: 'Fluxo vs Apps',
      desc: 'view_stream (fluxo contínuo vertical) vs apps (grade de apps, 2D cheio).',
      auto: 'view_stream',
      smart: 'apps'
    },
    {
      id: 10,
      name: 'Glifos de layout',
      desc: '‖ (uma coluna) vs ▦ (grade cheia) — sem fonte Material.',
      auto: '‖',
      smart: '▦',
      font: false
    }
  ];

  // 2ª leva: pares focados em "MENOR escala vs MAIOR escala" — o aspecto de
  // tamanho/medida em si, não de layout (colunas/grade).
  const autoSmartScalePairs = [
    {
      id: 1,
      name: 'Molduras',
      desc: 'photo_size_select_small (moldura pequena) vs photo_size_select_large (moldura grande) — o mais literal.',
      auto: 'photo_size_select_small',
      smart: 'photo_size_select_large',
      chosen: true
    },
    {
      id: 2,
      name: 'Lupa',
      desc: 'zoom_out (−) vs zoom_in (+) — diminuir vs aumentar o zoom.',
      auto: 'zoom_out',
      smart: 'zoom_in'
    },
    {
      id: 3,
      name: 'Expandir quadro',
      desc: 'close_fullscreen (encolher) vs open_in_full (expandir para o quadro todo).',
      auto: 'close_fullscreen',
      smart: 'open_in_full'
    },
    {
      id: 4,
      name: 'Tela cheia',
      desc: 'fullscreen_exit (sair) vs fullscreen (ocupar a tela inteira).',
      auto: 'fullscreen_exit',
      smart: 'fullscreen'
    },
    {
      id: 5,
      name: 'Fonte',
      desc: 'text_decrease (texto menor) vs text_increase (texto maior) — escala do conteúdo.',
      auto: 'text_decrease',
      smart: 'text_increase'
    },
    {
      id: 6,
      name: 'Largura',
      desc: 'width_normal (coluna estreita) vs width_full (ocupar toda a largura).',
      auto: 'width_normal',
      smart: 'width_full'
    },
    {
      id: 7,
      name: 'Medidor de nível',
      desc: 'signal_cellular_alt_1_bar (nível baixo) vs signal_cellular_alt_4_bar (nível alto) — escala medida.',
      auto: 'signal_cellular_alt_1_bar',
      smart: 'signal_cellular_alt_4_bar'
    },
    {
      id: 8,
      name: 'Desdobrar',
      desc: 'unfold_less (recolhido) vs unfold_more (desdobrado) — quanto ocupa verticalmente.',
      auto: 'unfold_less',
      smart: 'unfold_more'
    },
    {
      id: 9,
      name: 'Espaçamento de linhas',
      desc: 'table_rows_narrow (linhas apertadas) vs table_rows (linhas largas).',
      auto: 'table_rows_narrow',
      smart: 'table_rows'
    },
    {
      id: 10,
      name: 'Densidade',
      desc: 'density_small (grade esparsa) vs density_large (grade densa) — quantidade que preenche.',
      auto: 'density_small',
      smart: 'density_large'
    }
  ];

  // Galeria completa: 40 pares (Automático · Smart) agrupados por tema, pra
  // esgotar as opções de ícone antes de decidir. Automático = "encaixa no
  // menor"/menor escala; Smart = "maior escala"/mais preenchido.
  const autoSmartGallery = [
    // ---- Layout: colunas vs grade ----
    { group: 'Layout', id: 1, name: 'Empilhado vs Grade', desc: 'view_agenda (coluna empilhada) vs grid_view (grade cheia).', auto: 'view_agenda', smart: 'grid_view' },
    { group: 'Layout', id: 2, name: 'Coluna vs Módulos', desc: 'view_column (coluna única) vs view_module (módulos em grade).', auto: 'view_column', smart: 'view_module' },
    { group: 'Layout', id: 3, name: 'Texto vs Painel', desc: 'subject (texto em coluna) vs dashboard (painel que preenche).', auto: 'subject', smart: 'dashboard' },
    { group: 'Layout', id: 4, name: 'Lista vs Mosaico', desc: 'format_list_bulleted (lista) vs view_quilt (mosaico).', auto: 'format_list_bulleted', smart: 'view_quilt' },
    { group: 'Layout', id: 5, name: 'Faixa vs Semana', desc: 'view_day (uma faixa) vs view_week (faixas em colunas).', auto: 'view_day', smart: 'view_week' },
    { group: 'Layout', id: 6, name: 'Fluxo vs Custom', desc: 'view_stream (fluxo vertical) vs dashboard_customize (grade customizável).', auto: 'view_stream', smart: 'dashboard_customize' },
    { group: 'Layout', id: 7, name: 'Linhas vs Apps', desc: 'reorder (linhas enfileiradas) vs apps (grade de apps).', auto: 'reorder', smart: 'apps' },
    { group: 'Layout', id: 8, name: 'Notas vs Widgets', desc: 'notes (anotações em coluna) vs widgets (blocos que tomam o espaço).', auto: 'notes', smart: 'widgets' },
    { group: 'Layout', id: 9, name: 'Lista vs Grade 4×4', desc: 'list (lista) vs grid_4x4 (grade máxima).', auto: 'list', smart: 'grid_4x4' },
    { group: 'Layout', id: 10, name: 'Linhas apertadas vs largas', desc: 'table_rows_narrow vs table_rows — quanto cada linha ocupa.', auto: 'table_rows_narrow', smart: 'table_rows' },

    // ---- Escala: tamanho ----
    { group: 'Escala', id: 11, name: 'Molduras', desc: 'photo_size_select_small vs photo_size_select_large — moldura pequena vs grande.', auto: 'photo_size_select_small', smart: 'photo_size_select_large' },
    { group: 'Escala', id: 12, name: 'Lupa', desc: 'zoom_out (−) vs zoom_in (+) — diminuir vs aumentar.', auto: 'zoom_out', smart: 'zoom_in' },
    { group: 'Escala', id: 13, name: 'Expandir quadro', desc: 'close_fullscreen (encolher) vs open_in_full (expandir no quadro) — escolhido, com os nomes "Compacto"/"Amplo".', auto: 'close_fullscreen', smart: 'open_in_full', chosen: true },
    { group: 'Escala', id: 14, name: 'Tela cheia', desc: 'fullscreen_exit (sair) vs fullscreen (ocupar a tela inteira).', auto: 'fullscreen_exit', smart: 'fullscreen' },
    { group: 'Escala', id: 15, name: 'Fonte', desc: 'text_decrease (texto menor) vs text_increase (texto maior).', auto: 'text_decrease', smart: 'text_increase' },
    { group: 'Escala', id: 16, name: 'Largura', desc: 'width_normal (coluna estreita) vs width_full (toda a largura).', auto: 'width_normal', smart: 'width_full' },
    { group: 'Escala', id: 17, name: 'Comprimir vs Expandir', desc: 'compress vs expand — encolher até caber vs crescer ao máximo.', auto: 'compress', smart: 'expand' },
    { group: 'Escala', id: 18, name: 'Desdobrar', desc: 'unfold_less (recolhido) vs unfold_more (desdobrado).', auto: 'unfold_less', smart: 'unfold_more' },
    { group: 'Escala', id: 19, name: 'Densidade', desc: 'density_small (esparso) vs density_large (denso).', auto: 'density_small', smart: 'density_large' },
    { group: 'Escala', id: 20, name: 'Menos vs Mais', desc: 'remove (−) vs add (+) — o sinal universal de menor/maior.', auto: 'remove', smart: 'add' },

    // ---- Níveis: medidores ----
    { group: 'Níveis', id: 21, name: 'Sinal', desc: 'signal_cellular_alt_1_bar vs signal_cellular_alt_4_bar — nível baixo vs alto.', auto: 'signal_cellular_alt_1_bar', smart: 'signal_cellular_alt_4_bar' },
    { group: 'Níveis', id: 22, name: 'Bateria', desc: 'battery_1_bar vs battery_full — carga baixa vs cheia.', auto: 'battery_1_bar', smart: 'battery_full' },
    { group: 'Níveis', id: 23, name: 'Brilho', desc: 'brightness_low vs brightness_high — brilho baixo vs alto.', auto: 'brightness_low', smart: 'brightness_high' },
    { group: 'Níveis', id: 24, name: 'Volume', desc: 'volume_off vs volume_up — mudo vs volume alto.', auto: 'volume_off', smart: 'volume_up' },
    { group: 'Níveis', id: 25, name: 'Wi-Fi', desc: 'wifi_1_bar vs wifi_4_bar — sinal fraco vs forte.', auto: 'wifi_1_bar', smart: 'wifi_4_bar' },
    { group: 'Níveis', id: 26, name: 'Números 1 vs 4', desc: 'looks_one vs looks_4 — poucas vs muitas colunas, em número.', auto: 'looks_one', smart: 'looks_4' },
    { group: 'Níveis', id: 27, name: 'Filtros 1 vs 4', desc: 'filter_1 vs filter_4 — variante numerada da mesma metáfora.', auto: 'filter_1', smart: 'filter_4' },
    { group: 'Níveis', id: 28, name: 'Pino vs Grade', desc: 'pin (foco único) vs grid_on (grade ligada/cheia).', auto: 'pin', smart: 'grid_on' },
    { group: 'Níveis', id: 29, name: 'Pessoa vs Grupo', desc: 'person (uma pessoa) vs groups (muitos) — quantidade.', auto: 'person', smart: 'groups' },
    { group: 'Níveis', id: 30, name: 'Tendência', desc: 'trending_down (cai) vs trending_up (cresce) — direção da escala.', auto: 'trending_down', smart: 'trending_up' },

    // ---- Direção / foco / glifos ----
    { group: 'Direção & glifos', id: 31, name: 'Para fora vs Para dentro', desc: 'arrow_outward (cresce para fora) vs arrow_inward (contrai).', auto: 'arrow_outward', smart: 'arrow_inward' },
    { group: 'Direção & glifos', id: 32, name: 'Sobe vs Desce', desc: 'call_made (cresce para cima-direita) vs call_received (contrai).', auto: 'call_made', smart: 'call_received' },
    { group: 'Direção & glifos', id: 33, name: 'Bordas vs Tela', desc: 'crop_free (borda solta) vs fullscreen (borda até o fim).', auto: 'crop_free', smart: 'fullscreen' },
    { group: 'Direção & glifos', id: 34, name: 'Foco fraco vs forte', desc: 'center_focus_weak (foco pequeno) vs center_focus_strong (foco preenche).', auto: 'center_focus_weak', smart: 'center_focus_strong' },
    { group: 'Direção & glifos', id: 35, name: 'Recolher vs Expandir', desc: 'expand_less (recolhe) vs expand_more (expande).', auto: 'expand_less', smart: 'expand_more' },
    { group: 'Direção & glifos', id: 36, name: 'Zoom no mapa', desc: 'zoom_out_map (afasta) vs zoom_in_map (aproxima).', auto: 'zoom_out_map', smart: 'zoom_in_map' },
    { group: 'Direção & glifos', id: 37, name: 'Abas', desc: 'tab_unselected (sem seleção) vs tab (ocupado).', auto: 'tab_unselected', smart: 'tab' },
    { group: 'Direção & glifos', id: 38, name: 'Glifo coluna vs grade', desc: '‖ (uma coluna) vs ▦ (grade cheia) — sem fonte Material.', auto: '‖', smart: '▦', font: false },
    { group: 'Direção & glifos', id: 39, name: 'Glifo barras vs quadriculado', desc: '☰ (barras empilhadas) vs ▧ (quadriculado cheio) — sem fonte Material.', auto: '☰', smart: '▧', font: false },
    { group: 'Direção & glifos', id: 40, name: 'Glifo auto vs estrela', desc: '∞ (auto/ilimitado) vs ✦ (smart/destaque) — sem fonte Material.', auto: '∞', smart: '✦', font: false }
  ];

  const navArrowOptions = [
    {
      id: 1,
      name: '← → Atual',
      desc: 'Setas de texto — o que está no ar hoje.',
      left: '←',
      right: '→',
      font: false
    },
    {
      id: 2,
      name: 'arrow_left/right',
      desc: 'Variantes esquerda/direita do MI5 (arrow_upward/downward).',
      left: 'arrow_left',
      right: 'arrow_right',
      font: true
    },
    {
      id: 3,
      name: 'chevron_left/right',
      desc: 'Chevrons finos do Material — leves.',
      left: 'chevron_left',
      right: 'chevron_right',
      font: true
    },
    {
      id: 4,
      name: 'keyboard_arrow',
      desc: 'Chevrons de teclado — os mais usados em navegação.',
      left: 'keyboard_arrow_left',
      right: 'keyboard_arrow_right',
      font: true
    },
    {
      id: 5,
      name: 'keyboard_double',
      desc: 'Setas duplas — pular para início/fim também?',
      left: 'keyboard_double_arrow_left',
      right: 'keyboard_double_arrow_right',
      font: true
    },
    {
      id: 6,
      name: 'west/east',
      desc: 'Setas de direção (bússola) — mesmas do MI8, no eixo horizontal.',
      left: 'west',
      right: 'east',
      font: true
    },
    {
      id: 7,
      name: 'arrow_back/forward',
      desc: 'Setas preenchidas de voltar/avançar.',
      chosen: true,
      left: 'arrow_back',
      right: 'arrow_forward',
      font: true
    },
    {
      id: 8,
      name: 'navigate_before/next',
      desc: 'Setas de paginação do Material.',
      left: 'navigate_before',
      right: 'navigate_next',
      font: true
    },
    {
      id: 9,
      name: 'first_page/last_page',
      desc: 'Setas com barra — ir ao primeiro/último.',
      left: 'first_page',
      right: 'last_page',
      font: true
    },
    {
      id: 10,
      name: 'skip_previous/next',
      desc: 'Setas de mídia — pular pergunta.',
      left: 'skip_previous',
      right: 'skip_next',
      font: true
    }
  ];

  function go(path) {
    return (e) => {
      e.preventDefault();
      navigate(path);
    };
  }

  // Nome encurtado ("Ana Ribeiro" -> "Ana R.") — espelha displayName de
  // PresentationStage.svelte, pra os protótipos da fila de pendentes criarem o
  // mesmo efeito do telão real.
  function shortName(p) {
    const raw = (p.name || p.email || '').trim();
    const parts = raw.split(/\s+/).filter(Boolean);
    if (parts.length < 2) return parts[0] || raw;
    return `${parts[0]} ${parts[parts.length - 1][0].toUpperCase()}.`;
  }

  // 18 pendentes (> 14, o teto do telão): é o caso que hoje vira "+N" no
  // /apresentar — as opções abaixo mostram todos sem o balão.
  const pendingDemo = [
    { id: 'p1', name: 'Ana Ribeiro' },
    { id: 'p2', name: 'Bruno Carvalho' },
    { id: 'p3', name: 'Carla Mendes' },
    { id: 'p4', name: 'Diego Alves' },
    { id: 'p5', name: 'Elisa Farias' },
    { id: 'p6', name: 'Felipe Ramos' },
    { id: 'p7', name: 'Gabriela Souza' },
    { id: 'p8', name: 'Heitor Lima' },
    { id: 'p9', name: 'Isabela Cruz' },
    { id: 'p10', name: 'João Pedro Barros' },
    { id: 'p11', name: 'Karina Nunes' },
    { id: 'p12', name: 'Lucas Ferreira' },
    { id: 'p13', name: 'Mariana Duarte' },
    { id: 'p14', name: 'Nathan Oliveira' },
    { id: 'p15', name: 'Olívia Pereira' },
    { id: 'p16', name: 'Paulo Henrique Sena' },
    { id: 'p17', name: 'Renata Teixeira' },
    { id: 'p18', name: 'Sofia Martins' }
  ];

  // Cada variante: nome, descrição e a classe CSS que aplica o layout. (A
  // classe escolhida pode virar o novo "modo" do /apresentar.)
  const pendingOptions = [
    {
      id: 1,
      name: 'Grade em múltiplas linhas',
      cls: 'pr-a',
      desc: 'A fila quebra em linhas e mostra todo mundo de uma vez — sem rolar nem "+N". Custa mais altura: as zonas encolhem pra ceder espaço.',
    },
    {
      id: 2,
      name: 'Linha com rolagem horizontal',
      cls: 'pr-b',
      desc: 'Uma linha só, todos os rostos alcançáveis rolando na horizontal (como a fila do painel do organizador). O telão fica estático até rolar.',
    },
    {
      id: 3,
      name: 'Letreiro automático (marquee)',
      cls: 'pr-c',
      chosen: true,
      desc: 'Uma linha que desliza sozinha: todo nome passa pela tela ao longo do tempo, sem interação e sem "+N". A leitura é temporal, não instantânea.',
    },
    {
      id: 4,
      name: 'Nuvem de nomes (pílulas)',
      cls: 'pr-d',
      desc: 'Abre mão do avatar — só os nomes em pílulas compactas que quebram em linhas. É o que mostra mais gente no menor espaço; vira uma "lista de quem falta revelar".',
    },
    {
      id: 5,
      name: 'Mini-grade densa',
      cls: 'pr-e',
      desc: 'Avatares e nomes menores que quebram em colunas: mostra todos os rostos com bem menos altura que a grade 1, mantendo reação visual individual.',
    }
  ];
</script>

<main class="page page-wide opcoes">
  <div class="ic-head">
    <h1>Ícones — chevron do Painel</h1>
    <p class="text-muted">
      Página interna de decisão: compara 10 opções de seta para a linha de evento do
      Painel. Cada card mostra a versão <strong>web</strong> (grid) e a versão
      <strong>mobile</strong> (card empilhado) lado a lado.
    </p>
  </div>

  <div class="ic-grid">
    {#each options as opt (opt.id)}
      <div class="ic-card">
        <div class="ic-head">
          <strong>{opt.id} — {opt.name}</strong>
          {#if opt.chosen}
            <span class="ic-chosen">Escolhida</span>
          {/if}
        </div>
        <p class="text-muted ic-desc">{opt.desc}</p>

        <div class="ic-previews">
          <div class="ic-col">
            <span class="ic-label">Web</span>
            <div class="ic-row-web">
              <span class="icw-title">Dinâmica de boas-vindas</span>
              <span class="icw-chip">Coletando</span>
              <span
                class="icw-chev"
                style="color:{opt.color}; font-size:{opt.web.size}; font-weight:{opt.web.weight};"
              >
                {#if opt.glyph === 'svg'}
                  <span class="ic-badge">
                    <svg width="14" height="14" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                      <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                    </svg>
                  </span>
                {:else}
                  {opt.glyph}
                {/if}
              </span>
            </div>
          </div>

          <div class="ic-col">
            <span class="ic-label">Mobile</span>
            <div class="ic-row-mob">
              <span class="icm-title">Dinâmica de boas-vindas</span>
              <span class="icm-pin"><b>PIN</b> 1234</span>
              <span
                class="icm-chev"
                style="color:{opt.color}; font-size:{opt.mob.size}; font-weight:{opt.mob.weight};"
              >
                {#if opt.glyph === 'svg'}
                  <span class="ic-badge">
                    <svg width="14" height="14" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                      <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                    </svg>
                  </span>
                {:else}
                  {opt.glyph}
                {/if}
              </span>
            </div>
          </div>
        </div>

        <code class="ic-values">
          web {opt.web.size}/{opt.web.weight} · mob {opt.mob.size}/{opt.mob.weight}
        </code>
      </div>
    {/each}
  </div>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Botão "Recolher painel lateral" — /evento</h2>
      <p class="text-muted">
        O botão que alterna o trilho de configurações da tela de edição do evento
        (<code>.rail-toggle</code>). É um <code>.icon-btn</code> de 34×34 alinhado no
        canto superior direito do trilho. Cada card mostra a mesma opção nos dois
        estados: <strong>Aberto</strong> (trilho visível, ação "recolher") e
        <strong>Recolhido</strong> (trilho dobrado, ação "expandir").
      </p>
    </div>

    <div class="ic-grid">
      {#each railOptions as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>R{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>

          <div class="rt-previews">
            <div class="rt-col">
              <span class="ic-label">Aberto · recolher</span>
              <div class="rt-rail">
                <button type="button" class="rt-btn" aria-hidden="true" tabindex="-1">
                  {#if opt.font}
                    <span class="msr rt-glyph">{opt.open}</span>
                  {:else}
                    {opt.open}
                  {/if}
                </button>
                <div class="rt-lines">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              </div>
            </div>

            <div class="rt-col">
              <span class="ic-label">Recolhido · expandir</span>
              <div class="rt-rail collapsed">
                <button type="button" class="rt-btn" aria-hidden="true" tabindex="-1">
                  {#if opt.font}
                    <span class="msr rt-glyph">{opt.collapsed}</span>
                  {:else}
                    {opt.collapsed}
                  {/if}
                </button>
              </div>
            </div>
          </div>

          <code class="ic-values">{opt.open} · {opt.collapsed}</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Ações da pergunta — /evento</h2>
      <p class="text-muted">
        Os 4 botões da linha de pergunta (<code>↑</code> mover, <code>↓</code> mover,
        <code>✎</code> editar, <code>×</code> remover) — hoje <code>.icon-btn</code>
        ghost. Cada card mostra a mesma ação em um estilo diferente para comparar
        cor, tipo de botão e tamanho.
      </p>
    </div>

    <div class="ic-grid">
      {#each qaOptions as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>Q{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>

          <div class="qa-preview">
            <div class="qa-mini-row">
              <span class="qa-num">1</span>
              <span class="qa-title">Qual é o seu prato favorito?</span>
            </div>
            <div class="qa-actions" data-variant={opt.variant}>
              <button type="button" class="qa-btn qa-up" aria-hidden="true" tabindex="-1">↑</button>
              <button type="button" class="qa-btn qa-down" aria-hidden="true" tabindex="-1">↓</button>
              <button type="button" class="qa-btn qa-edit" aria-hidden="true" tabindex="-1">✎</button>
              <button type="button" class="qa-btn qa-delete" aria-hidden="true" tabindex="-1">×</button>
            </div>
          </div>

          <code class="ic-values">↑ ↓ ✎ ×</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Ícone do botão editar — /evento</h2>
      <p class="text-muted">
        10 opções de glifo para o botão <code>✎ Editar</code> da linha de pergunta,
        no estilo Q5 (preenchido suave). As opções com Material usam a fonte de
        ícones (<code>.msr</code>).
      </p>
    </div>

    <div class="ic-grid">
      {#each editIconOptions as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>EI{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>

          <div class="qa-preview">
            <div class="qa-mini-row">
              <span class="qa-num">1</span>
              <span class="qa-title">Qual é o seu prato favorito?</span>
            </div>
            <div class="qa-actions">
              <button type="button" class="qa-btn qa-up" aria-hidden="true" tabindex="-1">↑</button>
              <button type="button" class="qa-btn qa-down" aria-hidden="true" tabindex="-1">↓</button>
              <button type="button" class="qa-btn qa-edit qa-icon-demo" aria-hidden="true" tabindex="-1">
                {#if opt.font}
                  <span class="msr qa-msr">{opt.glyph}</span>
                {:else}
                  {opt.glyph}
                {/if}
              </button>
              <button type="button" class="qa-btn qa-delete" aria-hidden="true" tabindex="-1">×</button>
            </div>
          </div>

          <code class="ic-values">edit · ícone ↑ ↓ ✎ ×</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Cores do botão editar — /evento</h2>
      <p class="text-muted">
        Mesmo glifo <code>✎</code>, variando só a cor (fundo tint + texto) do botão
        editar, mantendo o Q5. Candidatas em tons de âmbar/laranja, mais as atuais
        para comparação.
      </p>
    </div>

    <div class="ic-grid">
      {#each editColorOptions as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>EC{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>

          <div class="qa-preview">
            <div class="qa-mini-row">
              <span class="qa-num">1</span>
              <span class="qa-title">Qual é o seu prato favorito?</span>
            </div>
            <div class="qa-actions">
              <button type="button" class="qa-btn qa-up" aria-hidden="true" tabindex="-1">↑</button>
              <button type="button" class="qa-btn qa-down" aria-hidden="true" tabindex="-1">↓</button>
              <button
                type="button"
                class="qa-btn qa-edit"
                aria-hidden="true"
                tabindex="-1"
                style="background:{opt.tint}; color:{opt.text};"
              >✎</button>
              <button type="button" class="qa-btn qa-delete" aria-hidden="true" tabindex="-1">×</button>
            </div>
          </div>

          <code class="ic-values">✎ · {opt.css}</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Setas mover (↑↓) — /evento</h2>
      <p class="text-muted">
        10 opções de seta para os botões <code>↑</code>/<code>↓</code> da linha de
        pergunta, no estilo Q5. As opções Material usam a fonte de ícones
        (<code>.msr</code>).
      </p>
    </div>

    <div class="ic-grid">
      {#each moveIconOptions as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>MI{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>

          <div class="qa-preview">
            <div class="qa-mini-row">
              <span class="qa-num">1</span>
              <span class="qa-title">Qual é o seu prato favorito?</span>
            </div>
            <div class="qa-actions">
              <button type="button" class="qa-btn qa-up" aria-hidden="true" tabindex="-1">
                {#if opt.font}
                  <span class="msr qa-msr">{opt.up}</span>
                {:else}
                  {opt.up}
                {/if}
              </button>
              <button type="button" class="qa-btn qa-down" aria-hidden="true" tabindex="-1">
                {#if opt.font}
                  <span class="msr qa-msr">{opt.down}</span>
                {:else}
                  {opt.down}
                {/if}
              </button>
              <button type="button" class="qa-btn qa-edit" aria-hidden="true" tabindex="-1">
                <span class="msr qa-msr">edit_square</span>
              </button>
              <button type="button" class="qa-btn qa-delete" aria-hidden="true" tabindex="-1">×</button>
            </div>
          </div>

          <code class="ic-values">↑ ↓ · edit · delete</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Botão remover (×) — /evento</h2>
      <p class="text-muted">
        10 opções de glifo para o botão <code>× Remover</code> da linha de pergunta,
        no estilo Q5 (fundo <code>--tint-danger</code>, hover vermelho sólido).
      </p>
    </div>

    <div class="ic-grid">
      {#each deleteIconOptions as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>DI{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>

          <div class="qa-preview">
            <div class="qa-mini-row">
              <span class="qa-num">1</span>
              <span class="qa-title">Qual é o seu prato favorito?</span>
            </div>
            <div class="qa-actions">
              <button type="button" class="qa-btn qa-up" aria-hidden="true" tabindex="-1">↑</button>
              <button type="button" class="qa-btn qa-down" aria-hidden="true" tabindex="-1">↓</button>
              <button type="button" class="qa-btn qa-edit" aria-hidden="true" tabindex="-1">
                <span class="msr qa-msr">edit_square</span>
              </button>
              <button type="button" class="qa-btn qa-delete qa-icon-demo" aria-hidden="true" tabindex="-1">
                {#if opt.font}
                  <span class="msr qa-msr">{opt.glyph}</span>
                {:else}
                  {opt.glyph}
                {/if}
              </button>
            </div>
          </div>

          <code class="ic-values">remove · ↑ ↓ edit</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Nome da classe dos modos de visualização — /palco</h2>
      <p class="text-muted">
        O elemento que agrupa os botões de modo de apresentação (densidade de
        colunas) no rodapé do Palco — antigo <code>.dock-right</code>.
        Escolhido: <code>view-modes</code>, com o rótulo "Modos de visualização
        das respostas".
      </p>
    </div>

    <div class="ic-grid">
      {#each dockRightNames as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>DR{opt.id} — <code>{opt.name}</code></strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>
          <div class="qa-preview">
            <div class="qa-actions dock-right-demo">
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1">A</button>
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1">★</button>
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1">1</button>
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1">2</button>
              <button type="button" class="mode-demo active" aria-hidden="true" tabindex="-1">3</button>
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1">4</button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Ícones dos modos de apresentação — /palco</h2>
      <p class="text-muted">
        10 conjuntos de ícones para os botões de modo do <code>view-modes</code>
        (Automático · Smart · 1 · 2 · 3 · 4 colunas). Os que usam Material vêm da
        fonte <code>.msr</code>.
      </p>
    </div>

    <div class="ic-grid">
      {#each presentModeSets as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>PM{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>
          <div class="qa-preview">
            <div class="qa-actions dock-right-demo">
              {#each opt.modes as glyph (glyph)}
                <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1">
                  {#if opt.font}
                    <span class="msr mode-demo-glyph">{glyph}</span>
                  {:else}
                    {glyph}
                  {/if}
                </button>
              {/each}
            </div>
          </div>
          <code class="ic-values">{opt.modes.join(' · ')}</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Auto vs Smart por aspecto — /palco</h2>
      <p class="text-muted">
        10 pares (Automático · Smart) escolhidos pelo <em>aspecto visual</em> do
        comportamento, não pelo nome: Automático = "encaixa no menor" (coluna
        única / empilhado / compacto); Smart = "maior escala" (grade cheia /
        células iguais / expande).
      </p>
    </div>

    <div class="ic-grid">
      {#each autoSmartPairs as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>AS{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Recomendado</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>
          <div class="qa-preview">
            <div class="qa-actions dock-right-demo">
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1" title="Compacto">
                {#if opt.font === false}
                  {opt.auto}
                {:else}
                  <span class="msr mode-demo-glyph">{opt.auto}</span>
                {/if}
              </button>
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1" title="Amplo">
                {#if opt.font === false}
                  {opt.smart}
                {:else}
                  <span class="msr mode-demo-glyph">{opt.smart}</span>
                {/if}
              </button>
            </div>
          </div>
          <code class="ic-values">{opt.auto} · {opt.smart}</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Auto vs Smart por escala — 2ª leva — /palco</h2>
      <p class="text-muted">
        10 pares (Automático · Smart) focados no aspecto de <em>tamanho/medida</em>:
        Automático = menor escala; Smart = maior escala. Ignoram o layout
        (colunas/grade) e falam direto de quanto cada coisa cresce ou encolhe.
      </p>
    </div>

    <div class="ic-grid">
      {#each autoSmartScalePairs as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>SS{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Recomendado</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>
          <div class="qa-preview">
            <div class="qa-actions dock-right-demo">
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1" title="Compacto">
                <span class="msr mode-demo-glyph">{opt.auto}</span>
              </button>
              <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1" title="Amplo">
                <span class="msr mode-demo-glyph">{opt.smart}</span>
              </button>
            </div>
          </div>
          <code class="ic-values">{opt.auto} · {opt.smart}</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Galeria completa — Auto vs Smart (40 pares) — /palco</h2>
      <p class="text-muted">
        Todas as opções de ícone para os modos <em>Compacto</em> e <em>Amplo</em>
        (antes "Automático"/"Smart"), agrupadas por tema (Layout · Escala · Níveis ·
        Direção/glifos). Compacto = "encaixa no menor" / menor escala; Amplo =
        "maior escala" / mais preenchido. O par escolhido está marcado.
      </p>
    </div>

    {#each ['Layout', 'Escala', 'Níveis', 'Direção & glifos'] as group}
      <h3 class="ic-group">{group}</h3>
      <div class="ic-grid">
        {#each autoSmartGallery.filter((o) => o.group === group) as opt (opt.id)}
          <div class="ic-card">
            <div class="ic-head">
              <strong>#{opt.id} — {opt.name}</strong>
              {#if opt.chosen}
                <span class="ic-chosen">Recomendado</span>
              {/if}
            </div>
            <p class="text-muted ic-desc">{opt.desc}</p>
            <div class="qa-preview">
              <div class="qa-actions dock-right-demo">
                <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1" title="Compacto">
                  {#if opt.font === false}
                    {opt.auto}
                  {:else}
                    <span class="msr mode-demo-glyph">{opt.auto}</span>
                  {/if}
                </button>
                <button type="button" class="mode-demo" aria-hidden="true" tabindex="-1" title="Amplo">
                  {#if opt.font === false}
                    {opt.smart}
                  {:else}
                    <span class="msr mode-demo-glyph">{opt.smart}</span>
                  {/if}
                </button>
              </div>
            </div>
            <code class="ic-values">{opt.auto} · {opt.smart}</code>
          </div>
        {/each}
      </div>
    {/each}
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Setas de navegação ←/→ — /palco</h2>
      <p class="text-muted">
        10 opções de par esquerda/direita para os botões <code>←</code>/<code>→</code>
        do <code>dock-nav</code> (pergunta anterior/próxima). As variantes
        esquerda/direita do MI5 estão no card 2.
      </p>
    </div>

    <div class="ic-grid">
      {#each navArrowOptions as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>NA{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>
          <div class="qa-preview">
            <div class="dock-nav-demo">
              <button type="button" class="nav-demo" aria-hidden="true" tabindex="-1">
                {#if opt.font}
                  <span class="msr nav-demo-glyph">{opt.left}</span>
                {:else}
                  {opt.left}
                {/if}
              </button>
              <span class="nav-demo-counter">1 / 4</span>
              <button type="button" class="nav-demo" aria-hidden="true" tabindex="-1">
                {#if opt.font}
                  <span class="msr nav-demo-glyph">{opt.right}</span>
                {:else}
                  {opt.right}
                {/if}
              </button>
            </div>
          </div>
          <code class="ic-values">{opt.left} · {opt.right}</code>
        </div>
      {/each}
    </div>
  </section>

  <section class="rt-section">
    <div class="ic-head">
      <h2>Fila de pendentes (telão) — mostrar todos os nomes — /apresentar</h2>
      <p class="text-muted">
        Hoje o <code>/apresentar</code> corta a fila em 14 rostos e mostra um balão "+N".
        Esta demo usa 18 pendentes (portanto "sobra") e as cinco direções mostram todos
        sem o balão. Cada uma é um trade-off de <strong>altura</strong> (as zonas encolhem pra
        ceder espaço), <strong>interação</strong> e <strong>leitura</strong>.
      </p>
    </div>

    <div class="ic-grid-stack">
      {#each pendingOptions as opt (opt.id)}
        <div class="ic-card">
          <div class="ic-head">
            <strong>P{opt.id} — {opt.name}</strong>
            {#if opt.chosen}
              <span class="ic-chosen">Escolhida</span>
            {/if}
          </div>
          <p class="text-muted ic-desc">{opt.desc}</p>

          <div class="pr {opt.cls}">
            {#if opt.id === 4}
              {#each pendingDemo as p (p.id)}
                <span class="pr-pill">
                  <span class="pr-pill-initial">{(p.name || p.email || '?')[0]}</span>
                  {p.name || ''}
                </span>
              {/each}
            {:else if opt.id === 3}
              <div class="pr-c-strip">
                {#each pendingDemo as p (p.id)}
                  <div class="pr-face">
                    <span class="pr-avatar">{(p.name || p.email || '?')[0]}</span>
                    <span class="pr-name">{shortName(p)}</span>
                  </div>
                {/each}
                {#each pendingDemo as p (p.id + '-dup')}
                  <div class="pr-face">
                    <span class="pr-avatar">{(p.name || p.email || '?')[0]}</span>
                    <span class="pr-name">{shortName(p)}</span>
                  </div>
                {/each}
              </div>
            {:else}
              {#each pendingDemo as p (p.id)}
                <div class="pr-face">
                  <span class="pr-avatar">{(p.name || p.email || '?')[0]}</span>
                  <span class="pr-name">{shortName(p)}</span>
                </div>
              {/each}
            {/if}
          </div>

          <code class="ic-values">.pending-row.{opt.cls}</code>
        </div>
      {/each}
    </div>
  </section>

  <footer class="ic-footer">
    <button
      type="button"
      class="ic-back"
      on:click={() => (history.length > 1 ? history.back() : navigate('/debug'))}
    >‹ Voltar</button>
    <span class="ic-footer-sep" aria-hidden="true">·</span>
    <a href="/mapa-do-site" on:click={go('/mapa-do-site')}>Ver mapa completo do site ›</a>
  </footer>
</main>

<style>
  .opcoes {
    align-items: stretch;
    gap: 24px;
    padding-bottom: 64px;
  }

  .ic-head h1 {
    margin: 0 0 8px;
    font-size: 2rem;
  }

  .ic-head p {
    max-width: 720px;
    margin: 0;
  }

  .ic-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(320px, 100%), 1fr));
    gap: 16px;
  }

  /* Variante empilhada: um card por linha, ocupando toda a largura — os
     exemplos seguem horizontais (na horizontal), mas cada um abaixo do
     outro, em vez de lado a lado (que esticava as opções pra baixo). */
  .ic-grid-stack {
    display: grid;
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .ic-group {
    margin: 20px 0 10px;
    font-size: 0.875rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-subtle);
  }

  .ic-card {
    display: flex;
    flex-direction: column;
    gap: 10px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    padding: 16px 16px 14px;
    /* Grid item com min-width:auto não encolhe abaixo do conteúdo; sem isso
       uma fila de rostos horizontais (nowrap) força o card pra além do
       viewport e cria rolagem horizontal. */
    min-width: 0;
  }

  .ic-card > .ic-head {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .ic-card > .ic-head strong {
    font-size: 0.9375rem;
  }

  .ic-chosen {
    flex-shrink: 0;
    font-size: 0.625rem;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--success);
    background: var(--tint-success);
    border-radius: 999px;
    padding: 3px 8px;
  }

  .ic-desc {
    margin: 0;
    font-size: 0.8125rem;
  }

  .ic-previews {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .ic-col {
    display: flex;
    flex-direction: column;
    gap: 6px;
    flex: 1 1 200px;
    min-width: 0;
  }

  .ic-label {
    font-size: 0.6875rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--text-muted);
  }

  .ic-row-web {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto 30px;
    align-items: center;
    gap: 8px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    padding: 10px 12px;
  }

  .icw-title {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.8125rem;
    font-weight: 700;
  }

  .icw-chip {
    flex-shrink: 0;
    font-size: 0.625rem;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--accent);
    background: var(--accent-soft);
    border-radius: 999px;
    padding: 3px 8px;
  }

  .icw-chev {
    justify-self: end;
    line-height: 1;
    text-align: right;
  }

  .ic-row-mob {
    position: relative;
    display: flex;
    flex-flow: row wrap;
    align-items: center;
    gap: 6px 10px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    padding: 12px 40px 12px 14px;
  }

  .icm-title {
    flex: 1 1 100%;
    min-width: 0;
    font-size: 0.8125rem;
    font-weight: 700;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    padding-right: 4px;
  }

  .icm-pin {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .icm-pin b {
    font-size: 0.625rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    color: var(--text-muted);
    margin-right: 4px;
  }

  .icm-chev {
    position: absolute;
    top: 50%;
    right: 14px;
    transform: translateY(-50%);
    line-height: 1;
  }

  .ic-badge {
    display: inline-grid;
    place-items: center;
    width: 26px;
    height: 26px;
    border: 1.5px solid currentColor;
    border-radius: 50%;
  }

  .ic-values {
    align-self: flex-start;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .rt-section {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .rt-section .ic-head {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .rt-section h2 {
    margin: 0;
    font-size: 1.5rem;
  }

  .rt-previews {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .rt-col {
    display: flex;
    flex-direction: column;
    gap: 6px;
    flex: 1 1 130px;
    min-width: 0;
  }

  .rt-rail {
    position: relative;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    padding: 8px;
    min-width: 0;
  }

  .rt-rail.collapsed {
    width: 40px;
    padding: 8px;
  }

  .rt-btn {
    width: 34px;
    height: 34px;
    display: grid;
    place-items: center;
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.9375rem;
    line-height: 1;
  }

  .rt-lines {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 10px;
  }

  .rt-lines span {
    height: 6px;
    border-radius: 3px;
    background: var(--surface-muted);
  }

  .rt-lines span:nth-child(1) {
    width: 100%;
  }

  .rt-lines span:nth-child(2) {
    width: 72%;
  }

  .rt-lines span:nth-child(3) {
    width: 46%;
  }

  .rt-glyph {
    font-size: 22px;
    font-variation-settings: 'FILL' 0, 'wght' 500, 'GRAD' 0, 'opsz' 24;
  }

  /* --- Ações da pergunta --- */

  .qa-preview {
    display: flex;
    flex-direction: column;
    gap: 10px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    padding: 12px;
  }

  .qa-mini-row {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .qa-num {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
    border-radius: 6px;
    background: var(--surface-muted);
    color: var(--text-muted);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.6875rem;
    font-weight: 800;
  }

  .qa-title {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.8125rem;
    font-weight: 700;
  }

  .qa-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .qa-btn {
    width: 30px;
    height: 30px;
    display: grid;
    place-items: center;
    border-radius: var(--radius-control);
    border: 1px solid transparent;
    background: transparent;
    color: var(--text-subtle);
    font-family: var(--font-ui);
    font-size: 0.875rem;
    line-height: 1;
    cursor: pointer;
    transition: color 0.15s ease, background-color 0.15s ease, border-color 0.15s ease;
  }

  .qa-btn:hover {
    border-color: var(--accent);
    color: var(--accent);
  }

  .qa-actions .qa-delete:hover {
    border-color: var(--danger);
    color: var(--danger);
  }

  /* 2 — Borda sutil */
  .qa-actions[data-variant='outline'] .qa-btn {
    border-color: var(--border);
    color: var(--text-muted);
  }

  /* 3 — Cor por função */
  .qa-actions[data-variant='colored'] .qa-btn {
    color: var(--text-muted);
  }

  .qa-actions[data-variant='colored'] .qa-delete {
    color: var(--danger);
  }

  .qa-actions[data-variant='colored'] .qa-btn:hover {
    border-color: var(--border-strong);
  }

  .qa-actions[data-variant='colored'] .qa-delete:hover {
    border-color: var(--danger);
    background: var(--tint-danger);
  }

  /* 4 — Círculo */
  .qa-actions[data-variant='circle'] .qa-btn {
    border-radius: 50%;
    border-color: var(--border);
    color: var(--text-muted);
  }

  /* 5 — Preenchido suave */
  .qa-actions[data-variant='filled-soft'] .qa-btn {
    background: var(--accent-soft);
    color: var(--accent);
    border-color: transparent;
  }

  .qa-actions[data-variant='filled-soft'] .qa-delete {
    background: var(--tint-danger);
    color: var(--danger);
  }

  .qa-actions[data-variant='filled-soft'] .qa-btn:hover {
    background: var(--accent);
    color: var(--on-accent);
  }

  .qa-actions[data-variant='filled-soft'] .qa-delete:hover {
    background: var(--danger);
    color: var(--on-accent);
  }

  /* Estilo padrão dos cards de ícone/cor do editar: Q5 em todos os botões. */
  .qa-actions .qa-btn {
    background: var(--accent-soft);
    color: var(--accent);
    border-color: transparent;
  }

  .qa-actions .qa-edit {
    background: var(--tint-cyan);
    color: var(--cyan-text);
  }

  .qa-actions .qa-delete {
    background: var(--tint-danger);
    color: var(--danger);
  }

  .qa-actions .qa-btn:hover {
    background: var(--accent);
    color: var(--on-accent);
    border-color: transparent;
  }

  .qa-actions .qa-delete:hover {
    background: var(--danger);
    color: var(--on-accent);
    border-color: transparent;
  }

  .qa-icon-demo {
    width: auto;
    min-width: 30px;
    padding: 0 8px;
  }

  .qa-msr {
    font-size: 18px;
    font-variation-settings: 'FILL' 0, 'wght' 500, 'GRAD' 0, 'opsz' 24;
  }

  /* 6 — Preenchido sólido */
  .qa-actions[data-variant='filled-solid'] .qa-btn {
    background: var(--accent);
    color: var(--on-accent);
    border-color: transparent;
  }

  .qa-actions[data-variant='filled-solid'] .qa-delete {
    background: var(--danger);
  }

  .qa-actions[data-variant='filled-solid'] .qa-btn:hover {
    background: var(--accent-hover);
  }

  .qa-actions[data-variant='filled-solid'] .qa-delete:hover {
    background: var(--danger-hover);
  }

  /* 7 — Groupado */
  .qa-actions[data-variant='grouped'] {
    gap: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    overflow: hidden;
    align-self: flex-start;
  }

  .qa-actions[data-variant='grouped'] .qa-btn {
    border: none;
    border-radius: 0;
    color: var(--text-muted);
  }

  .qa-actions[data-variant='grouped'] .qa-btn + .qa-btn {
    border-left: 1px solid var(--border);
  }

  /* 8 — Pill group */
  .qa-actions[data-variant='pill-group'] {
    gap: 0;
    background: var(--surface-muted);
    border-radius: 999px;
    padding: 3px;
    align-self: flex-start;
  }

  .qa-actions[data-variant='pill-group'] .qa-btn {
    border: none;
    border-radius: 999px;
    color: var(--text-muted);
  }

  .qa-actions[data-variant='pill-group'] .qa-btn:hover {
    background: var(--bg-elev);
    color: var(--accent);
    border-color: transparent;
  }

  /* 9 — Ícones Material (ghost) */
  .qa-actions[data-variant='material'] .qa-btn {
    color: var(--text-muted);
  }

  /* 10 — Touch 44px */
  .qa-actions[data-variant='touch'] {
    gap: 8px;
  }

  .qa-actions[data-variant='touch'] .qa-btn {
    width: 44px;
    height: 44px;
    border-color: var(--border);
    color: var(--text-muted);
    font-size: 1.125rem;
  }

  .qa-actions[data-variant='touch'] .qa-btn:hover {
    border-color: var(--accent);
    color: var(--accent);
  }

  .qa-actions[data-variant='touch'] .qa-delete:hover {
    border-color: var(--danger);
    color: var(--danger);
  }

  /* --- Palco: modos de apresentação e navegação --- */

  .dock-right-demo {
    gap: 6px;
  }

  .mode-demo {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    display: grid;
    place-items: center;
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 800;
  }

  .mode-demo.active {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
  }

  .mode-demo-glyph {
    font-size: 16px;
    font-variation-settings: 'FILL' 0, 'wght' 500, 'GRAD' 0, 'opsz' 20;
  }

  .dock-nav-demo {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .nav-demo {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-control);
    border: 1px solid var(--border-strong);
    background: var(--bg-elev);
    color: var(--text);
    display: grid;
    place-items: center;
    font-size: 1rem;
    cursor: pointer;
  }

  .nav-demo-glyph {
    font-size: 20px;
    font-variation-settings: 'FILL' 0, 'wght' 500, 'GRAD' 0, 'opsz' 24;
  }

  .nav-demo-counter {
    min-width: 56px;
    text-align: center;
    font-size: 0.875rem;
    font-weight: 800;
    letter-spacing: 0.04em;
  }

  .ic-footer {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    row-gap: 4px;
    gap: 10px;
    font-size: 0.8125rem;
    font-weight: 700;
  }

  .ic-footer a,
  .ic-back {
    display: inline-flex;
    align-items: center;
    min-height: 44px;
    padding: 0 12px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 700;
    cursor: pointer;
  }

  .ic-footer a:hover,
  .ic-back:hover {
    color: var(--accent);
  }

  .ic-back:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .ic-footer-sep {
    color: var(--border-strong);
  }

  /* --- Demo: fila de pendentes (mostrar todos os nomes) --- */

  /* Base do rosto (avatares + nome) — espelha PresentationStage. */
  .pr {
    display: flex;
    gap: 12px;
    padding: 12px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    /* Espalha os rostos pela largura disponível (como o telão real usa
       space-evenly na fila), pra a caixa acompanhar a largura da tela em vez
       de amontoar tudo à esquerda. */
    justify-content: space-evenly;
    /* Nunca deixa a fila estourar a largura do card (e exige que o conteúdo
       nowrap role/corte em vez de vazar). */
    min-width: 0;
    max-width: 100%;
  }

  .pr-face {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 5px;
    width: 76px;
    flex-shrink: 0;
  }

  .pr-avatar {
    width: 56px;
    height: 56px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--accent);
    color: var(--on-accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 800;
    font-size: 1.0625rem;
    overflow: hidden;
  }

  .pr-name {
    max-width: 76px;
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* 1 — grade em múltiplas linhas */
  .pr.pr-a {
    flex-wrap: wrap;
  }

  /* 2 — linha com rolagem horizontal */
  .pr.pr-b {
    flex-wrap: nowrap;
    overflow-x: auto;
    overflow-y: hidden;
    justify-content: flex-start;
  }

  /* 3 — letreiro automático (marquee) */
  .pr.pr-c {
    flex-wrap: nowrap;
    overflow: hidden;
    justify-content: flex-start;
    padding: 12px 0;
  }

  .pr-c-strip {
    display: flex;
    gap: 12px;
    padding-left: 12px;
    animation: pr-marquee 22s linear infinite;
  }

  /* A fila precisa de ambos os "sets" no mesmo bloco animado pra o loop ser
     contínuo: deslocando a tira em -50% a segunda cópia entra no lugar da
     primeira (duplicadas no markup, com chave p.id + '-dup'). */
  @keyframes pr-marquee {
    from {
      transform: translateX(0);
    }
    to {
      transform: translateX(calc(-50% - 6px));
    }
  }

  /* 4 — nuvem de nomes (pílulas, sem avatar) */
  .pr.pr-d {
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .pr-pill {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 5px 12px 5px 6px;
    border-radius: 999px;
    background: var(--surface-muted);
    border: 1px solid var(--border);
    color: var(--text);
    font-size: 0.8125rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .pr-pill-initial {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--accent);
    color: var(--on-accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.6875rem;
    font-weight: 800;
  }

  /* 5 — mini-grade densa (avatares + nomes menores) */
  .pr.pr-e {
    flex-wrap: wrap;
    justify-content: flex-start;
    gap: 8px;
  }
  .pr.pr-e .pr-face {
    width: 56px;
    gap: 3px;
  }

  .pr.pr-e .pr-avatar {
    width: 34px;
    height: 34px;
    font-size: 0.8125rem;
  }

  .pr.pr-e .pr-name {
    max-width: 56px;
    font-size: 0.71875rem;
  }
</style>
