<script>
  import { onMount } from 'svelte';
  import { initAuth, authReady, user } from './lib/authStore.js';
  import { route, navigate } from './lib/router.js';
  import Particles from './components/Particles.svelte';
  import Toast from './components/Toast.svelte';
  import Home from './views/Home.svelte';
  import Login from './views/Login.svelte';
  import Register from './views/Register.svelte';
  import Dashboard from './views/Dashboard.svelte';
  import EventCreate from './views/EventCreate.svelte';
  import EventEdit from './views/EventEdit.svelte';

  onMount(initAuth);

  $: eventMatch = /^\/events\/(\d+)$/.exec($route);

  $: if ($authReady) {
    if ($route === '/' && $user) navigate('/dashboard');
    if (($route === '/dashboard' || $route === '/events/new' || eventMatch) && !$user) {
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
      <h1 class="home-logo">DevOps Conecta</h1>
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
  {:else if $route === '/events/new'}
    <EventCreate />
  {:else if eventMatch}
    <EventEdit id={eventMatch[1]} />
  {:else}
    <main class="page">
      <p>Página não encontrada.</p>
      <a href="/" on:click|preventDefault={() => navigate('/')}>Voltar ao início</a>
    </main>
  {/if}
</div>
