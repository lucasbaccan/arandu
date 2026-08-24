<script>
  /*
   * Topbar de 56px do shell do organizador: só identidade e conta (símbolo,
   * área, tema, ajuda, avatar). Nunca muda de tela para tela — contexto e
   * ações da tela vivem no CrumbBar logo abaixo.
   */
  import { user, logout } from '../lib/authStore.js';
  import { navigate } from '../lib/router.js';
  import ThemeToggle from './ThemeToggle.svelte';
  import HelpButton from './HelpButton.svelte';

  export let area = 'Organizador';
  // Telas públicas com topbar (Responder, Plateia no celular) não têm conta.
  export let showAccount = true;

  let menuOpen = false;
  let avatarEl;
  let menuEl;

  $: initial = (($user && $user.name) || '?')[0].toUpperCase();

  function toggleMenu() {
    menuOpen = !menuOpen;
  }

  function handleWindowMousedown(e) {
    if (!menuOpen) return;
    if (avatarEl?.contains(e.target) || menuEl?.contains(e.target)) return;
    menuOpen = false;
  }

  async function handleLogout() {
    menuOpen = false;
    await logout();
    navigate('/');
  }
</script>

<svelte:window on:mousedown={handleWindowMousedown} />

<header class="topbar">
  <img class="topbar-logo" src="/img/arandu-logo.png" alt="Arandu" />
  <span class="topbar-sep"></span>
  <span class="topbar-area">{area}</span>
  <span class="topbar-spacer"></span>
  <slot name="extra" />
  <ThemeToggle />
  <HelpButton />
  {#if showAccount}
    <div class="user-menu-wrap">
      <button
        type="button"
        class="avatar-btn"
        bind:this={avatarEl}
        title={($user && $user.name) || ''}
        aria-haspopup="menu"
        aria-expanded={menuOpen}
        on:click={toggleMenu}
      >
        {initial}
      </button>
      {#if menuOpen}
        <div class="user-menu" role="menu" bind:this={menuEl}>
          <p class="user-menu-name">Olá, {($user && $user.name) || '…'}</p>
          <button type="button" class="user-menu-item" role="menuitem" on:click={handleLogout}>
            Sair
          </button>
        </div>
      {/if}
    </div>
  {:else}
    <slot name="account" />
  {/if}
</header>

<style>
  .topbar {
    position: sticky;
    top: 0;
    z-index: 20;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 12px;
    height: var(--topbar-h);
    padding: 0 20px;
    background: var(--bg-elev);
    border-bottom: 1px solid var(--border);
  }

  .topbar-logo {
    height: 24px;
    width: auto;
    flex-shrink: 0;
  }

  .topbar-sep {
    width: 1px;
    height: 22px;
    background: var(--border);
    flex-shrink: 0;
  }

  .topbar-area {
    min-width: 0;
    font-size: 0.875rem;
    font-weight: 700;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .topbar-spacer {
    flex: 1;
  }

  .user-menu-wrap {
    position: relative;
    flex-shrink: 0;
  }

  .avatar-btn {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: none;
    background: var(--accent);
    color: var(--on-accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-ui);
    font-size: 0.8125rem;
    font-weight: 700;
    cursor: pointer;
  }

  .user-menu {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    z-index: 30;
    display: flex;
    flex-direction: column;
    min-width: 160px;
    padding: 8px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    box-shadow: var(--shadow);
  }

  .user-menu-name {
    margin: 2px 8px 6px;
    font-size: 0.8125rem;
    color: var(--text-muted);
    white-space: nowrap;
  }

  .user-menu-item {
    text-align: left;
    padding: 8px 10px;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--danger);
    font-family: var(--font-ui);
    font-size: 0.875rem;
    font-weight: 700;
    cursor: pointer;
  }

  .user-menu-item:hover {
    background: var(--tint-danger);
  }
</style>
