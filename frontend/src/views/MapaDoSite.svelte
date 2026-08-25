<script>
  import { navigate } from '../lib/router.js';

  function go(path) {
    return (e) => {
      e.preventDefault();
      navigate(path);
    };
  }

  // Mapa do site: toda rota resolvida em App.svelte, agrupada por fluxo.
  // Rotas com {id} são dinâmicas (precisam de um identificador real); as
  // estáticas são links clicáveis. Manter em sincronia com App.svelte e com a
  // tabela "Screens" do CLAUDE.md.
  const groups = [
    {
      title: 'Públicas — entrada',
      access: 'Público',
      desc: 'Sem login: onde todo mundo começa. O participante digita o PIN na home; quem organiza cria conta ou entra.',
      items: [
        {
          route: '/',
          view: 'Inicio.svelte',
          desc: 'Landing page — digitar o PIN do evento para entrar como participante; links de Entrar/Criar conta para quem organiza.'
        },
        { route: '/entrar', view: 'Entrar.svelte', desc: 'Login de quem organiza (e-mail + senha).' },
        { route: '/criar-conta', view: 'CriarConta.svelte', desc: 'Criação de conta de quem organiza.' },
        {
          route: '/como-funciona',
          view: 'ComoFunciona.svelte',
          desc: 'Tutorial: como funciona o Arandu para quem participa e para quem organiza.'
        },
        { route: '/privacidade', view: 'Privacidade.svelte', desc: 'Política de privacidade.' }
      ]
    },
    {
      title: 'Participantes — evento ao vivo',
      access: 'Participante',
      desc: 'Sem login, com o código/link do evento. A sessão vive na URL (?pin= ou token), não no navegador.',
      items: [
        {
          route: '/responder/{id}',
          view: 'Responder.svelte',
          desc: 'Formulário pré-evento: o participante se identifica, responde às perguntas e pode enviar foto.'
        },
        {
          route: '/plateia/{id}',
          view: 'Plateia.svelte',
          desc: 'Tela pública ao vivo — telão (projeção) ou celular (reações e Q&A).'
        }
      ]
    },
    {
      title: 'Organizador — autenticado',
      access: 'Organizador',
      desc: 'Exigem login. Gerenciamento do evento e da apresentação ao vivo.',
      items: [
        {
          route: '/painel',
          view: 'Painel.svelte',
          desc: 'Home de quem organiza: lista de eventos com situação e PIN.'
        },
        { route: '/evento/novo', view: 'EventoNovo.svelte', desc: 'Criação de evento (título e PIN).' },
        {
          route: '/evento/{id}',
          view: 'EventoEditar.svelte',
          desc: 'Gestão do evento: dados, perguntas e respostas.'
        },
        {
          route: '/palco/{id}',
          view: 'Palco.svelte',
          desc: 'Painel de controle da apresentação ao vivo (revelar, mensagens, interações).'
        },
        {
          route: '/palco/{id}/apresentar',
          view: 'PalcoApresentar.svelte',
          desc: 'Janela de projeção do palco — o que vai para o telão.'
        }
      ]
    },
    {
      title: 'Internas — desenvolvimento',
      access: 'Interno',
      desc: 'Fora do fluxo do produto: referências e ferramentas internas de desenvolvimento.',
      items: [
        {
          route: '/tela',
          view: 'Tela.svelte',
          desc: 'Guia de componentes e design tokens — a referência viva do design.'
        },
        {
          route: '/mapa-do-site',
          view: 'MapaDoSite.svelte',
          desc: 'Este mapa do site — todos os URLs da aplicação.'
        },
        {
          route: '/debug',
          view: 'Debug.svelte',
          desc: 'Tela interna de desenvolvimento, com atalhos para /mapa-do-site e /tela.'
        }
      ]
    }
  ];

  const accessMeta = {
    Público: { cls: 'ok', label: 'Público' },
    Participante: { cls: 'cyan', label: 'Participante' },
    Organizador: { cls: 'purple', label: 'Organizador' },
    Interno: { cls: 'warn', label: 'Interno' }
  };

  // Rota estática (sem placeholder {id}) vira link; a dinâmica fica como padrão.
  function isStatic(route) {
    return !route.includes('{');
  }
</script>

<main class="page page-wide mapa-do-site">
  <div class="sm-head">
    <h1>Mapa do site — Arandu</h1>
    <p class="text-muted">
      Todas as URLs da aplicação, agrupadas por fluxo. Rotas com <code>{'{id}'}</code> são dinâmicas e
      precisam de um identificador real (ex.: <code>/evento/42</code>); as demais são links
      clicáveis. Página interna de desenvolvimento — não faz parte do fluxo do produto.
    </p>
  </div>

  {#each groups as g (g.title)}
    <section class="sm-section">
      <div class="sm-section-head">
        <h2>{g.title}</h2>
        <span class="sm-access {accessMeta[g.access].cls}">{accessMeta[g.access].label}</span>
      </div>
      <p class="text-muted sm-section-desc">{g.desc}</p>
      <div class="sm-table">
        <div class="sm-row sm-row-head">
          <div class="sm-route">Rota</div>
          <div class="sm-view">View</div>
          <div class="sm-desc">Descrição</div>
        </div>
        {#each g.items as item (item.route)}
          <div class="sm-row">
            <div class="sm-route">
              {#if isStatic(item.route)}
                <a href={item.route} on:click={go(item.route)}>{item.route}</a>
              {:else}
                <code>{item.route}</code>
              {/if}
            </div>
            <div class="sm-view text-muted">{item.view}</div>
            <div class="sm-desc">{item.desc}</div>
          </div>
        {/each}
      </div>
    </section>
  {/each}

  <footer class="sm-footer">
    <a href="/debug" on:click={go('/debug')}>‹ Debug</a>
    <a href="/tela" on:click={go('/tela')}>Guia de componentes (/tela)</a>
  </footer>
</main>

<style>
  .mapa-do-site {
    align-items: stretch;
    gap: 24px;
    padding-bottom: 64px;
  }

  .sm-head h1 {
    margin: 0 0 8px;
    font-size: 2rem;
  }

  .sm-head p {
    max-width: 680px;
    margin: 0;
  }

  .sm-head code {
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 1px 5px;
    font-size: 0.85em;
  }

  .sm-section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    padding: 20px 24px;
  }

  .sm-section-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
  }

  .sm-section-head h2 {
    margin: 0;
    font-size: 1.15rem;
  }

  .sm-section-desc {
    margin: 6px 0 16px;
    max-width: 680px;
  }

  .sm-access {
    font-size: 0.6875rem;
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    border-radius: 999px;
    padding: 3px 10px;
    flex-shrink: 0;
  }

  .sm-access.ok {
    background: var(--tint-success);
    color: var(--success);
  }

  .sm-access.cyan {
    background: var(--tint-cyan);
    color: var(--cyan);
  }

  .sm-access.purple {
    background: var(--tint-purple);
    color: var(--purple);
  }

  .sm-access.warn {
    background: var(--tint-yellow);
    color: var(--yellow);
  }

  .sm-table {
    display: flex;
    flex-direction: column;
  }

  .sm-row {
    display: grid;
    grid-template-columns: 180px 190px 1fr;
    gap: 12px;
    padding: 10px 0;
    border-top: 1px solid var(--border);
    align-items: baseline;
  }

  .sm-row-head {
    border-top: none;
    font-size: 0.6875rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--text-subtle);
    padding-bottom: 6px;
  }

  .sm-route {
    font-family: var(--font-mono);
    font-size: 0.85rem;
    font-weight: 700;
    min-width: 0;
    overflow-wrap: break-word;
  }

  .sm-route code {
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 1px 5px;
    font-size: 0.85em;
  }

  .sm-view {
    font-family: var(--font-mono);
    font-size: 0.78rem;
    min-width: 0;
    overflow-wrap: break-word;
  }

  .sm-desc {
    font-size: 0.875rem;
    min-width: 0;
  }

  .sm-footer {
    display: flex;
    justify-content: center;
    gap: 24px;
    flex-wrap: wrap;
    font-size: 0.8125rem;
    font-weight: 700;
  }

  .sm-footer a {
    color: var(--text-muted);
  }

  .sm-footer a:hover {
    color: var(--accent);
  }

  @media (max-width: 760px) {
    .sm-row {
      grid-template-columns: 1fr;
      gap: 3px;
      padding: 12px 0;
    }

    .sm-row-head {
      display: none;
    }
  }
</style>
