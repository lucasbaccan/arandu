<script>
  import { onMount } from 'svelte';
  import { initAuth, authReady, user } from './lib/authStore.js';
  import { route, navigate } from './lib/router.js';
  // Importar já aplica o tema salvo em <html data-theme>, antes da primeira tela.
  import './lib/themeStore.js';
  import Particles from './components/Particles.svelte';
  import Toast from './components/Toast.svelte';
  import Home from './views/Home.svelte';
  import Login from './views/Login.svelte';
  import Register from './views/Register.svelte';
  import Dashboard from './views/Dashboard.svelte';
  import EventCreate from './views/EventCreate.svelte';
  import EventEdit from './views/EventEdit.svelte';
  import Answer from './views/Answer.svelte';
  import Stage from './views/Stage.svelte';
  import StagePresentation from './views/StagePresentation.svelte';
  import Audience from './views/Audience.svelte';
  import StyleGuide from './views/StyleGuide.svelte';

  onMount(initAuth);

  $: eventMatch = /^\/events\/(\d+)$/.exec($route);
  $: answerMatch = /^\/answer\/(\d+)$/.exec($route);
  $: stageMatch = /^\/stage\/(\d+)$/.exec($route);
  // present: placar com colunas automáticas; present1..4: força 1–4 colunas
  // (ver PresentationStage.svelte, prop forceCols, e o menu flutuante
  // PresentVariantMenu.svelte).
  $: stagePresentMatch = /^\/stage\/(\d+)\/present([1-4])?$/.exec($route);
  $: stagePresentCols = stagePresentMatch ? Number(stagePresentMatch[2] || 0) : 0;
  $: audienceMatch = /^\/audience\/(\d+)$/.exec($route);

  $: if ($authReady) {
    if ($route === '/' && $user) navigate('/dashboard');
    if (
      ($route === '/dashboard' ||
        $route === '/events/new' ||
        eventMatch ||
        stageMatch ||
        stagePresentMatch) &&
      !$user
    ) {
      navigate('/');
    }
    if (($route === '/login' || $route === '/register') && $user) navigate('/dashboard');
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
    <Home />
  {:else if $route === '/login'}
    <Login />
  {:else if $route === '/register'}
    <Register />
  {:else if $route === '/dashboard'}
    <Dashboard />
  {:else if $route === '/tela'}
    <StyleGuide />
  {:else if $route === '/events/new'}
    <EventCreate />
  {:else if eventMatch}
    <EventEdit id={eventMatch[1]} />
  {:else if answerMatch}
    <Answer id={answerMatch[1]} />
  {:else if stagePresentMatch}
    <StagePresentation id={stagePresentMatch[1]} cols={stagePresentCols} />
  {:else if stageMatch}
    <Stage id={stageMatch[1]} />
  {:else if audienceMatch}
    <Audience id={audienceMatch[1]} />
  {:else}
    <main class="page">
      <p>Página não encontrada.</p>
      <a href="/" on:click|preventDefault={() => navigate('/')}>Voltar ao início</a>
    </main>
  {/if}
</div>
