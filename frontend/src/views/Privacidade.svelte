<script>
  import TopBar from '../components/TopBar.svelte';
  import { navigate } from '../lib/router.js';

  const PRIVACY_ITEMS = [
    {
      title: 'O que pedimos',
      body: 'Nome, e-mail e, se você quiser, uma foto de rosto. Nada além disso.'
    },
    {
      title: 'O que aparece no telão',
      body: 'Nome, foto e a resposta escolhida — e só no momento em que quem organiza revelar você. Antes disso, o servidor não envia sua resposta para a tela pública.'
    },
    {
      title: 'O que ninguém vê',
      body: 'Seu e-mail nunca aparece na apresentação. Ele fica visível apenas para quem organiza o evento, para identificar sua participação.'
    },
    {
      title: 'A foto é opcional',
      body: 'Sem foto, você aparece com a inicial do seu nome. A dinâmica funciona do mesmo jeito.'
    },
    {
      title: 'As respostas ficam no evento',
      body: 'Elas existem para serem mostradas na apresentação daquele evento e nada mais: não viram relatório, não são cruzadas com outros eventos nem compartilhadas fora dele.'
    },
    {
      title: 'Quanto tempo fica guardado',
      body: 'Os dados ficam no evento enquanto ele existir. Quando quem organiza apaga o evento, as respostas e fotos vão junto.'
    }
  ];

  function go(path) {
    return (e) => {
      e.preventDefault();
      navigate(path);
    };
  }
</script>

<div class="shell">
  <TopBar area="Política de privacidade" showAccount={false} />

  <main class="content">
    <div class="intro">
      <h1 class="section-title">Seus dados servem para uma coisa só: a dinâmica</h1>
      <p class="section-sub">
        O Arandu existe para o grupo se conhecer. Nada do que você envia aqui vira anúncio,
        ranking ou lista de contatos.
      </p>
    </div>

    <div class="privacidade-items">
      {#each PRIVACY_ITEMS as p (p.title)}
        <div class="privacidade-item">
          <span class="privacidade-item-title">{p.title}</span>
          <span class="text-muted">{p.body}</span>
        </div>
      {/each}
    </div>

    <p class="privacidade-note text-muted">
      Dúvidas ou pedido de remoção: fale com quem organiza o evento. O Arandu roda na
      infraestrutura de quem hospeda a plataforma; não enviamos seus dados para serviços de
      terceiros.
    </p>
  </main>

  <footer class="como-footer">
    <a href="/" on:click={go('/')}>‹ Voltar ao início</a>
    <a href="/como-funciona" on:click={go('/como-funciona')}>Como funciona o Arandu</a>
  </footer>
</div>

<style>
  .shell {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 48px 24px 40px;
  }

  .intro {
    width: 100%;
    max-width: 640px;
    text-align: center;
  }

  .intro .section-sub {
    margin-top: 8px;
  }

  .privacidade-items {
    width: 100%;
    max-width: 640px;
    margin-top: 32px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .privacidade-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--border);
    font-size: 0.9375rem;
    line-height: 1.5;
  }

  .privacidade-item:last-child {
    border-bottom: none;
  }

  .privacidade-item-title {
    font-weight: 800;
  }

  .privacidade-note {
    width: 100%;
    max-width: 640px;
    margin-top: 24px;
    font-size: 0.8125rem;
    line-height: 1.5;
  }

  .como-footer {
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    gap: 24px;
    padding: 20px 24px;
    border-top: 1px solid var(--border);
    background: var(--bg-elev);
  }

  .como-footer a {
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--text-muted);
  }

  .como-footer a:hover {
    color: var(--accent);
  }
</style>
