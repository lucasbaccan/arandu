import { writable } from 'svelte/store';

// Diferente do toastStore (slot único, substitui o item anterior), reações
// podem chegar em rajada e cada uma precisa aparecer na tela — por isso é
// uma lista, não um valor único.
export const reactions = writable([]);

let counter = 0;

export function fireReaction(emoji) {
  const id = ++counter;
  reactions.update((list) => [...list, { id, emoji }]);
  setTimeout(() => {
    reactions.update((list) => list.filter((r) => r.id !== id));
  }, 2000);
}
