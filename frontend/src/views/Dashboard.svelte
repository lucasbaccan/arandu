<script>
  import { user, logout } from '../lib/authStore.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Card from '../components/Card.svelte';

  async function handleLogout() {
    await logout();
    navigate('/');
  }
</script>

<main class="page">
  <Card title={$user ? `Olá, ${$user.name}` : 'Olá!'}>
    <p class="subtitle">Seu painel está em construção.</p>
    <Button block on:click={handleLogout}>Sair</Button>
  </Card>
</main>
