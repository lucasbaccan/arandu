<script>
  import { fly, fade } from 'svelte/transition';
  import { flip } from 'svelte/animate';

  // pending: participantes ainda não revelados pra pergunta atual (id, email, photo).
  export let pending = [];
  // groups: buckets já visíveis desde o início (opções da pergunta, ou
  // respostas abertas distintas) — { label, participants: [...] }.
  export let groups = [];
  // onFaceClick: se informado, os rostos pendentes viram botão clicável
  // (tela do admin). Se omitido, ficam só leitura (tela pública).
  export let onFaceClick = null;
  // hideZones: esconde as zonas de resposta (opções e quem foi revelado
  // nelas) — usado pela plateia quando o organizador ativa "esconder
  // respostas". A fila de pendentes continua visível independente disso.
  export let hideZones = false;
  // showNames: mostra a legenda de nome sob cada rosto, pendente ou
  // revelado — controlado pelo switch "Ocultar nomes" (estado do servidor),
  // refletido tanto na janela de apresentação do organizador
  // (StagePresentation) quanto na tela da plateia (Audience). O painel do
  // próprio organizador (/stage) não usa este componente e sempre mostra os
  // nomes, independente do switch.
  export let showNames = false;

  $: totalParticipants =
    pending.length + groups.reduce((sum, g) => sum + g.participants.length, 0);

  function firstName(p) {
    return (p.name || p.email).split(' ')[0];
  }
</script>

<div class="pending-row">
  {#each pending as p (p.id)}
    <div class="face-wrap" animate:flip={{ duration: 350 }} out:fade={{ duration: 150 }}>
      <button
        type="button"
        class="face pending"
        class:static={!onFaceClick}
        disabled={!onFaceClick}
        title={p.name || p.email}
        aria-label={onFaceClick ? `Revelar resposta de ${p.name || p.email}` : p.name || p.email}
        on:click={() => onFaceClick && onFaceClick(p)}
      >
        {#if p.photo}
          <img src={p.photo} alt="" />
        {:else}
          <span class="face-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
        {/if}
      </button>
      {#if showNames}
        <span class="face-name">{firstName(p)}</span>
      {/if}
    </div>
  {/each}
  {#if pending.length === 0}
    <p class="text-muted present-empty">
      {totalParticipants === 0 ? 'Ninguém respondeu ainda.' : 'Todas as respostas foram reveladas.'}
    </p>
  {/if}
</div>

<div class="zones">
  {#if hideZones}
    <p class="text-muted present-empty">O organizador escondeu as respostas por enquanto.</p>
  {:else}
    {#if groups.length === 0}
      <p class="text-muted present-empty">Ninguém respondeu ainda.</p>
    {/if}
    {#each groups as group (group.label)}
      <div class="zone">
        <div class="zone-label">
          <span>{group.label}</span>
          <span class="zone-count">{group.participants.length}</span>
        </div>
        <div class="zone-faces">
          {#each group.participants as p (p.id)}
            <div class="face-wrap" animate:flip={{ duration: 350 }} in:fly={{ y: -30, duration: 350 }}>
              <button
                type="button"
                class="face"
                class:static={!onFaceClick}
                disabled={!onFaceClick}
                title={p.name || p.email}
                aria-label={onFaceClick ? `Desrevelar resposta de ${p.name || p.email}` : p.name || p.email}
                on:click={() => onFaceClick && onFaceClick(p)}
              >
                {#if p.photo}
                  <img src={p.photo} alt="" />
                {:else}
                  <span class="face-placeholder">{(p.name || p.email)[0].toUpperCase()}</span>
                {/if}
              </button>
              {#if showNames}
                <span class="face-name">{firstName(p)}</span>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {/each}
  {/if}
</div>

<style>
  .present-empty {
    margin: 0;
  }

  .pending-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    min-height: 64px;
    padding: 12px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 12px;
  }

  .face-wrap {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 5px;
    width: 60px;
  }

  .face-name {
    max-width: 60px;
    font-size: 0.7rem;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .face {
    width: 52px;
    height: 52px;
    flex-shrink: 0;
    border-radius: 50%;
    border: none;
    padding: 0;
    overflow: hidden;
    cursor: pointer;
    background: var(--accent);
    transition: transform 0.1s ease;
  }

  .face:not(:disabled):hover {
    transform: scale(1.05);
  }

  .face.static {
    cursor: default;
    display: block;
  }

  .face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .face-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--accent);
    color: #fff;
    font-weight: 600;
  }

  /* Pendente (ainda não revelado): só a borda tracejada marca a diferença —
     a "arte" do rosto (foto ou inicial) é a mesma de quando revelado, igual
     ao avatar do painel de respostas (ResponsesPanel), pra não ter dois
     estilos de avatar diferentes na mesma pessoa. */
  .face.pending {
    border: 2px dashed var(--border-strong);
  }

  .face.pending:not(:disabled):hover {
    border-color: var(--accent);
  }

  .zones {
    flex: 1;
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
  }

  .zone {
    flex: 1;
    min-width: 220px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 14px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 12px;
  }

  .zone-label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-weight: 600;
  }

  .zone-count {
    color: var(--text-muted);
    font-weight: 400;
  }

  .zone-faces {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    min-height: 52px;
  }
</style>
