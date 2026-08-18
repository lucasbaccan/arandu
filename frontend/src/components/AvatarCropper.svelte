<script>
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';

  const STAGE_SIZE = 280;
  const CROP_SIZE = 200;
  const CROP_OFFSET = (STAGE_SIZE - CROP_SIZE) / 2;
  const OUTPUT_SIZE = 320;
  const MIN_ZOOM = 1;
  const MAX_ZOOM = 3;

  const dispatch = createEventDispatcher();

  // Modo linha: a foto vira uma linha do formulário (círculo de 64px + texto)
  // em vez de um bloco de 200px que empurra os campos para fora da dobra.
  // Ao escolher um arquivo, o editor de recorte aparece normalmente.
  export let compact = false;

  let fileInput;
  let img = null;
  let objectUrl = null;
  let naturalW = 0;
  let naturalH = 0;
  let baseScale = 1;
  let zoom = 1;
  let lastZoom = 1;
  let offsetX = 0;
  let offsetY = 0;
  let dragging = false;
  let dragStartX = 0;
  let dragStartY = 0;
  let dragStartOffsetX = 0;
  let dragStartOffsetY = 0;

  // Reactive versions for template bindings only. Logic below always recomputes
  // from the raw values directly, since these can lag a tick behind an
  // assignment made in the same synchronous handler (Svelte flushes reactive
  // statements on the next microtask, not immediately).
  $: dispW = naturalW * baseScale * zoom;
  $: dispH = naturalH * baseScale * zoom;

  function currentSize(z = zoom) {
    return { w: naturalW * baseScale * z, h: naturalH * baseScale * z };
  }

  function pickFile() {
    fileInput.click();
  }

  function onFileChange(e) {
    const file = e.target.files && e.target.files[0];
    if (!file) return;
    if (objectUrl) URL.revokeObjectURL(objectUrl);
    const url = URL.createObjectURL(file);
    objectUrl = url;
    const el = new Image();
    el.onload = () => {
      naturalW = el.naturalWidth;
      naturalH = el.naturalHeight;
      baseScale = Math.max(CROP_SIZE / naturalW, CROP_SIZE / naturalH);
      zoom = 1;
      lastZoom = 1;
      offsetX = (STAGE_SIZE - naturalW * baseScale) / 2;
      offsetY = (STAGE_SIZE - naturalH * baseScale) / 2;
      img = el;
      exportCrop();
    };
    el.src = url;
  }

  function clampOffsets() {
    const { w, h } = currentSize();
    const minX = CROP_OFFSET + CROP_SIZE - w;
    const maxX = CROP_OFFSET;
    const minY = CROP_OFFSET + CROP_SIZE - h;
    const maxY = CROP_OFFSET;
    offsetX = Math.min(maxX, Math.max(minX, offsetX));
    offsetY = Math.min(maxY, Math.max(minY, offsetY));
  }

  function onZoomInput() {
    // keep whatever image point was at the crop circle's center still centered
    const cropCenterX = CROP_OFFSET + CROP_SIZE / 2;
    const cropCenterY = CROP_OFFSET + CROP_SIZE / 2;
    const ratio = zoom / lastZoom;
    offsetX = cropCenterX - (cropCenterX - offsetX) * ratio;
    offsetY = cropCenterY - (cropCenterY - offsetY) * ratio;
    lastZoom = zoom;
    clampOffsets();
    exportCrop();
  }

  function startDrag(e) {
    if (!img) return;
    dragging = true;
    dragStartX = e.clientX;
    dragStartY = e.clientY;
    dragStartOffsetX = offsetX;
    dragStartOffsetY = offsetY;
    e.currentTarget.setPointerCapture(e.pointerId);
  }

  function onDragMove(e) {
    if (!dragging) return;
    offsetX = dragStartOffsetX + (e.clientX - dragStartX);
    offsetY = dragStartOffsetY + (e.clientY - dragStartY);
    clampOffsets();
  }

  function endDrag() {
    if (!dragging) return;
    dragging = false;
    exportCrop();
  }

  function exportCrop() {
    if (!img) return;
    const { w, h } = currentSize();
    const canvas = document.createElement('canvas');
    canvas.width = OUTPUT_SIZE;
    canvas.height = OUTPUT_SIZE;
    const ctx = canvas.getContext('2d');
    const exportScale = OUTPUT_SIZE / CROP_SIZE;
    ctx.drawImage(
      img,
      (offsetX - CROP_OFFSET) * exportScale,
      (offsetY - CROP_OFFSET) * exportScale,
      w * exportScale,
      h * exportScale
    );
    dispatch('change', canvas.toDataURL('image/jpeg', 0.85));
  }

  function removePhoto() {
    if (objectUrl) URL.revokeObjectURL(objectUrl);
    objectUrl = null;
    img = null;
    naturalW = 0;
    naturalH = 0;
    fileInput.value = '';
    dispatch('change', '');
  }
</script>

<div class="avatar-cropper" class:compact>
  <input
    bind:this={fileInput}
    type="file"
    accept="image/*"
    class="sr-only"
    on:change={onFileChange}
  />

  {#if !img && compact}
    <button type="button" class="picker-row" on:click={pickFile}>
      <span class="picker-row-circle" aria-hidden="true">+</span>
      <span class="picker-row-text">
        <span class="picker-row-title">Adicionar foto</span>
        <span class="picker-row-hint">Opcional — sem ela aparece sua inicial.</span>
      </span>
    </button>
  {:else if !img}
    <button type="button" class="picker" on:click={pickFile}>
      <span class="picker-icon" aria-hidden="true">+</span>
      <span>Adicionar foto</span>
    </button>
  {:else}
    <div class="editor">
      <div
        class="stage"
        style="width: {STAGE_SIZE}px; height: {STAGE_SIZE}px;"
        on:pointerdown={startDrag}
        on:pointermove={onDragMove}
        on:pointerup={endDrag}
        on:pointercancel={endDrag}
      >
        <img
          src={img.src}
          alt=""
          draggable="false"
          style="left: {offsetX}px; top: {offsetY}px; width: {dispW}px; height: {dispH}px;"
        />
        <div class="crop-mask" style="width: {CROP_SIZE}px; height: {CROP_SIZE}px;">
          <span class="guide guide-h"></span>
          <span class="guide guide-v"></span>
        </div>
      </div>
      <p class="text-muted stage-hint">Arraste para posicionar o rosto no círculo</p>
      <label class="zoom-row">
        <span class="text-muted">Zoom</span>
        <input
          type="range"
          min={MIN_ZOOM}
          max={MAX_ZOOM}
          step="0.01"
          bind:value={zoom}
          on:input={onZoomInput}
        />
      </label>
      <div class="editor-actions">
        <Button variant="secondary" type="button" on:click={pickFile}>Trocar foto</Button>
        <Button variant="secondary" type="button" on:click={removePhoto}>Remover</Button>
      </div>
    </div>
  {/if}
</div>

<style>
  .avatar-cropper {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
  }

  .avatar-cropper.compact {
    align-items: stretch;
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .picker {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: 200px;
    height: 200px;
    border-radius: 50%;
    border: 2px dashed var(--border);
    background: var(--bg-input);
    color: var(--text-muted);
    cursor: pointer;
    transition: border-color 0.15s ease, color 0.15s ease;
  }

  .picker:hover {
    border-color: var(--accent);
    color: var(--accent);
  }

  .picker-icon {
    font-size: 1.8rem;
    line-height: 1;
  }

  .picker-row {
    display: flex;
    align-items: center;
    gap: 14px;
    width: 100%;
    padding: 0;
    border: none;
    background: transparent;
    color: var(--text);
    font-family: var(--font-ui);
    text-align: left;
    cursor: pointer;
  }

  .picker-row-circle {
    width: 64px;
    height: 64px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    border: 1.5px dashed var(--border-strong);
    background: var(--surface-muted);
    color: var(--text-subtle);
    font-size: 1.375rem;
    line-height: 1;
    transition: border-color 0.15s ease, color 0.15s ease;
  }

  .picker-row:hover .picker-row-circle {
    border-color: var(--accent);
    color: var(--accent);
  }

  .picker-row-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .picker-row-title {
    font-size: 0.875rem;
    font-weight: 700;
  }

  .picker-row-hint {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .editor {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .stage {
    position: relative;
    overflow: hidden;
    border-radius: 14px;
    background: var(--bg-input);
    cursor: grab;
    touch-action: none;
  }

  .stage:active {
    cursor: grabbing;
  }

  .stage img {
    position: absolute;
    max-width: none;
    user-select: none;
    -webkit-user-drag: none;
  }

  .crop-mask {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    border-radius: 50%;
    border: 2px solid var(--accent);
    box-shadow: 0 0 0 9999px rgba(10, 14, 22, 0.62);
    pointer-events: none;
  }

  .guide {
    position: absolute;
    background: rgba(255, 255, 255, 0.55);
    pointer-events: none;
  }

  .guide-h {
    top: 50%;
    left: 10%;
    right: 10%;
    height: 1px;
    transform: translateY(-50%);
  }

  .guide-v {
    left: 50%;
    top: 10%;
    bottom: 10%;
    width: 1px;
    transform: translateX(-50%);
  }

  .stage-hint {
    margin: 0;
    font-size: 0.75rem;
  }

  .zoom-row {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 200px;
    font-size: 0.8rem;
  }

  .zoom-row input[type='range'] {
    flex: 1;
    accent-color: var(--accent);
  }

  .editor-actions {
    display: flex;
    gap: 8px;
  }
</style>
