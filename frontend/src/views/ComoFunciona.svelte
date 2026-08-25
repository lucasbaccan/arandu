<script>
  import TopBar from '../components/TopBar.svelte';
  import { navigate } from '../lib/router.js';

  function go(path) {
    return (e) => {
      e.preventDefault();
      navigate(path);
    };
  }
</script>

<div class="shell">
  <TopBar area="Como funciona" showAccount={false} />

  <main class="content">
    <div class="intro">
      <h1 class="section-title">Como funciona o Arandu</h1>
      <p class="section-sub">
        Uma dinâmica ao vivo pra o grupo se conhecer melhor: quem organiza monta as perguntas e
        conduz a apresentação; quem participa entra só com um código, sem conta e sem instalar
        nada.
      </p>
    </div>

    <section class="como-section">
      <h2 class="como-heading">Para quem participa</h2>
      <ol class="como-steps">
        <li>
          <strong>Recebe o código do evento.</strong> Quem organiza compartilha um código (PIN) ou
          um link — no grupo, por e-mail, do jeito que preferir.
        </li>
        <li>
          <strong>Entra sem conta.</strong> Digita o código em arandu, informa nome e e-mail e, se
          quiser, envia uma foto de rosto.
        </li>
        <li>
          <strong>Responde no seu ritmo.</strong> As perguntas podem ser de escolha única, texto
          livre ou por grupos — dá pra voltar e mudar qualquer resposta antes do evento.
        </li>
        <li>
          <strong>Acompanha o telão.</strong> Sua resposta só aparece pra todo mundo no momento em
          que quem organiza revelar você.
        </li>
      </ol>
    </section>

    <section class="como-section">
      <h2 class="como-heading">Para quem organiza</h2>
      <ol class="como-steps">
        <li>
          <strong>Cria uma conta e um evento.</strong> Define um título e um código (PIN) pro
          grupo usar.
        </li>
        <li>
          <strong>Monta as perguntas.</strong> Escolha única, texto livre ou por grupos — na ordem
          que fizer sentido pra dinâmica.
        </li>
        <li>
          <strong>Abre as respostas.</strong> Compartilha o código ou o link e acompanha, em tempo
          real, quem já respondeu.
        </li>
        <li>
          <strong>Conduz a apresentação ao vivo.</strong> Revela participante por participante, com
          uma janela própria pra projetar na tela ou compartilhar numa chamada.
        </li>
      </ol>
    </section>

    <section class="como-section">
      <h2 class="como-heading">Seus dados</h2>
      <p class="text-muted">
        As respostas e fotos existem só pra dinâmica de cada evento — não viram anúncio, lista de
        contatos nem relatório cruzado com outros eventos.
        <a href="/privacidade" on:click={go('/privacidade')}>Veja a política de privacidade completa</a>.
      </p>
    </section>
  </main>

  <footer class="como-footer">
    <a href="/" on:click={go('/')}>‹ Voltar ao início</a>
    <a href="/privacidade" on:click={go('/privacidade')}>Política de privacidade</a>
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

  .como-section {
    width: 100%;
    max-width: 640px;
    margin-top: 32px;
    padding-top: 28px;
    border-top: 1px solid var(--border);
  }

  .como-heading {
    margin: 0 0 16px;
    font-size: 1.0625rem;
    font-weight: 800;
  }

  .como-steps {
    margin: 0;
    padding-left: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    color: var(--text-muted);
    font-size: 0.9375rem;
    line-height: 1.5;
  }

  .como-steps strong {
    color: var(--text);
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
