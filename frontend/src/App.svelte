<script>
  import { onMount } from 'svelte';
  import { initAuth, user } from './lib/authStore.js';
  import { route, navigate } from './lib/router.js';
  import Home from './views/Home.svelte';
  import Login from './views/Login.svelte';
  import Register from './views/Register.svelte';
  import Dashboard from './views/Dashboard.svelte';

  onMount(initAuth);

  $: if ($route === '/') {
    if ($user) navigate('/dashboard');
  }
  $: if ($route === '/dashboard' && !$user) navigate('/');
</script>

{#if $route === '/'}
  <Home />
{:else if $route === '/login'}
  <Login />
{:else if $route === '/register'}
  <Register />
{:else if $route === '/dashboard'}
  <Dashboard />
{:else}
  <main class="page">
    <p>Página não encontrada.</p>
    <a href="/" on:click|preventDefault={() => navigate('/')}>Voltar ao início</a>
  </main>
{/if}
