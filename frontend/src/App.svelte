<script>
  import { onMount } from 'svelte';
  import { initAuth, authReady, user } from './lib/authStore.js';
  import { route, navigate } from './lib/router.js';
  import Home from './views/Home.svelte';
  import Login from './views/Login.svelte';
  import Register from './views/Register.svelte';
  import Dashboard from './views/Dashboard.svelte';
  import EventCreate from './views/EventCreate.svelte';

  onMount(initAuth);

  $: if ($authReady) {
    if ($route === '/' && $user) navigate('/dashboard');
    if (
      ($route === '/login' ||
        $route === '/register' ||
        $route === '/dashboard' ||
        $route === '/events/new') &&
      !$user
    ) {
      navigate('/');
    }
  }
</script>

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
{:else}
  <main class="page">
    <p>Página não encontrada.</p>
    <a href="/" on:click|preventDefault={() => navigate('/')}>Voltar ao início</a>
  </main>
{/if}
