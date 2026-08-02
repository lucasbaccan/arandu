<script>
  export let variant = 'primary';
  export let type = 'button';
  export let disabled = false;
  export let block = false;
</script>

<button
  class="btn btn-{variant} {block ? 'btn-block' : ''}"
  {type}
  {disabled}
  on:click
>
  <slot />
</button>
