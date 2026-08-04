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

  $: totalParticipants =
    pending.length + groups.reduce((sum, g) => sum + g.participants.length, 0);
</script>

<div class="pending-row">
  {#each pending as p (p.id)}
    <button
      type="button"
      class="face"
      class:static={!onFaceClick}
      disabled={!onFaceClick}
      title={p.email}
      aria-label={onFaceClick ? `Revelar resposta de ${p.email}` : p.email}
      on:click={() => onFaceClick && onFaceClick(p)}
      animate:flip={{ duration: 350 }}
      out:fade={{ duration: 150 }}
    >
      {#if p.photo}
        <img src={p.photo} alt="" />
      {:else}
        <span class="face-placeholder">{p.email[0].toUpperCase()}</span>
      {/if}
    </button>
  {/each}
  {#if pending.length === 0}
    <p class="text-muted present-empty">
      {totalParticipants === 0 ? 'Ninguém respondeu ainda.' : 'Todas as respostas foram reveladas.'}
    </p>
  {/if}
</div>

<div class="zones">
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
          <span
            class="face static"
            title={p.email}
            animate:flip={{ duration: 350 }}
            in:fly={{ y: -30, duration: 350 }}
          >
            {#if p.photo}
              <img src={p.photo} alt="" />
            {:else}
              <span class="face-placeholder">{p.email[0].toUpperCase()}</span>
            {/if}
          </span>
        {/each}
      </div>
    </div>
  {/each}
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

  .face {
    width: 52px;
    height: 52px;
    flex-shrink: 0;
    border-radius: 50%;
    border: 2px solid var(--border);
    padding: 0;
    overflow: hidden;
    cursor: pointer;
    background: var(--bg-input);
    transition: border-color 0.15s ease, transform 0.1s ease;
  }

  .face:not(:disabled):hover {
    border-color: var(--accent);
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
