<script>
  import { api } from '../lib/api.js';
  import { user } from '../lib/authStore.js';
  import { formatDateTime } from '../lib/formatDate.js';
  import { showToast } from '../lib/toastStore.js';
  import { navigate } from '../lib/router.js';
  import Button from '../components/Button.svelte';
  import Chip from '../components/Chip.svelte';
  import CopyButton from '../components/CopyButton.svelte';
  import CrumbBar from '../components/CrumbBar.svelte';
  import PinChip from '../components/PinChip.svelte';
  import Switch from '../components/Switch.svelte';
  import Tabs from '../components/Tabs.svelte';
  import TopBar from '../components/TopBar.svelte';
  import { statusInfo } from '../lib/eventStatus.js';

  const TABS = [
    { value: 'usuarios', label: 'Usuários' },
    { value: 'eventos', label: 'Eventos' },
    { value: 'configuracoes', label: 'Configurações' }
  ];

  let tab = 'usuarios';

  // --- Usuários ---
  let usuarios = [];
  let usuariosLoading = true;
  let usuariosError = '';
  // id do usuário -> { url, expiresAt } do último link gerado (mostrado inline
  // até a pessoa fechar ou a tela ser recarregada — o token já foi copiado
  // pro clipboard, não precisa persistir).
  let linksGerados = {};

  async function carregarUsuarios() {
    usuariosLoading = true;
    try {
      const data = await api.admin.usuarios.listar();
      usuarios = data.users || [];
    } catch (e) {
      usuariosError = e.message;
    } finally {
      usuariosLoading = false;
    }
  }
  carregarUsuarios();

  function roleInfo(role) {
    if (role === 'super_admin') return { label: 'Super admin', tint: 'var(--tint-purple)', color: 'var(--purple-text)' };
    return { label: 'Organizador', tint: 'var(--tint-neutral)', color: 'var(--text-muted)' };
  }

  async function gerarLink(u) {
    try {
      const { token, expiresAt } = await api.admin.usuarios.gerarLinkRedefinicao(u.id);
      const url = `${window.location.origin}/redefinir-senha?token=${token}`;
      linksGerados = { ...linksGerados, [u.id]: { url, expiresAt } };
    } catch (e) {
      showToast(e.message, 'error');
    }
  }

  function fecharLink(u) {
    const { [u.id]: _omit, ...rest } = linksGerados;
    linksGerados = rest;
  }

  async function excluirUsuario(u) {
    if (
      !window.confirm(
        `Excluir a conta de ${u.name} (${u.email})? Isso também apaga os eventos dela. Essa ação não pode ser desfeita.`
      )
    ) {
      return;
    }
    try {
      await api.admin.usuarios.excluir(u.id);
      usuarios = usuarios.filter((x) => x.id !== u.id);
      showToast('Conta excluída.');
    } catch (e) {
      showToast(e.message, 'error');
    }
  }

  // --- Eventos ---
  let eventos = [];
  let eventosLoading = true;
  let eventosError = '';

  async function carregarEventos() {
    eventosLoading = true;
    try {
      const data = await api.admin.eventos.listar();
      eventos = data.events || [];
    } catch (e) {
      eventosError = e.message;
    } finally {
      eventosLoading = false;
    }
  }
  carregarEventos();

  // --- Configurações ---
  let configLoading = true;
  let configError = '';
  let registrationEnabled = true;
  let savingConfig = false;

  async function carregarConfiguracoes() {
    configLoading = true;
    try {
      const data = await api.admin.configuracoes.buscar();
      registrationEnabled = data.registrationEnabled;
    } catch (e) {
      configError = e.message;
    } finally {
      configLoading = false;
    }
  }
  carregarConfiguracoes();

  async function alternarRegistro() {
    const next = !registrationEnabled;
    savingConfig = true;
    try {
      await api.admin.configuracoes.atualizar({ registrationEnabled: next });
      registrationEnabled = next;
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      savingConfig = false;
    }
  }
</script>

<main class="shell">
  <TopBar area="Administração" />
  <CrumbBar crumbs={[{ label: 'Administração' }]} />

  <div class="shell-body">
    <Tabs tabs={TABS} bind:value={tab} />

    {#if tab === 'usuarios'}
      <section class="tab-panel">
        {#if usuariosLoading}
          <p class="text-muted">Carregando…</p>
        {:else if usuariosError}
          <p class="form-error">{usuariosError}</p>
        {:else}
          <div class="table-head">
            <span class="col-name">Nome</span>
            <span class="col-role">Papel</span>
            <span class="col-events">Eventos</span>
            <span class="col-created">Criado em</span>
            <span class="col-actions">Ações</span>
          </div>
          <div class="table-body">
            {#each usuarios as u (u.id)}
              <div class="row">
                <div class="row-main">
                  <div class="col-name">
                    <span class="user-name">{u.name}</span>
                    <span class="text-muted user-email">{u.email}</span>
                  </div>
                  <span class="col-role">
                    <Chip label={roleInfo(u.role).label} tint={roleInfo(u.role).tint} color={roleInfo(u.role).color} />
                  </span>
                  <span class="col-events text-muted">{u.eventCount}</span>
                  <span class="col-created text-muted">{formatDateTime(u.createdAt)}</span>
                  <span class="col-actions">
                    <Button size="sm" variant="secondary" on:click={() => gerarLink(u)}>Gerar link de redefinição</Button>
                    {#if $user && u.id !== $user.id}
                      <Button size="sm" variant="danger" on:click={() => excluirUsuario(u)}>Excluir</Button>
                    {/if}
                  </span>
                </div>
                {#if linksGerados[u.id]}
                  <div class="link-panel">
                    <code class="link-text">{linksGerados[u.id].url}</code>
                    <CopyButton text={linksGerados[u.id].url} label="Copiar link" />
                    <span class="text-muted link-expiry">
                      Expira às {formatDateTime(linksGerados[u.id].expiresAt)}
                    </span>
                    <button type="button" class="link-close" aria-label="Fechar" on:click={() => fecharLink(u)}>✕</button>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </section>
    {:else if tab === 'eventos'}
      <section class="tab-panel">
        {#if eventosLoading}
          <p class="text-muted">Carregando…</p>
        {:else if eventosError}
          <p class="form-error">{eventosError}</p>
        {:else if eventos.length === 0}
          <p class="text-muted">Nenhum evento ainda.</p>
        {:else}
          <div class="table-head events-grid">
            <span>Evento</span>
            <span>Dono</span>
            <span>Situação</span>
            <span>PIN</span>
          </div>
          <div class="table-body">
            {#each eventos as ev (ev.id)}
              <div
                class="row events-grid"
                role="button"
                tabindex="0"
                on:click={() => navigate(`/evento/${ev.id}`)}
                on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && navigate(`/evento/${ev.id}`)}
              >
                <span class="evt-title">{ev.title}</span>
                <span class="col-owner">
                  <span class="user-name">{ev.ownerName}</span>
                  <span class="text-muted user-email">{ev.ownerEmail}</span>
                </span>
                <span>
                  <Chip dot label={statusInfo(ev.status).label} tint={statusInfo(ev.status).tint} color={statusInfo(ev.status).color} />
                </span>
                <span><PinChip pin={ev.pinCode} /></span>
              </div>
            {/each}
          </div>
        {/if}
      </section>
    {:else}
      <section class="tab-panel">
        {#if configLoading}
          <p class="text-muted">Carregando…</p>
        {:else if configError}
          <p class="form-error">{configError}</p>
        {:else}
          <div class="config-row">
            <div class="config-text">
              <span class="config-label">Permitir novos cadastros de organizador</span>
              <span class="text-muted config-hint">
                Desativado, a tela de criar conta some para quem não está logado — só quem já tem conta continua entrando.
              </span>
            </div>
            <Switch checked={registrationEnabled} disabled={savingConfig} on:change={alternarRegistro} />
          </div>
        {/if}
      </section>
    {/if}
  </div>
</main>

<style>
  .shell {
    flex: none;
    height: 100vh;
    height: 100dvh;
  }

  .shell-body {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .tab-panel {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .table-head {
    display: grid;
    grid-template-columns: minmax(160px, 1.4fr) 140px 90px 170px minmax(200px, auto);
    gap: 16px;
    padding: 0 18px;
    color: var(--text-muted);
    font-size: 0.6875rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }

  .table-body {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .row {
    display: flex;
    flex-direction: column;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
  }

  .row-main {
    display: grid;
    grid-template-columns: minmax(160px, 1.4fr) 140px 90px 170px minmax(200px, auto);
    gap: 16px;
    align-items: center;
    min-height: 64px;
    padding: 12px 18px;
  }

  .events-grid {
    grid-template-columns: minmax(160px, 1.4fr) minmax(160px, 1fr) 170px 140px;
  }

  .row.events-grid {
    display: grid;
    cursor: pointer;
    align-items: center;
    min-height: 56px;
    padding: 12px 18px;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
  }

  .row.events-grid:hover,
  .row.events-grid:focus-visible {
    border-color: var(--accent);
    box-shadow: var(--shadow-hover);
    outline: none;
  }

  .col-name,
  .col-owner {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .user-name {
    font-weight: 700;
    font-size: 0.9375rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .user-email {
    font-size: 0.75rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .evt-title {
    font-weight: 700;
    font-size: 0.9375rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .col-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
    flex-wrap: wrap;
  }

  .link-panel {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 0 18px 12px;
    padding: 10px 12px;
    border-radius: var(--radius-control);
    background: var(--surface-muted);
    border: 1px solid var(--border);
    flex-wrap: wrap;
  }

  .link-text {
    flex: 1;
    min-width: 200px;
    font-size: 0.8125rem;
    overflow-wrap: anywhere;
  }

  .link-expiry {
    font-size: 0.75rem;
    white-space: nowrap;
  }

  .link-close {
    border: none;
    background: transparent;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 0.875rem;
    padding: 4px;
  }

  .link-close:hover {
    color: var(--text);
  }

  .config-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 16px 18px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius-row);
  }

  .config-text {
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-width: 520px;
  }

  .config-label {
    font-weight: 700;
    font-size: 0.9375rem;
  }

  .config-hint {
    font-size: 0.8125rem;
  }

  @media (max-width: 900px) {
    .table-head {
      display: none;
    }

    .row-main {
      grid-template-columns: 1fr;
      gap: 8px;
    }

    .col-actions {
      justify-content: flex-start;
    }

    .events-grid {
      grid-template-columns: 1fr;
    }

    .row.events-grid {
      gap: 6px;
    }
  }
</style>
