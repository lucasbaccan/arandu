/*
 * Situação do evento numa fonte única: rótulo + cor da paleta da logo (ciano
 * coletando, laranja em preparação, azul-esverdeado ao vivo, cinza encerrado).
 * Dashboard, Editar evento e Palco leem daqui — antes cada tela tinha o seu.
 */
const META = {
  PREPARATION: { label: 'Em preparação', tint: 'var(--tint-orange)', color: 'var(--orange-text)' },
  OPEN_FOR_ANSWERS: { label: 'Coletando respostas', tint: 'var(--tint-cyan)', color: 'var(--cyan-text)' },
  CLOSED_FOR_ANSWERS: { label: 'Respostas encerradas', tint: 'var(--tint-neutral)', color: 'var(--text-muted)' },
  PRESENTING: { label: 'Ao vivo', tint: 'var(--tint-success)', color: 'var(--success-text)' },
  FINISHED: { label: 'Finalizado', tint: 'var(--tint-neutral)', color: 'var(--text-muted)' }
};

export function statusInfo(status) {
  return META[status] || { label: status, tint: 'var(--tint-neutral)', color: 'var(--text-muted)' };
}

export function statusLabel(status) {
  return statusInfo(status).label;
}

/* Tipo de pergunta usa o mesmo chip, em outra variante de cor. */
export function questionKindInfo(type) {
  if (type === 'OPEN_TEXT') {
    return { label: 'Resposta aberta', tint: 'var(--tint-orange)', color: 'var(--orange-text)' };
  }
  if (type === 'GROUP') {
    return { label: 'Múltipla escolha', tint: 'var(--tint-cyan)', color: 'var(--cyan-text)' };
  }
  return { label: 'Individual', tint: 'var(--tint-purple)', color: 'var(--purple-text)' };
}
