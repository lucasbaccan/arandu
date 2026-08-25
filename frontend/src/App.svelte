<script>
  import { onMount } from 'svelte';
  import { initAuth, authReady, user } from './lib/authStore.js';
  import { route, navigate } from './lib/router.js';
  // Importar já aplica o tema salvo em <html data-theme>, antes da primeira tela.
  import './lib/themeStore.js';
  import Particles from './components/Particles.svelte';
  import Toast from './components/Toast.svelte';
  import Inicio from './views/Inicio.svelte';
  import Entrar from './views/Entrar.svelte';
  import CriarConta from './views/CriarConta.svelte';
  import Painel from './views/Painel.svelte';
  import EventoNovo from './views/EventoNovo.svelte';
  import EventoEditar from './views/EventoEditar.svelte';
  import Responder from './views/Responder.svelte';
  import Palco from './views/Palco.svelte';
  import PalcoApresentar from './views/PalcoApresentar.svelte';
  import Plateia from './views/Plateia.svelte';
  import ComoFunciona from './views/ComoFunciona.svelte';
  import Privacidade from './views/Privacidade.svelte';
  import Tela from './views/Tela.svelte';
  import MapaDoSite from './views/MapaDoSite.svelte';
  import Debug from './views/Debug.svelte';

  onMount(initAuth);

  $: eventoMatch = /^\/eventos\/(\d+)$/.exec($route);
  $: responderMatch = /^\/responder\/(\d+)$/.exec($route);
  $: palcoMatch = /^\/palco\/(\d+)$/.exec($route);
  // A densidade do placar (auto/smart/1–4 colunas) não é mais parte da
  // rota — é estado ao vivo (modoDensidadeApresentacao), escolhido pelos
  // botões de modo no rodapé de Palco.svelte e refletido em tempo real via
  // SSE. Ver PresentationStage.svelte (props forceCols/smart) e
  // PalcoApresentar.svelte.
  $: palcoApresentarMatch = /^\/palco\/(\d+)\/apresentar$/.exec($route);
  $: plateiaMatch = /^\/plateia\/(\d+)$/.exec($route);

  $: if ($authReady) {
    if ($route === '/' && $user) navigate('/painel');
    if (
      ($route === '/painel' ||
        $route === '/eventos/novo' ||
        eventoMatch ||
        palcoMatch ||
        palcoApresentarMatch) &&
      !$user
    ) {
      navigate('/');
    }
    if (($route === '/entrar' || $route === '/criar-conta') && $user) navigate('/painel');
  }
</script>

<Particles />
<Toast />
<div class="app-view">
  {#if !$authReady}
    <main class="page">
      <img class="loading-logo" src="/img/arandu-completo.png" alt="Arandu" />
      <p class="home-tagline">Carregando…</p>
    </main>
  {:else if $route === '/'}
    <Inicio />
  {:else if $route === '/entrar'}
    <Entrar />
  {:else if $route === '/criar-conta'}
    <CriarConta />
  {:else if $route === '/painel'}
    <Painel />
  {:else if $route === '/tela'}
    <Tela />
  {:else if $route === '/mapa-do-site'}
    <MapaDoSite />
  {:else if $route === '/debug'}
    <Debug />
  {:else if $route === '/como-funciona'}
    <ComoFunciona />
  {:else if $route === '/privacidade'}
    <Privacidade />
  {:else if $route === '/eventos/novo'}
    <EventoNovo />
  {:else if eventoMatch}
    <EventoEditar id={eventoMatch[1]} />
  {:else if responderMatch}
    <Responder id={responderMatch[1]} />
  {:else if palcoApresentarMatch}
    <PalcoApresentar id={palcoApresentarMatch[1]} />
  {:else if palcoMatch}
    <Palco id={palcoMatch[1]} />
  {:else if plateiaMatch}
    <Plateia id={plateiaMatch[1]} />
  {:else}
    <main class="page">
      <p>Página não encontrada.</p>
      <a href="/" on:click|preventDefault={() => navigate('/')}>Voltar ao início</a>
    </main>
  {/if}
</div>
