# Auditoria mobile — Arandu

> Data: auditoria realizada por 15 subagentes (1 por tela), cada um com critérios de
> viewport 375px (também 360px/414px), verificação de overflow horizontal, sobreposição
> de componentes, disposição empilhada, alvos de toque ≥ 44px, legibilidade/contraste
> (temas claro e escuro) e cobertura de media queries. As correções sugeridas estão
> citadas com arquivo:linha.
>
> **Status: correções aplicadas.** As recomendações desta auditoria foram implementadas
> (23 arquivos em `frontend/src`, ver resumo ao final): design system primeiro
> (inputs 16px no mobile, alvos de toque, CrumbBar com wrap, tokens de contraste) e em
> seguida as telas 🔴 (Plateia, Painel, Responder, EventoEditar, Palco,
> PalcoApresentar/PresentationStage) e os itens 🟠/🟡 (EventoNovo, Inicio, Debug,
> MapaDoSite, Tela/QuestionForm, ComoFunciona, Privacidade). 188 testes de frontend
> passando e build `vite build` ok.

---

## Resumo das correções aplicadas (por área)

### Design system (componentes compartilhados)
- `app.css`: inputs ≥16px em ≤480px (zoom iOS); padding do `.card` e do
  `.shell-body` reduzidos no mobile; `.switch a` com área de toque; `.icon-btn`
  44px em `@media (pointer: coarse)`; `.page-wide` finalmente definida;
  `.overline` com contraste AA (`--text-muted`).
- `CrumbBar.svelte`: `flex-wrap` em ≤640px (migalhas em linha própria) — resolve o
  overflow com ação primária cortada no Palco e no EventoEditar.
- `ThemeToggle`/`HelpButton`/`PublicShell`: alvos 44px em telas de toque; popover
  da ajuda contido no viewport (`min(280px, calc(100vw - 32px))`); telas baixas com
  logo menor.
- `CopyButton`: área de toque de ~40px via `::after` (visual inalterado).
- `Tabs.svelte`: `flex-wrap` em ≤640px.
- `ReactionBar.svelte`: variante padrão (56px) encolhe para 44px em ≤480px; compact
  sobe de 36px para 44px (dock da Plateia agora empilha).
- `QuestionForm.svelte`: `.type-toggle` empilha em ≤480px (fim do "Múltipla es…").
- `AvatarCropper.svelte`: stage derivado do viewport em <480px (não estoura mais o
  modal em 375px).

### Telas 🔴
- **Plateia**: dock empilhado em ≤520px (caixa de pergunta volta a ter largura),
  `qa-input` 16px, safe-area inferior, `.error-actions` com wrap, botão remover
  pergunta 32px.
- **Painel**: card empilhado em ≤520px com PIN e respostas de volta (eram escondidos
  no mobile), busca reexibida em ≤640px, filtros com 40px de altura, título com 2
  linhas + `title`.
- **Responder**: `justify-content: safe center` no step de pergunta e de privacidade
  (topo nunca inalcançável), segmentos de progresso com 44px tocáveis (traço visual de
  4px), e-mail da revisão com `overflow-wrap`, event-chip com ellipsis e cor
  tema-aware, me-pill/link-btn com alvos maiores, safe-area no dock, `backLabel`.
- **EventoEditar**: `people-col` com `--people-basis` + altura liberada em ≤760px
  (fim da sobreposição da lista pelo painel de detalhes), perguntas com ações
  empilhadas e botões 44px em ≤560px, `insert-btn` visível em `(hover: none)`,
  badges "Individual"/"Resposta aberta" definidos (antes: classes mortas), modal com
  padding menor no mobile.
- **Palco**: shell com `height:auto` + trilho ≤45vh em ≤640px (preview do placar
  volta a ter espaço), preview em `layout="compact"` no mobile, "💡 Dicas" oculto em
  touch (é só-hover), alvos nav/mode/qa-dismiss ≥44px, micro-texto com `--text-muted`,
  `title` nas perguntas do trilho.
- **PalcoApresentar + PresentationStage**: fila de pendentes quebra em linhas com
  rostos menores em ≤640px (antes: ~4 de 14 rostos cortados sem acesso), overlay de
  mensagem com `overflow-y:auto`, `layout="compact"` no mobile + aviso "abra numa
  tela grande", título do evento com ellipsis.

### Telas 🟠/🟡
- **EventoNovo**: `.form-actions` empilha em ≤560px (botões quebravam em 360px),
  prévia do PIN com `--text-muted`.
- **Inicio**: erro do PIN com o ícone do design system, placeholder com `--text-muted`.
- **Debug**: grid sem overflow <336px (`min(240px,100%)`), rótulos `dt` com contraste
  AA, "Janela" reativa a resize/rotação, botão "‹ Voltar", alvos do rodapé ≥44px.
- **MapaDoSite**: chips "Participante"/"Organizador"/"Interno" com tokens tema-aware
  (`--cyan-hover`, `--accent`, `--orange`/`--tint-orange`), links com área de toque.
- **ComoFunciona/Privacidade**: rodapés com `flex-wrap`, links com 44px de altura,
  nota da privacidade em 14px.

### Não aplicado (fora de escopo ou exige UI nova)
- "Revelar por zona" (⚡) e tooltips continuam só-hover no touch — corrigir exige uma
  UI de fallback (clique no rótulo da zona), não apenas CSS.
- Pinch-zoom no AvatarCropper (hoje só slider) — polimento.
- Swatches do /tela não re-resolvem ao trocar o tema — tela interna.


## Visão geral

| Resultado | Telas |
|---|---|
| ✅ Saudáveis (só polimento) | ComoFunciona, Privacidade, Inicio, Entrar, CriarConta, Debug |
| 🟠 Problemas médios, sem quebra | Tela, MapaDoSite, EventoNovo |
| 🔴 Quebram ou perdem função no mobile | **Painel, Plateia, Responder, EventoEditar, Palco, PalcoApresentar** |

Dois padrões dominam: **(A)** barras de ações do CrumbBar que não quebram linha e cortam
a ação primária em 375px, e **(B)** problemas transversais do design system que apareceram
em quase todas as auditorias.

---

## 🔴 Críticos (corrigir primeiro)

### 1. CrumbBar estoura em ≤414px — ação primária cortada + scroll horizontal
Afeta **Palco** (`Palco.svelte:509-541`) e **EventoEditar** (`EventoEditar.svelte:502-512`).
A faixa de 44px (`CrumbBar.svelte:37-50`) carrega numa linha única chip + PIN + 2–3 ações +
breadcrumbs; só os itens rígidos (nowrap) somam mais que os ~335px úteis, e a ação principal
("Modo apresentação" / "Abrir painel ao vivo") fica parcialmente fora da área tocável.

```css
/* CrumbBar.svelte — correção no componente compartilhado */
@media (max-width: 640px) {
  .crumbbar { flex-wrap: wrap; min-height: auto; padding: 6px 12px; row-gap: 4px; }
  .crumbs { width: 100%; }
  .crumb-spacer { display: none; }
}
```

### 2. Painel: grid de eventos estoura em todos os celulares
`Painel.svelte:339-342` — no breakpoint de 900px o grid vira `minmax(120px,1fr) 150px 24px`
com gap 16 + padding: largura mínima intrínseca **362px** contra 296–350px disponíveis em
360–414px → scroll horizontal e chevron/PIN cortados. O mesmo `@media` **esconde busca e PIN**
— justamente o dado que o organizador precisa ao vivo. Correção: breakpoint de telefone
(~520px) com `minmax(0,1fr)` e card empilhado (PIN + status dentro da linha); busca como
linha full-width.

### 3. Plateia: dock de linha única colapsa a caixa de pergunta
`Plateia.svelte:416-431` — 5 reações compactas (204px fixos) + botão "Enviar" (~71px) deixam
o input da pergunta com **~22px em 375px, ~7px em 360px e vazando em 320px**. Correção:

```css
@media (max-width: 520px) {
  .dock-row { flex-direction: column; align-items: stretch; gap: 10px; }
  .dock-row :global(.reaction-bar.compact) { justify-content: space-between; }
}
```

### 4. Responder: centragem insegura pode tornar o topo inalcançável
`Responder.svelte:1034-1040` (`.privacidade-body`) e `765-772` (`.q-main`) —
`justify-content: center` + `overflow-y: auto` com conteúdo maior que a área rolável corta o
início do scroll no Firefox/Safari (topo da política/prévia da pergunta inacessível).
Correção de 1 linha: `justify-content: safe center;` (ou `margin: 0 auto` no filho).

### 5. EventoEditar: lista de participantes coberta pelo painel de detalhes
`EventoEditar.svelte:538` (`flex-basis: 250px` inline) — em coluna (≤760px) o basis vira
altura fixa de 250px, mas o `@media 980px` destrava o `people-list` para `overflow: visible`:
com muitos participantes, o excedente fica pintado **atrás** do `detail-col` (sobreposição
real, conteúdo inacessível sem busca). Correção: variável CSS (`--people-basis`) +
`flex-basis:auto; max-height:45vh` no mobile.

### 6. Palco: orçamento vertical — trilho de 60vh esmaga o painel
`Palco.svelte:733-737` (shell `100dvh`) + `947-960` (`.palco-rail max-height:60vh`) — em
375×667 sobra ~0–77px para o painel: o preview do placar vira uma tira e o organizador rola
entre dock e trilho a cada ação ao vivo. Correção: `@media (max-width:640px)` com
`height:auto` no shell, trilho ≤45vh (ideal: bottom sheet recolhível).

### 7. PalcoApresentar: fila de pendentes cortada sem acesso
`PresentationStage.svelte:776-805` — `nowrap + overflow:hidden` com rostos de 64px: em 375px
cabem ~4 de até 14 rostos, e o chip "+N" fica fora do recorte, sem scroll. O componente
**não tem nenhuma media query**; a primeira deve tratar a fila (wrap ou `overflow-x:auto` +
rostos menores). Complementos: overlay de mensagem sem `overflow-y:auto` (mensagens longas
cortadas) e `layout="screen"` fixo — o modo `compact` **já existe** no PresentationStage e
não é usado por nenhuma view.

---

## 🟠 Transversais do design system (ganho multiplicador)

1. **Zoom do iOS em inputs < 16px** — `.field input { font-size: 0.9375rem }` (`app.css:208`)
   afeta Entrar, CriarConta, EventoNovo, Responder; também `Plateia.svelte:452` (14px) e
   input do aviso em `Palco.svelte:682-696`. Correção única em `app.css`:
   ```css
   @media (max-width: 480px) { .field input, .field textarea, .field select { font-size: 1rem; } }
   ```
2. **Controles de canto 34×34px** (`ThemeToggle.svelte:21-24`, `HelpButton.svelte:71-74`,
   voltar do `PublicShell.svelte:53-70`) — citados por Inicio, Entrar, CriarConta,
   ComoFunciona, Privacidade. Alvo ≥44px em `@media (pointer: coarse)` ou `::after` estendido.
3. **CopyButton 26×26px** (`CopyButton.svelte:58-71`) — usado em todo fluxo de PIN
   (EventoEditar, Painel, Responder, Palco); o pior alvo do app.
4. **Contraste de `--text-subtle`** (~2.3–2.5:1 no claro) em textos pequenos: rótulos `dt`
   do Debug, prévia do PIN em EventoNovo, placeholder do PIN em Inicio, micro-texto do Palco,
   nota em 13px (Privacidade). Trocar por `--text-muted` nesses usos.
5. **Chips de acesso do MapaDoSite**: `--purple` **não é redefinido no tema escuro** (~1.2:1 —
   invisível), `--yellow` como cor de texto (~1.4:1 no claro), `--cyan` (~1.9:1). Correções:
   `var(--accent)`, `--orange`/`--tint-orange`, `--cyan-hover`.
6. **`.page-wide` é classe morta** (usada em Tela/Debug/MapaDoSite, nunca definida) — o
   layout depende de ordem de injeção; definir ou remover.
7. **`.shell-body` padding 32px laterais** (`app.css:477`) aperta tudo em 375px (311px
   úteis); reduzir para ~20px no mobile.
8. **ReactionBar padrão estoura a seção ≤360px** (`ReactionBar.svelte:49-60`, 312px de
   botões) — espelhar a MQ do modo compacto (44px/gap 6) no padrão.
9. **HelpButton popover 280px fixo abre fora da viewport** quando ancorado à esquerda:
   `width: min(280px, calc(100vw - 32px))`.

## 🟡 Recorrentes (baixo esforço, bom retorno)

- **Alvos < 44px**: Tabs ~28px, Segmented ~36px, Switch 44×24, `.icon-btn` 34px,
  nav/mode buttons do Palco (32–36px), segmentos de progresso do Responder (**4px de altura
  clicável**), `.qa-dismiss` 26px.
- **Hover-only sem fallback touch**: "revelar por zona" no Palco (tooltip só mouse),
  divisores "+ Adicionar pergunta aqui" do EventoEditar (`opacity:0` até hover) — usar
  `@media (hover:none)`.
- **Safe-area**: dock da Plateia e do Responder sem `env(safe-area-inset-bottom)`.
- **Zoom/pinch no AvatarCropper**: só slider (stage fixo de 280px ainda estoura o modal do
  EventoEditar em 375px — derivar do viewport).

## ✅ O que já está bom (não mexer)

- Telas públicas simples (Inicio, Entrar, CriarConta, ComoFunciona, Privacidade): coluna
  440px colapsa via `max-width:100%`, sem overflow/sobreposição — a ausência de `@media`
  nelas é aceitável.
- Estrutura geral: sem `100vw`, sem margens negativas, shells sticky com z-index ordenado,
  `100dvh` bem usado, tokens de tema respeitados (os contrastes ruins são casos pontuais de
  token errado, não do sistema).
- Alvos bons: linhas de participante (~52px), opções de resposta (60px), botões `block`.

## Detalhamento por tela

### Inicio.svelte (rota /) — 🟡
- Controles de canto (tema/ajuda) 34×34px < 44px → ampliar em `@media (pointer: coarse)`.
- Erro do PIN sem ícone exigido pelo design system (`.code-note.error`, L53-55/150-153):
  adicionar `::before` com "!" circular, como `.form-error`.
- Placeholder do PIN com `--text-subtle` (~2.4:1 no claro, L136-138) → `--text-muted`.
- Bom: coluna 440px colapsa, logo `max-width:60vw`, form flexível, @media 520px empilha os
  botões do organizador, sem overflow/sobreposição.

### Entrar.svelte (rota /entrar) — 🟡
- Zoom iOS em inputs 15px (app.css:208) → `font-size:1rem` em ≤480px (design system).
- Controles de canto < 44px (PublicShell/ThemeToggle/HelpButton).
- Card com padding 32px aperta em 375px (conteúdo útil ~263px) → reduzir padding em ≤414px.
- Link "Criar conta" com ~20px de altura → `padding: 8px 4px`.
- Folga vertical em telas baixas (logo 168px domina) → encolher logo/paddings com
  `@media (max-height: 700px)`.

### CriarConta.svelte (rota /criar-conta) — 🟡
- Mesmos itens de Entrar (zoom iOS, controles de canto, padding do card).
- Link "Entrar" (`CriarConta.svelte:115-118`) com área de toque pequena.
- Textos auxiliares 12px (`.strength`, erros) no limite; opcional subir para 13–14px.

### ComoFunciona.svelte (rota /como-funciona) — 🟡
- Links do rodapé 13px (~16–20px de altura) → `min-height:44px` / `display:inline-flex`
  (`:151-155`, link inline `:75`).
- Fila de links sem `flex-wrap` aperta em 360px (`:141-149`) → wrap ou empilhar ≤380px.
- Botões 34px da TopBar (design system, não da tela).

### Privacidade.svelte (rota /privacidade) — 🟡
- Rodapé sem `flex-wrap` (`:135-143`) → `flex-wrap: wrap; row-gap: 8px`.
- Links do rodapé ~20px de altura → `padding: 12px 4px`.
- Nota em 13px (`.privacidade-note`, `:127-133`) → 0.875rem; opcional `text-wrap: balance`.

### Debug.svelte (rota /debug) — 🟡
- Grid de cards: `minmax(240px,1fr)` estoura < 336px → `minmax(min(240px,100%),1fr)` (L133).
- Rótulos `dt` 11px com `--text-subtle` (2.45:1 claro / 3.81:1 escuro, L189-195) →
  `--text-muted`.
- Link do rodapé ~18px → `min-height:44px` (L96-98/207-216).
- Valor "Janela" capturado só no mount (L34-38) → reativo a resize/rotação.
- Sem affordance de "voltar" no mobile → link "‹ Voltar".

### EventoNovo.svelte (rota /evento/novo) — 🟠
- `.form-actions` não colapsa (L111-123/224-228): em 375px folga zero; em 360px "Criar
  evento" quebra em 2 linhas → empilhar no `@media (max-width:560px)` existente (L230-234):
  `flex-direction: column; align-items: stretch`.
- Contraste da prévia do PIN (`.pin-choice-value` com `--text-subtle` ≈2.5:1 no claro,
  L210-218) → `--text-muted`.
- Zoom iOS (app.css) e padding 32px do `.shell-body` (app.css:477) — design system.

### Painel.svelte (rota /painel) — 🔴
- **Overflow do grid em todos os celulares** (L339-342): largura mínima 362px vs 296–350px
  disponíveis → breakpoint ~520px com `minmax(0,1fr)` e card empilhado.
- **Busca some no mobile** (L335-337 `.search { display:none }`) → linha full-width no topo.
- **PIN/respostas escondidos no mobile** (L344-347) → reexibir no card empilhado (PinChip
  com copiar); PIN é dado essencial ao vivo.
- Filtros ~29px de altura (L223-234) → `min-height:40px`, gap 8px.
- Título truncado sem `title` (L307-313) → `title={ev.title}` e/ou 2 linhas no mobile.

### Plateia.svelte (rota /plateia/{id}) — 🔴
- **Dock colapsa a caixa de pergunta** (L416-431): em 375px o input fica ~22px → empilhar
  `.dock-row` em `@media (max-width:520px)` (reações `space-between` na largura).
- Reações compactas 40px/36px com gap 6 (ReactionBar.svelte:76-93) → 44px e gap 8.
- `qa-input` 14px (L452) → 1rem (zoom iOS).
- Dock sem `env(safe-area-inset-bottom)` (L406-414).
- `.error-actions` sem wrap (L313-316) → `flex-wrap: wrap`.
- Botão remover pergunta 26×26px (L494-509) → ≥32px (ideal 44px).

### EventoEditar.svelte (rota /evento/{id}) — 🔴
- **CrumbBar estoura** (L502-512): chip status + PIN + "Abrir painel ao vivo" ≈460px+ vs
  335px úteis → correção no CrumbBar (design system) + opcional esconder chip no mobile.
- **Lista de participantes coberta** (L538, `flex-basis:250px` inline; @media 980px
  destrava overflow) → `--people-basis` + `flex-basis:auto; max-height:45vh` em ≤760px.
- **AvatarCropper stage 280px estoura o modal** (AvatarCropper.svelte:5/170-171) → derivar
  do viewport (`STAGE_SIZE = Math.max(200, innerWidth - 130)` <480px).
- Linha das abas estoura com "+ Adicionar pergunta" (Tabs.svelte:45-55) → `flex-wrap` ou
  esconder botão duplicado em ≤480px.
- Linha da pergunta espremida + botões 30px (L1212-1223/1363-1368) → empilhar ações no
  mobile e alvos 44px.
- `insert-btn` `opacity:0` até hover (L1255-1273) → visível em `@media (hover:none)`.
- Configurações enterradas abaixo das perguntas no mobile → âncora/card-resumo.
- Classes mortas `.badge-individual`/`.badge-open-text` (L642-648) — sem definição.

### Palco.svelte (rota /palco/{id}) — 🔴
- **CrumbBar transborda** (L509-541): 5 ações + chip + PIN → wrap (design system) ou mover
  "Dicas"/"Reiniciar tudo" para o trilho.
- **Orçamento vertical** (L733-737 shell 100dvh + L947-960 trilho 60vh): painel ~0–77px em
  375×667 → `height:auto` + trilho ≤45vh (ideal bottom sheet recolhível).
- Preview do placar com layout de projetor em 335px (L572-584) → `layout="compact"` via
  matchMedia ≤640px ou ocultar.
- "Revelar por zona" (⚡) e tooltips só-hover (PresentationStage.svelte:572-573/625-626/
  683-684) → `on:click` no rótulo da zona como fallback touch.
- Alvos < 44px: `.nav-btn` 36, `.mode-btn` 32, `.hint-toggle` 32, `.qa-dismiss` 26.
- Título da pergunta truncado sem `title` (L1040-1048) → `title={q.title}` ou clamp 2.
- Micro-texto com `--text-subtle` (2.3:1 claro) em `.qa-item-author`/`.question-n` →
  `--text-muted`.
- Input do aviso 15px (L682-696) → 1rem no mobile.

### PalcoApresentar.svelte (rota /palco/{id}/apresentar) — 🔴
- **Fila de pendentes cortada** (PresentationStage.svelte:776-805): ~4 de 14 rostos, sem
  scroll e sem "+N" → primeira MQ do componente (`flex-wrap` ou swipe + rostos 44px).
- **Overlay de mensagem cortado sem scroll** (L216-233) → `overflow-y:auto` + fonte menor.
- `layout="screen"` fixo → `compact` via matchMedia ≤640px (o modo compact já existe).
- Título do evento sem ellipsis (L147-163) → `min-width:0` + ellipsis.
- Contraste do `.zone-count` no claro (PresentationStage.svelte:1050-1054, amarelo 1.5:1 /
  ciano 2:1) → escurecer no light.
- Sem aviso de "tela grande" → chip discreto ≤640px (nunca no projetor).

### Tela.svelte (rota /tela, showcase interno) — 🟠
- **ReactionBar demo estoura a seção ≤360px** (ReactionBar.svelte:49-60, 312px vs 271px
  úteis) → MQ ≤480px no padrão (44px/gap 6), como o compact.
- QuestionForm: `.type-toggle` trunca "Múltipla es…" em 375px (QuestionForm.svelte:144-176)
  → empilhar opções ≤480px.
- Popover do HelpButton abre fora da viewport → `width: min(280px, calc(100vw - 32px))`.
- Alvos de toque expostos: CopyButton 26px, `.icon-btn` 34, Tabs ~28, Segmented ~36,
  Switch 44×24.
- AvatarCropper stage 280px fixo (L171) → `min(280px, 100%)`.
- Swatches não re-resolvem ao trocar o tema (L64-74).

### MapaDoSite.svelte (rota /mapa-do-site) — 🟠
- Chip "Organizador": `--purple` não redefinido no tema escuro (~1.2:1, L235-238) →
  `color: var(--accent)`.
- Chip "Interno": `--yellow` como cor de texto (~1.4:1 no claro, L240-243) → `--orange` +
  `--tint-orange`.
- Chip "Participante": `--cyan` ~1.9:1 no claro (L230-233) → `--cyan-hover` (revisar
  também `.sm-access.ok` ~3.2:1).
- Links de rota/rodapé ~16px de altura → área de toque no @media 760px.
- `.page-wide` morta (L118) → definir em app.css ou remover.

### Responder.svelte (rota /responder/{id}) — 🔴
- **Centragem insegura** em `.privacidade-body` (L1034-1040) e `.q-main` (L765-772):
  topo inalcançável no Firefox/Safari → `justify-content: safe center`.
- Segmentos de progresso 4px clicáveis (L734-748) → `padding: 12px 0; margin: -12px 0`.
- E-mail longo na revisão estoura (L503-512/946-953) → `overflow-wrap: anywhere`.
- Chip do título do evento: nowrap sem ellipsis + `--cyan-hover` 2.5:1 no claro (L620-624)
  → `max-width:100%` + ellipsis + `--cyan`/Chip.
- Alvos < 44px: `.me-pill` ~36px, CopyButton 26px, `.link-btn` ~20px (L595-604/668-679).
- Zoom/pinch no AvatarCropper só por slider (polimento).
- Dock sem `env(safe-area-inset-bottom)` (L1190-1192).
- Identificação sem `backLabel` no PublicShell (L291).

## Ordem de execução sugerida

1. **Design system primeiro** (resolve ~60% dos problemas de uma vez): inputs 16px no
   mobile, alvos de toque (CopyButton/corner controls/icon-btn), `flex-wrap` no CrumbBar,
   `--text-muted` no lugar de `--text-subtle`, `.page-wide`.
2. **As 6 telas 🔴**: Plateia (dock empilhado), Painel (grid + PIN/busca), Responder
   (`safe center`), EventoEditar (people-col), Palco (CrumbBar + orçamento vertical),
   PalcoApresentar (fila de pendentes + layout compact).
3. **🟠/🟡 restantes** por tela, usando os relatórios individuais com linha exata e CSS
   pronto.
---

# 2ª rodada — reauditoria e correções

> Reauditada por 15 subagentes (1 por tela) sobre o estado JÁ CORRIGIDO da 1ª
> rodada: verificação das correções, caça a regressões e itens que ainda falham
> no mobile. Correções aplicadas em seguida; 188 testes + `vite build` OK.

## Regressões da 1ª rodada encontradas e corrigidas

1. 🔴 **Fila de pendentes some no compact** — o template do `PresentationStage`
   só renderizava a fila em `layout === 'screen'`; com o compact forçado em ≤640px
   (Palco e PalcoApresentar), a janela de projeção e o painel do organizador
   perdiam a lista de rostos e a revelação individual no celular. Corrigido: a
   fila agora renderiza em `screen` **e** `compact` (`PresentationStage.svelte:552-558`).
2. 🔴 **Placar cortado na base no compact** — página `100dvh + overflow:hidden`
   com zonas compactas maiores que o viewport. Corrigido: `overflow-y: auto` na
   página em ≤640px (`PalcoApresentar.svelte`).
3. 🔴 **Trilho do Palco com 45vh engolido pelo chrome fixo (~296px)** — o painel
   de Q&A/Aviso ficava com ~0-70px. Corrigido: `max-height: none` no trilho e
   `max-height: 45vh` só no painel interno (`Palco.svelte`).
4. 🟠 **Segmentos de progresso do Responder sem os 44px prometidos** —
   `.segments` não esticava (`align-items: center` na faixa): alvo real ~16px.
   Corrigido com `align-self: stretch` (`Responder.svelte`).
5. 🟠 **`row-gap: 4px` morto nos rodapés** (antes do `gap: 24px`) + rodapé no
   limite dos 375px. Corrigido: `gap: 4px 20px` (ComoFunciona e Privacidade).
6. 🟠 **Erro de carregamento do EventoEditar invisível no mobile** (renderizado
   só dentro do rail) — agora um banner `loadError` no topo (separado dos erros
   de formulário, que continuam no rail).

## Problemas restantes encontrados e corrigidos

### Sistema / compartilhados
- 🔴/🟠 **Popover do HelpButton abre fora da tela** (âncora `right:0` + botão à
  esquerda no demo; vazava 24px na TopBar em 320px): `width: min(280px, calc(100vw - 64px))`
  + demo do /tela alinhado à direita.
- 🟠 **Contraste dos chips incompleto na 1ª rodada** — criados tokens de texto
  tema-aware em `app.css` (`--cyan-text`/`--success-text`/`--orange-text`/
  `--purple-text`; valores verificados ≥4.5:1) e aplicados em `eventStatus.js`
  (Painel/EventoEditar/Palco), MapaDoSite, chip "N na sala" do Palco, event-chip
  do Responder e badges do EventoEditar.
- 🟠 **Zoom do iOS em landscape** — MQ dos inputs ≥16px estendida para
  `(max-width: 480px), (max-height: 480px)` (app.css); busca do Painel, 
  `qa-input` da Plateia (agora 1rem incondicional) e `answer-text-input` do
  EventoEditar cobertos.
- 🟠 **Safe-area inerte** — `viewport-fit=cover` no `index.html` (ativa o
  `env(safe-area-inset-bottom)` dos docks) + `interactive-widget=resizes-content`
  (teclado iOS não cobre o input da Plateia).
- 🟡 **`--danger` #e5484d 3.9:1** → `#d13c42` (4.73:1, AA no claro); placeholder
  global `--text-subtle` → `--text-muted`; `.switch a` para ~44px; `.btn-sm` com
  `min-height: 44px` em `pointer:coarse`; alvos de Tabs/Segmented/Switch em
  `pointer:coarse`; `.overline`/cabeçalhos de tabela com `--text-muted`.
- 🟡 **PublicShell**: `justify-content: safe center` (topo acessível em telas
  baixas/landscape — pré-existente) e `padding-top: 60px` (logo não roça nos
  cantos em 320px). `title={area}` no rótulo da TopBar.
- 🟡 **AvatarCropper**: STAGE_SIZE agora reativo a resize/rotação
  (`svelte:window on:resize`).
- 🟡 **`.zone-count` ilegível no tema claro** (amarelo 1.4:1/ciano 1.9:1):
  `filter: brightness(0.55)` só no claro (PresentationStage).

### Telas
- **Palco**: matchMedia cobre landscape `(max-width:640px), (max-height:480px)`;
  trilho sem cap + painel 45vh; 6 botões de modo com wrap em ≤360px; safe-area no
  dock/rodapé; "Reiniciar tudo" oculto ≤480px (CrumbBar vira 2 linhas);
  `qa-dismiss` 44px; chip "N na sala" com `--cyan-text`.
- **PalcoApresentar**: matchMedia com altura; aviso fixo contido (left/right 12px,
  texto quebrado e mais curto); página rola no compact; `title` no nome do evento.
- **Painel**: busca 16px ≤640px; chip de status trunca na coluna de 150px; filtros
  44px em coarse; cabeçalho da tabela e rótulo "PIN" com `--text-muted`.
- **Plateia**: `plateia-center` com z-index acima das reações; lista "Minhas
  perguntas" com teto de 40vh; overlay com `overflow-wrap`; remover pergunta com
  hit area via `::after`; comentário corrigido.
- **Responder**: `link-btn` 44px; `overflow-wrap` em opção/título; `enterkeyhint`
  no textarea; `title` no me-pill.
- **EventoEditar**: `(hover:none)` do insert-btn desaninhado (valia só em retrato);
  44px das ações em `pointer:coarse` (tablets/landscape); `answer-text-input` 16px;
  `.person-order` muted; `title` nos nomes; people-col `min(45vh, 320px)`.
- **EventoNovo**: `title` na prévia do PIN.
- **Inicio**: `autocapitalize`/`spellcheck`/`autocorrect` no input do código;
  `text-wrap: balance` no título; `aria-live` no erro; mensagem pt-BR de rede.
- **CriarConta**: erros de campo somem ao digitar (`on:input` no form — a barra
  de força volta e o erro obsoleto desaparece).
- **Debug**: rodapé com `flex-wrap`; `history.back()` com fallback para
  `/mapa-do-site`; `:focus-visible` no Voltar.
- **MapaDoSite**: links de rota/rodapé com `min-height: 44px`; cabeçalho da
  tabela com `--text-muted`.
- **Tela**: swatches re-resolvem ao trocar o tema (`$: if ($theme)`); demo do
  HelpButton alinhado à direita; `sg-section` com padding 16px ≤480px; grids com
  `min(220px, 100%)`.

## Ainda aberto (documentado, fora de escopo de CSS)
- "Revelar por zona" (⚡) e tooltips seguem só-hover em touch — exige UI de
  fallback (clique no rótulo da zona).
- Mitigação total do teclado iOS cobrindo o dock do Responder no passo de texto
  aberto (depende de atalho/UI, não só CSS).
- Pinch-zoom no AvatarCropper (só slider).
