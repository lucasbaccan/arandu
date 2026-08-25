<script>
  import { onMount } from 'svelte';
  import { showToast } from '../lib/toastStore.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import Select from '../components/Select.svelte';
  import Segmented from '../components/Segmented.svelte';
  import Switch from '../components/Switch.svelte';
  import Card from '../components/Card.svelte';
  import Chip from '../components/Chip.svelte';
  import HelpButton from '../components/HelpButton.svelte';
  import PinChip from '../components/PinChip.svelte';
  import Tabs from '../components/Tabs.svelte';
  import ThemeToggle from '../components/ThemeToggle.svelte';
  import ReactionBar from '../components/ReactionBar.svelte';
  import QuestionForm from '../components/QuestionForm.svelte';
  import AvatarCropper from '../components/AvatarCropper.svelte';
  import { statusInfo, questionKindInfo } from '../lib/eventStatus.js';

  const statusDemo = ['PREPARATION', 'OPEN_FOR_ANSWERS', 'PRESENTING', 'FINISHED'];
  const kindDemo = ['SINGLE_CHOICE', 'OPEN_TEXT', 'GROUP'];
  let tabValue = 'a';

  const colorTokens = [
    { group: 'Neutros', name: '--bg', desc: 'Fundo da página' },
    { group: 'Neutros', name: '--bg-elev', desc: 'Superfície elevada (cards, painéis)' },
    { group: 'Neutros', name: '--surface-muted', desc: 'Superfície neutra dentro de cards (linhas de apoio, chips)' },
    { group: 'Neutros', name: '--border', desc: 'Bordas e divisores sutis (cards)' },
    { group: 'Neutros', name: '--border-strong', desc: 'Borda de maior contraste (campos de texto)' },
    { group: 'Neutros', name: '--text', desc: 'Texto principal' },
    { group: 'Neutros', name: '--text-muted', desc: 'Texto secundário' },
    { group: 'Neutros', name: '--text-subtle', desc: 'Texto terciário (overlines, contadores, placeholders)' },
    { group: 'Marca', name: '--accent', desc: 'Roxo Arandu — cor estrutural, ação primária' },
    { group: 'Marca', name: '--accent-hover', desc: 'Roxo em hover' },
    { group: 'Marca', name: '--accent-soft', desc: 'Fundo de hover do botão fantasma' },
    { group: 'Estados', name: '--danger', desc: 'Erros, exclusão' },
    { group: 'Estados', name: '--danger-hover', desc: 'Vermelho em hover' },
    { group: 'Estados', name: '--success', desc: 'Sucesso, "ao vivo"' },
    { group: 'Estados', name: '--success-hover', desc: 'Verde-azulado em hover' },
    { group: 'Secundárias', name: '--purple', desc: 'Roxo (mesmo valor de --accent)' },
    { group: 'Secundárias', name: '--cyan', desc: 'Ciano' },
    { group: 'Secundárias', name: '--cyan-hover', desc: 'Ciano em hover' },
    { group: 'Secundárias', name: '--yellow', desc: 'Amarelo' },
    { group: 'Secundárias', name: '--yellow-hover', desc: 'Amarelo em hover' },
    { group: 'Secundárias', name: '--pink', desc: 'Rosa' },
    { group: 'Secundárias', name: '--pink-hover', desc: 'Rosa em hover' },
    { group: 'Secundárias', name: '--orange', desc: 'Laranja' },
    { group: 'Secundárias', name: '--orange-hover', desc: 'Laranja em hover' }
  ];

  const tintTokens = [
    { name: '--tint-purple', desc: 'Fundo alternativo com lavagem roxa (marca)' },
    { name: '--tint-cyan', desc: 'Fundo alternativo com lavagem ciano' },
    { name: '--tint-yellow', desc: 'Fundo alternativo com lavagem amarela' },
    { name: '--tint-pink', desc: 'Fundo alternativo com lavagem rosa' },
    { name: '--tint-orange', desc: 'Fundo alternativo com lavagem laranja' },
    { name: '--tint-success', desc: 'Fundo alternativo de sucesso (já usado em badge-presenting)' },
    { name: '--tint-danger', desc: 'Fundo alternativo de erro/alerta' }
  ];

  let resolvedColors = [];
  let resolvedTints = [];

  onMount(() => {
    const styles = getComputedStyle(document.documentElement);
    resolvedColors = colorTokens.map((c) => ({
      ...c,
      hex: styles.getPropertyValue(c.name).trim()
    }));
    resolvedTints = tintTokens.map((c) => ({
      ...c,
      hex: styles.getPropertyValue(c.name).trim()
    }));
  });

  const components = [
    { name: 'Button', file: 'components/Button.svelte', desc: 'Botão com variantes primary/secondary/ghost/danger, largura total (block) e estado desabilitado.' },
    { name: 'TopBar', file: 'components/TopBar.svelte', desc: 'Topbar de 56px do shell do organizador: símbolo, área, tema, ajuda e menu da conta.' },
    { name: 'CrumbBar', file: 'components/CrumbBar.svelte', desc: 'Faixa de 44px com o caminho da tela e os slots de situação (status) e ações.' },
    { name: 'PublicShell', file: 'components/PublicShell.svelte', desc: 'Shell das telas públicas de entrada: marca a 168px, coluna de 440px, cantos com voltar/tema/ajuda.' },
    { name: 'Tabs', file: 'components/Tabs.svelte', desc: 'Abas sublinhadas com contador (Editar evento e trilho do Palco).' },
    { name: 'Chip', file: 'components/Chip.svelte', desc: 'Chip único do sistema (situação, tipo de pergunta), em variantes pill/square.' },
    { name: 'PinChip', file: 'components/PinChip.svelte', desc: 'PIN em monoespaçada com botão copiar, nas variantes inline/boxed/stage.' },
    { name: 'ThemeToggle', file: 'components/ThemeToggle.svelte', desc: 'Alterna claro/escuro (lib/themeStore.js). Fica nos dois shells.' },
    { name: 'HelpButton', file: 'components/HelpButton.svelte', desc: 'Botão "?" com um resumo de como o Arandu funciona.' },
    { name: 'Input', file: 'components/Input.svelte', desc: 'Campo de texto com label, dica, erro e opção de caixa alta (uppercase). Aceita qualquer type nativo (text, date, time, etc.).' },
    { name: 'Select', file: 'components/Select.svelte', desc: 'Dropdown custom (não usa <select> nativo, então o menu de opções é 100% estilizável). Mesma API do Input (label, hint, error, options).' },
    { name: 'Segmented', file: 'components/Segmented.svelte', desc: 'Seletor segmentado (grupo de botões exclusivos) para poucas opções lado a lado.' },
    { name: 'Switch', file: 'components/Switch.svelte', desc: 'Alternador on/off acessível (role="switch").' },
    { name: 'Card', file: 'components/Card.svelte', desc: 'Contêiner de conteúdo, com título/subtítulo opcionais, versão clicável e versão larga (wide).' },
    { name: 'CopyButton', file: 'components/CopyButton.svelte', desc: 'Botão pequeno que copia um texto pro clipboard e mostra um toast de confirmação.' },
    { name: 'Toast', file: 'components/Toast.svelte', desc: 'Notificação flutuante (sucesso/erro/info), global — disparada via lib/toastStore.js.' },
    { name: 'ReactionBar', file: 'components/ReactionBar.svelte', desc: 'Barra de emojis de reação, usada nas telas de Palco e Plateia.' },
    { name: 'ReactionBurstLayer', file: 'components/ReactionBurstLayer.svelte', desc: 'Camada de animação que faz as reações subirem na tela (usa lib/reactionStore.js). Sem controles próprios.' },
    { name: 'QuestionForm', file: 'components/QuestionForm.svelte', desc: 'Formulário completo de criação/edição de pergunta (título, tipo, opções).' },
    { name: 'AvatarCropper', file: 'components/AvatarCropper.svelte', desc: 'Seletor de foto com recorte circular e zoom, usado no fluxo de resposta.' },
    { name: 'ResponsesPanel', file: 'components/ResponsesPanel.svelte', desc: 'Painel de respostas de um evento (busca dados da API — não dá pra demonstrar sem um evento real).' },
    { name: 'PresentationStage', file: 'components/PresentationStage.svelte', desc: 'Grade de rostos pendentes/revelados usada no Palco e na Plateia.' },
    { name: 'Particles', file: 'components/Particles.svelte', desc: 'Fundo animado de partículas coloridas — sem props, só decorativo (já está atrás desta página).' }
  ];

  let switchOn = false;
  let inputBasic = '';
  let inputHint = '';
  let inputError = 'senha123';
  let inputUppercase = '';

  const eventTypeOptions = [
    { value: 'palestra', label: 'Palestra' },
    { value: 'workshop', label: 'Workshop' },
    { value: 'dinamica', label: 'Dinâmica de grupo' }
  ];
  let selectBasic = '';
  let selectError = '';
  let inputDate = '';
  let inputTime = '';

  const periodOptions = [
    { value: 'manha', label: 'Manhã' },
    { value: 'tarde', label: 'Tarde' },
    { value: 'noite', label: 'Noite' }
  ];
  let periodValue = 'tarde';

  function demoQuestionSubmit(e) {
    showToast(`Pergunta "${e.detail.title || '(sem título)'}" enviada (demo, nada foi salvo).`, 'info');
  }
</script>

<main class="page page-wide tela">
  <div class="sg-head">
    <h1>Guia de componentes — Arandu</h1>
    <p class="text-muted">
      Referência viva de tudo que já existe pronto no app: cores, tipografia e os
      componentes reutilizáveis. Use esta página pra saber o que já dá pra montar
      sem escrever CSS novo.
    </p>
  </div>

  <section class="sg-section">
    <h2>Cores</h2>
    <p class="text-muted">
      Definidas em <code>app.css</code> como variáveis CSS. Use sempre <code>var(--nome)</code>,
      nunca hex fixo — é o que garante consistência quando a paleta mudar de novo.
    </p>
    <div class="sg-swatches">
      {#each resolvedColors as c (c.name)}
        <div class="sg-swatch">
          <div class="sg-swatch-color" style="background: {c.hex}"></div>
          <div class="sg-swatch-info">
            <strong>{c.name}</strong>
            <span class="text-muted">{c.hex}</span>
            <span class="text-muted sg-swatch-desc">{c.desc}</span>
          </div>
        </div>
      {/each}
    </div>
  </section>

  <section class="sg-section">
    <h2>Fundos alternativos</h2>
    <p class="text-muted">
      Lavagens de cor translúcidas (<code>--tint-*</code>) pra usar como fundo de card/seção quando
      um branco puro (<code>--bg-elev</code>) ou o cinza neutro (<code>--bg-input</code>) não bastam
      pra destacar algo — mesma técnica já usada em <code>.badge-presenting</code>.
    </p>
    <div class="sg-swatches">
      {#each resolvedTints as t (t.name)}
        <div class="sg-tint-card" style="background: {t.hex}">
          <strong>{t.name}</strong>
          <span class="text-muted">{t.hex}</span>
          <span class="text-muted sg-swatch-desc">{t.desc}</span>
        </div>
      {/each}
    </div>
  </section>

  <section class="sg-section">
    <h2>Tipografia</h2>
    <p class="text-muted">
      Fonte única no app: <code>Nunito Sans</code> (carregada via Google Fonts no
      <code>index.html</code>), com fallback pro conjunto padrão do sistema.
    </p>
    <div class="sg-type-samples">
      <p style="font-size: 2rem; font-weight: 800; margin: 0;">Peso 800 — títulos grandes</p>
      <p style="font-size: 1.4rem; font-weight: 700; margin: 0;">Peso 700 — títulos de card</p>
      <p style="font-size: 1rem; font-weight: 600; margin: 0;">Peso 600 — botões e destaques</p>
      <p style="font-size: 1rem; font-weight: 400; margin: 0;">Peso 400 — texto corrido normal</p>
    </div>
  </section>

  <section class="sg-section">
    <h2>Botões — <code>&lt;Button&gt;</code></h2>
    <p class="text-muted">
      Props: <code>variant</code> (primary/secondary/ghost/danger), <code>size</code> (sm/md/lg),
      <code>block</code>, <code>disabled</code>, <code>type</code>. São só quatro de propósito:
      <strong>uma ação primária por tela</strong>, secundária para o caminho alternativo, fantasma
      para o terciário e danger só para destruição.
    </p>
    <div class="sg-row">
      <Button>Primary</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="danger">Danger</Button>
      <Button disabled>Desabilitado</Button>
    </div>
    <p class="text-muted" style="margin-top: 16px;">Tamanhos:</p>
    <div class="sg-row">
      <Button size="sm">Small</Button>
      <Button size="md">Medium</Button>
      <Button size="lg">Large</Button>
    </div>
    <div class="sg-row" style="max-width: 320px; margin-top: 16px;">
      <Button block>Block (largura total)</Button>
    </div>
    <p class="text-muted" style="margin-top: 16px;">
      Utilitário CSS <code>.icon-btn</code> (não é componente, é uma classe pra botões só-ícone):
    </p>
    <div class="sg-row">
      <button class="icon-btn" title="Editar">✎</button>
      <button class="icon-btn danger" title="Excluir">🗑</button>
      <button class="icon-btn" disabled title="Desabilitado">⋯</button>
    </div>
  </section>

  <section class="sg-section">
    <h2>Campos de texto — <code>&lt;Input&gt;</code></h2>
    <p class="text-muted">
      Props: <code>label</code>, <code>type</code>, <code>value</code> (bind), <code>error</code>,
      <code>hint</code>, <code>placeholder</code>, <code>uppercase</code>, <code>required</code>,
      <code>autocomplete</code>.
    </p>
    <div class="sg-grid">
      <Input label="Campo simples" bind:value={inputBasic} placeholder="Digite algo" />
      <Input label="Com dica" bind:value={inputHint} hint="Isso é um texto de ajuda (hint)." />
      <Input label="Com erro" type="password" bind:value={inputError} error="Senha muito curta." />
      <Input label="Caixa alta" bind:value={inputUppercase} placeholder="Ex: DEV-TEAM" uppercase />
    </div>
  </section>

  <section class="sg-section">
    <h2>Seletores — <code>&lt;Select&gt;</code>, <code>&lt;Input type="date"&gt;</code>, <code>&lt;Segmented&gt;</code></h2>
    <p class="text-muted">
      Dropdown, seleção de data/hora e um seletor segmentado (grupo de opções exclusivas),
      pros casos em que <code>&lt;Input&gt;</code> sozinho não basta.
    </p>
    <div class="sg-grid">
      <Select
        label="Tipo de evento (dropdown)"
        bind:value={selectBasic}
        options={eventTypeOptions}
        placeholder="Selecione um tipo"
      />
      <Select
        label="Com erro"
        bind:value={selectError}
        options={eventTypeOptions}
        placeholder="Selecione um tipo"
        error="Selecione um tipo de evento."
      />
      <Input label="Data" type="date" bind:value={inputDate} />
      <Input label="Hora" type="time" bind:value={inputTime} />
    </div>
    <p class="text-muted" style="margin-top: 16px;">
      <code>&lt;Segmented&gt;</code> — props <code>options</code>, <code>value</code> (bind),
      <code>disabled</code>. Evento <code>on:change</code>.
    </p>
    <div class="sg-row">
      <Segmented options={periodOptions} bind:value={periodValue} />
      <span class="text-muted">Selecionado: {periodValue}</span>
    </div>
  </section>

  <section class="sg-section">
    <h2>Switch — <code>&lt;Switch&gt;</code></h2>
    <p class="text-muted">Props: <code>checked</code>, <code>disabled</code>. Evento: <code>on:change</code>.</p>
    <div class="sg-row">
      <Switch checked={switchOn} on:change={() => (switchOn = !switchOn)} />
      <span>{switchOn ? 'Ligado' : 'Desligado'}</span>
      <Switch checked={true} disabled />
      <span class="text-muted">Desabilitado (sempre ligado)</span>
    </div>
  </section>

  <section class="sg-section">
    <h2>Card — <code>&lt;Card&gt;</code></h2>
    <p class="text-muted">
      Props: <code>title</code>, <code>subtitle</code>, <code>clickable</code>, <code>wide</code>.
      Sem <code>title</code>/<code>subtitle</code>, use o slot livremente (é o que Login/Registro fazem).
    </p>
    <div class="sg-row sg-row-wrap">
      <Card title="Card simples" subtitle="Com título e subtítulo via prop">
        <p class="text-muted">Conteúdo qualquer aqui dentro.</p>
      </Card>
      <Card clickable on:click={() => showToast('Card clicado!', 'success')}>
        <p>Este card é clicável — clique ou aperte Enter/Espaço com foco nele.</p>
      </Card>
    </div>
  </section>

  <section class="sg-section">
    <h2>Chip — <code>&lt;Chip&gt;</code></h2>
    <p class="text-muted">
      Um chip só no sistema inteiro: fundo tonal + texto colorido. Situação do evento, tipo de
      pergunta e rótulo de opção são variantes da mesma peça — <code>shape</code> (pill/square),
      <code>tint</code>, <code>color</code>, <code>dot</code>. As cores saem de
      <code>lib/eventStatus.js</code>, não são escolhidas na tela.
    </p>
    <div class="sg-row">
      {#each statusDemo as s (s)}
        <Chip dot label={statusInfo(s).label} tint={statusInfo(s).tint} color={statusInfo(s).color} />
      {/each}
    </div>
    <div class="sg-row" style="margin-top: 12px;">
      {#each kindDemo as t (t)}
        <Chip
          shape="square"
          label={questionKindInfo(t).label}
          tint={questionKindInfo(t).tint}
          color={questionKindInfo(t).color}
        />
      {/each}
    </div>
  </section>

  <section class="sg-section">
    <h2>PIN — <code>&lt;PinChip&gt;</code></h2>
    <p class="text-muted">
      PIN sempre em monoespaçada com botão copiar. <code>variant</code>: <code>inline</code> (linha da
      tabela), <code>boxed</code> (faixa de breadcrumb) e <code>stage</code> (telão).
    </p>
    <div class="sg-row">
      <PinChip pin="dev-team" />
      <PinChip pin="dev-team" variant="boxed" />
      <PinChip pin="dev-team" variant="stage" copyable={false} />
    </div>
  </section>

  <section class="sg-section">
    <h2>Abas — <code>&lt;Tabs&gt;</code></h2>
    <p class="text-muted">
      Abas sublinhadas com contador. Mesmo componente no Editar evento (Perguntas · Respostas) e no
      trilho do Palco (<code>compact</code>). O contador fica fora do nome acessível.
    </p>
    <Tabs
      bind:value={tabValue}
      tabs={[
        { value: 'a', label: 'Perguntas', count: 8 },
        { value: 'b', label: 'Respostas', count: 17 }
      ]}
    />
    <p class="text-muted" style="margin-top: 12px;">Aba ativa: {tabValue}</p>
  </section>

  <section class="sg-section">
    <h2>Shells de tela <span class="text-muted">(TopBar · CrumbBar · PublicShell)</span></h2>
    <p class="text-muted">
      Toda tela do app usa um dos dois shells. <strong>Organizador</strong>:
      <code>&lt;TopBar&gt;</code> de 56px (só identidade e conta — nunca muda de tela pra tela) +
      <code>&lt;CrumbBar&gt;</code> de 44px, que carrega o caminho, a situação e a ação primária
      daquela tela. <strong>Público</strong>: <code>&lt;PublicShell&gt;</code>, com a marca
      centralizada a 168px, coluna de 440px e os ícones fixos no canto (voltar, tema e ajuda).
    </p>
    <div class="sg-row">
      <ThemeToggle />
      <span class="text-muted">Alternar tema (claro/escuro) — presente nos dois shells</span>
    </div>
    <div class="sg-row" style="margin-top: 12px;">
      <HelpButton />
      <span class="text-muted">Ajuda</span>
    </div>
  </section>

  <section class="sg-section">
    <h2>Toast — <code>&lt;Toast&gt;</code></h2>
    <p class="text-muted">
      Renderizado uma vez em <code>App.svelte</code> (global). Dispare de qualquer
      lugar com <code>showToast(mensagem, tipo)</code> — tipos: success, error, info.
    </p>
    <div class="sg-row">
      <Button on:click={() => showToast('Deu tudo certo!', 'success')}>Disparar sucesso</Button>
      <Button variant="secondary" on:click={() => showToast('Algo deu errado.', 'error')}>Disparar erro</Button>
      <Button variant="secondary" on:click={() => showToast('Só um aviso.', 'info')}>Disparar info</Button>
    </div>
  </section>

  <section class="sg-section">
    <h2>Reações — <code>&lt;ReactionBar&gt;</code></h2>
    <p class="text-muted">Usada nas telas de Palco/Plateia. Evento <code>on:react</code> recebe o emoji clicado.</p>
    <ReactionBar on:react={(e) => showToast(`Reação: ${e.detail}`, 'info', 1200)} />
  </section>

  <section class="sg-section">
    <h2>Formulário de pergunta — <code>&lt;QuestionForm&gt;</code></h2>
    <p class="text-muted">
      Componente completo (título + tipo + opções). Aqui é só demonstração — o
      envio não salva nada de verdade.
    </p>
    <QuestionForm on:submit={demoQuestionSubmit} on:cancel={() => showToast('Cancelado.', 'info')} showCancel />
  </section>

  <section class="sg-section">
    <h2>Seletor de foto — <code>&lt;AvatarCropper&gt;</code></h2>
    <p class="text-muted">Escolha uma imagem qualquer pra ver o recorte circular com zoom.</p>
    <AvatarCropper on:change={(e) => showToast(e.detail ? 'Foto recortada!' : 'Foto removida.', 'success')} />
  </section>

  <section class="sg-section">
    <h2>Lista de componentes</h2>
    <p class="text-muted">Todos os arquivos em <code>frontend/src/components/</code> hoje.</p>
    <div class="sg-table">
      {#each components as c (c.name)}
        <div class="sg-table-row">
          <div class="sg-table-name">{c.name}</div>
          <div class="sg-table-file text-muted">{c.file}</div>
          <div class="sg-table-desc">{c.desc}</div>
        </div>
      {/each}
    </div>
  </section>
</main>

<style>
  .tela {
    align-items: stretch;
    gap: 32px;
    padding-bottom: 64px;
  }

  .sg-head h1 {
    margin: 0 0 8px;
    font-size: 2rem;
  }

  .sg-head p {
    max-width: 640px;
    margin: 0;
  }

  .sg-section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    padding: 24px 28px;
  }

  .sg-section h2 {
    margin: 0 0 6px;
    font-size: 1.2rem;
  }

  .sg-section > p {
    margin: 0 0 16px;
  }

  .sg-section code {
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 1px 5px;
    font-size: 0.85em;
  }

  .sg-row {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .sg-row-wrap {
    align-items: stretch;
  }

  .sg-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
  }

  .sg-swatches {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 12px;
  }

  .sg-swatch {
    display: flex;
    gap: 12px;
    align-items: center;
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 8px;
  }

  .sg-swatch-color {
    width: 44px;
    height: 44px;
    border-radius: 8px;
    border: 1px solid var(--border);
    flex-shrink: 0;
  }

  .sg-swatch-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .sg-swatch-info strong {
    font-family: monospace;
    font-size: 0.85rem;
  }

  .sg-swatch-desc {
    font-size: 0.78rem;
  }

  .sg-tint-card {
    display: flex;
    flex-direction: column;
    gap: 2px;
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 12px;
  }

  .sg-tint-card strong {
    font-family: monospace;
    font-size: 0.85rem;
  }

  .sg-type-samples {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .sg-table {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .sg-table-row {
    display: grid;
    grid-template-columns: 160px 240px 1fr;
    gap: 12px;
    padding: 10px 0;
    border-top: 1px solid var(--border);
  }

  .sg-table-row:first-child {
    border-top: none;
  }

  .sg-table-name {
    font-weight: 700;
    min-width: 0;
  }

  .sg-table-file {
    font-family: monospace;
    font-size: 0.82rem;
    min-width: 0;
    overflow-wrap: break-word;
  }

  .sg-table-desc {
    min-width: 0;
  }

  @media (max-width: 720px) {
    .sg-table-row {
      grid-template-columns: 1fr;
      gap: 4px;
    }
  }
</style>
