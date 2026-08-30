<script>
  export let id = '';

  import { api } from '../lib/api.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Input from '../components/Input.svelte';
  import PinChip from '../components/PinChip.svelte';
  import PublicShell from '../components/PublicShell.svelte';
  import ReactionBar from '../components/ReactionBar.svelte';
  import ReactionBurstLayer from '../components/ReactionBurstLayer.svelte';
  import ThemeToggle from '../components/ThemeToggle.svelte';
  import { fireReaction } from '../lib/reactionStore.js';

  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]+$/;

  // A URL é a única fonte da verdade pra sessão: nada fica salvo no
  // navegador. Sem PIN na URL, não tem como ter entrado — volta pra Home.
  // Um F5 recarrega essa mesma checagem do zero, então só continua na tela
  // de identificação quem chegou com o PIN na barra de endereço (seja pela
  // Home, seja por um link compartilhado que já vem com ?pin=).
  const params = new URLSearchParams(window.location.search);
  const initialPin = (params.get('pin') || '').trim();
  if (!initialPin) {
    navigate('/');
  }

  /*
   * Tela da plateia: o apresentador comanda o conteúdo no /palco (ou na
   * janela de apresentação); aqui a plateia acompanha e interage — reações
   * animando no fundo e, no rodapé, reações + pergunta ao apresentador numa
   * linha só. Não mostra a fila de pendentes nem as perguntas/respostas.
   */
  let step = 'identify'; // identify | watching
  const pinCode = initialPin;
  let email = '';
  let identifyError = '';
  let joining = false;

  let token = '';
  let snapshot = null;
  let loadError = '';
  let connected = true;
  let eventSource = null;

  let qaDraft = '';
  let qaSent = false;
  let qaError = '';

  // Identificador do navegador: quem pergunta pode remover a própria pergunta
  // (o backend casa este id com o do envio). Gerado uma vez e persistido em
  // localStorage pra sobreviver a reloads — é o "identificador pelo navegador"
  // que autoriza a remoção.
  const BROWSER_ID_KEY = 'arandu.browserId';
  let browserId = localStorage.getItem(BROWSER_ID_KEY) || '';
  if (!browserId) {
    browserId =
      typeof crypto !== 'undefined' && crypto.randomUUID
        ? crypto.randomUUID()
        : 'b' + Date.now().toString(36) + Math.random().toString(36).slice(2, 10);
    localStorage.setItem(BROWSER_ID_KEY, browserId);
  }

  // Perguntas enviadas DESTE navegador (id devolvido pelo servidor + texto) —
  // cada uma com botão de remover. Fica em memória pela sessão da página.
  let myQuestions = [];

  async function joinAs(asGuest) {
    identifyError = '';
    const trimmedEmail = asGuest ? '' : email.trim();
    if (!asGuest && !emailRe.test(trimmedEmail)) {
      identifyError = 'Informe um e-mail válido.';
      return;
    }
    joining = true;
    try {
      const result = await api.publico.eventos.aoVivo.entrar(id, {
        pinCode,
        email: trimmedEmail
      });
      token = result.token;
      step = 'watching';
      startWatching();
    } catch (e) {
      identifyError = e.message;
    } finally {
      joining = false;
    }
  }

  async function startWatching() {
    loadError = '';
    try {
      snapshot = await api.publico.eventos.aoVivo.estado(id, token);
    } catch (e) {
      loadError = e.message;
      return;
    }
    connectStream();
  }

  function connectStream() {
    if (eventSource) eventSource.close();
    eventSource = new EventSource(api.publico.eventos.aoVivo.urlFluxo(id, token));
    eventSource.onmessage = (e) => {
      connected = true;
      snapshot = JSON.parse(e.data);
    };
    eventSource.onerror = () => {
      connected = false;
    };
    // A própria reação de quem manda só aparece via este eco do SSE (sem
    // disparo otimista no clique) — evita rajada dupla, já que quem manda
    // também está assinando o próprio stream.
    eventSource.addEventListener('reaction', (e) => {
      fireReaction(JSON.parse(e.data).emoji);
    });
  }

  function handleReact(e) {
    api.publico.eventos.aoVivo.reagir(id, token, e.detail).catch(() => {});
  }

  async function submitQuestion() {
    const text = qaDraft.trim();
    if (!text) return;
    qaError = '';
    try {
      const res = await api.publico.eventos.aoVivo.enviarPergunta(id, token, text, browserId);
      qaDraft = '';
      qaSent = true;
      if (res && res.messageId) {
        myQuestions = [...myQuestions, { id: res.messageId, text }];
      }
      setTimeout(() => {
        qaSent = false;
      }, 3000);
    } catch (e) {
      qaError = e.message;
    }
  }

  // Remove a pergunta própria: pede pro servidor (que casa pelo clientId do
  // navegador) e tira da lista local mesmo se ela já tiver sido dispensada
  // pelo organizador — nesse caso o servidor responde "não encontrada".
  async function removeQuestion(q) {
    try {
      await api.publico.eventos.aoVivo.removerPergunta(id, token, q.id, browserId);
    } catch {
      // já removida no servidor ou sem rede — tira da lista local assim mesmo
    } finally {
      myQuestions = myQuestions.filter((m) => m.id !== q.id);
    }
  }

  function backToJoin() {
    if (eventSource) eventSource.close();
    token = '';
    snapshot = null;
    loadError = '';
    step = 'identify';
  }
</script>

<main class="plateia">
  {#if !initialPin}
    <div class="plateia-center"><p class="text-muted">Redirecionando…</p></div>
  {:else if step === 'identify'}
    <PublicShell>
      <div class="card join-card">
        <div class="card-head">
          <h1>Entrar na apresentação</h1>
          <p class="card-sub">Identifique-se com seu e-mail ou entre como convidado.</p>
        </div>
        <form class="form" novalidate on:submit|preventDefault={() => joinAs(false)}>
          <Input
            label="E-mail"
            type="email"
            bind:value={email}
            error={identifyError}
            placeholder="seu@melhor.email"
            autocomplete="email"
          />
          <Button type="submit" block disabled={joining}>
            {joining ? 'Entrando…' : 'Entrar'}
          </Button>
          <Button
            type="button"
            variant="ghost"
            block
            disabled={joining}
            on:click={() => joinAs(true)}
          >
            Entrar como convidado
          </Button>
        </form>
      </div>
    </PublicShell>
  {:else if loadError}
    <div class="plateia-center">
      <div class="card">
        <h1>Não foi possível entrar</h1>
        <p class="form-error">{loadError}</p>
        <div class="error-actions">
          <Button type="button" on:click={startWatching}>Tentar novamente</Button>
          <Button type="button" variant="secondary" on:click={backToJoin}>Voltar</Button>
        </div>
      </div>
    </div>
  {:else if !snapshot}
    <div class="plateia-center"><p class="text-muted">Carregando…</p></div>
  {:else}
    <!-- Tela da plateia: menu no topo, reações animando no fundo e, no
         rodapé, reações + pergunta ao apresentador numa linha só. -->
    <div class="live">
      <header class="live-head">
        <button
          type="button"
          class="live-logo-btn"
          title="Ir para o início"
          aria-label="Ir para o início"
          on:click={() => navigate('/')}
        >
          <img class="live-logo" src="/img/arandu-logo.png" alt="Arandu" />
        </button>
        <span class="live-event">{snapshot.eventTitle}</span>
        {#if !connected}<span class="text-muted live-reconnect">Reconectando…</span>{/if}
        <span class="live-spacer"></span>
        <PinChip pin={pinCode} variant="boxed" copyable={false} />
        <ThemeToggle />
      </header>

      <div class="live-body">
        {#if snapshot.message}
          <div class="live-overlay"><p>{snapshot.message}</p></div>
        {:else if snapshot.blanked}
          <div class="live-overlay"><p class="text-muted">Aguarde, já voltamos…</p></div>
        {:else}
          <p class="live-hint">Reaja à apresentação ou mande uma pergunta pro apresentador.</p>
        {/if}
      </div>

      {#if snapshot.interactionsEnabled}
        <div class="live-dock">
          <div class="dock-row">
            <ReactionBar compact on:react={handleReact} />
            <form class="qa-form" novalidate on:submit|preventDefault={submitQuestion}>
              <span class="qa-icon" aria-hidden="true">💬</span>
              <input
                class="qa-input"
                bind:value={qaDraft}
                aria-label="Pergunta pro apresentador"
                title="Mande uma pergunta pro apresentador"
                placeholder="Pergunte ao apresentador…"
              />
              <Button type="submit" size="sm" disabled={!qaDraft.trim()}>Enviar</Button>
            </form>
          </div>
          {#if myQuestions.length > 0}
            <ul class="my-questions" aria-label="Minhas perguntas">
              {#each myQuestions as q (q.id)}
                <li class="my-question">
                  <span class="my-question-text" title={q.text}>{q.text}</span>
                  <button
                    type="button"
                    class="my-question-remove"
                    aria-label="Remover minha pergunta"
                    title="Remover minha pergunta"
                    on:click={() => removeQuestion(q)}
                  >✕</button>
                </li>
              {/each}
            </ul>
          {/if}
          {#if qaError}<p class="form-error">{qaError}</p>{/if}
          {#if qaSent}<p class="text-muted qa-sent">Enviado!</p>{/if}
        </div>
      {/if}
    </div>
  {/if}

  <!-- As reações ficam no FUNDO, atrás da tela (z-0; o conteúdo sobe pra
       z-1) — os emojis sobem pela área livre da tela. -->
  <ReactionBurstLayer background />
</main>

<style>
  .plateia {
    flex: 1;
    display: flex;
    flex-direction: column;
    width: 100%;
    min-height: 100vh;
    min-height: 100dvh;
  }

  .plateia-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    /* Acima da camada fixa de reações (z-0): emojis não flutuam sobre o card
       de erro/carregando/identificação. */
    position: relative;
    z-index: 1;
  }

  .join-card {
    max-width: none;
  }

  .card-head {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .card-sub {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.875rem;
  }

  .error-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 10px;
  }

  /* --- Tela da plateia --- */

  /* O conteúdo sobe pra z-1: a camada de reações (ReactionBurstLayer com
     background, position:fixed z-0) fica atrás de tudo. */
  .live {
    position: relative;
    z-index: 1;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .live-head {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: var(--topbar-h);
    padding: 0 16px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
  }

  .live-logo-btn {
    padding: 0;
    border: none;
    background: none;
    display: flex;
    align-items: center;
    cursor: pointer;
  }

  .live-logo {
    height: 22px;
    width: auto;
  }

  .live-event {
    flex: 1;
    min-width: 0;
    font-size: 0.875rem;
    font-weight: 700;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .live-reconnect {
    flex-shrink: 0;
    font-size: 0.8125rem;
  }

  .live-spacer {
    flex: 1;
  }

  .live-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    padding: 20px 24px;
    overflow-y: auto;
  }

  .live-hint {
    margin: 0;
    max-width: 480px;
    text-align: center;
    font-size: 0.9375rem;
    color: var(--text-muted);
  }

  .live-overlay {
    display: flex;
    align-items: center;
    justify-content: center;
    max-width: 720px;
    padding: 32px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
    text-align: center;
  }

  .live-overlay p {
    margin: 0;
    font-size: clamp(1.25rem, 2vw + 0.5rem, 2rem);
    font-weight: 700;
    overflow-wrap: anywhere;
  }

  /* Rodapé: reações e pergunta ao apresentador na MESMA linha. */
  .live-dock {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
    background: var(--bg-elev);
    border-top: 1px solid var(--border);
  }

  .dock-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .qa-form {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 6px 6px 12px;
    border: 1.5px solid var(--border);
    border-radius: var(--radius-row);
    transition: border-color 0.15s ease;
  }

  .qa-form:focus-within {
    border-color: var(--accent);
  }

  /* Indicador de que a caixa é pra perguntar ao apresentador. */
  .qa-icon {
    flex-shrink: 0;
    font-size: 0.9375rem;
    line-height: 1;
  }

  .qa-input {
    flex: 1;
    min-width: 0;
    border: none;
    outline: none;
    background: transparent;
    font-family: var(--font-ui);
    /* 1rem incondicional: 14px em landscape (568-640px) disparava zoom iOS. */
    font-size: 1rem;
    color: var(--text);
  }

  .qa-input::placeholder {
    color: var(--text-subtle);
  }

  .qa-sent {
    margin: 0;
    font-size: 0.8125rem;
  }

  /* Minhas perguntas (deste navegador) — cada uma com botão de remover. */
  .my-questions {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    /* Com 10+ itens o dock (flex-shrink:0) empurrava o form abaixo da dobra. */
    max-height: 40vh;
    overflow-y: auto;
  }

  .my-question {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    background: var(--surface-muted);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
  }

  .my-question-text {
    flex: 1;
    min-width: 0;
    font-size: 0.8125rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .my-question-remove {
    flex-shrink: 0;
    position: relative;
    width: 32px;
    height: 32px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-size: 0.875rem;
    line-height: 1;
    cursor: pointer;
    transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
  }

  /* Área de toque ~44px sem mudar o visual (padrão do CopyButton). */
  .my-question-remove::after {
    content: '';
    position: absolute;
    inset: -6px;
    border-radius: 50%;
  }

  .my-question-remove:hover {
    color: var(--danger);
    border-color: var(--danger);
    background: var(--surface-muted);
  }

  /* Mobile: a linha única do dock colapsa a caixa de pergunta (reações
     fixas ~252px com botões de 44px + botão Enviar ~71px deixam ~0-22px pro
     input). Empilha: reações espalhadas na largura, caixa de pergunta em
     linha própria. */
  @media (max-width: 520px) {
    .dock-row {
      flex-direction: column;
      align-items: stretch;
      gap: 10px;
    }

    .dock-row :global(.reaction-bar.compact) {
      justify-content: space-between;
    }
  }
</style>
