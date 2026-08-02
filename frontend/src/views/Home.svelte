<script>
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
</script>

<main class="home">
  <div class="home-content">
    <h1 class="home-logo">DevOps Conecta</h1>
    <p class="home-tagline">Dinâmicas de grupo ao vivo</p>
    <div class="home-actions">
      <Button on:click={() => navigate('/login')}>Entrar</Button>
      <Button variant="secondary" on:click={() => navigate('/register')}>
        Criar conta
      </Button>
    </div>
  </div>
</main>

<style>
  .home {
    flex: 1;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .home-content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .home-actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: 8px;
  }

  .home-actions :global(.btn) {
    min-width: 240px;
  }
</style>
