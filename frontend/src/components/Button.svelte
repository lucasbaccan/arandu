<script>
  export let variant = 'primary';
  export let size = 'md';
  export let type = 'button';
  export let disabled = false;
  export let block = false;
</script>

<button
  class="btn btn-{variant} {size !== 'md' ? `btn-${size}` : ''} {block ? 'btn-block' : ''}"
  {type}
  {disabled}
  on:click
>
  <slot />
</button>
