<script>
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
</script>

<main class="page">
  <h1 class="home-logo">DevOps Conecta</h1>
  <p class="home-tagline">Dinâmicas de grupo ao vivo</p>
  <Button block on:click={() => navigate('/login')}>Entrar</Button>
  <Button block variant="secondary" on:click={() => navigate('/register')}>
    Criar conta
  </Button>
</main>
