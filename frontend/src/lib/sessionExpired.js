/*
 * Ponto de encontro entre api.js e authStore.js para o caso "a sessão caiu no
 * meio do uso". Fica num módulo à parte porque authStore importa api.js — se o
 * api.js importasse o authStore de volta, o ciclo quebraria a carga dos dois.
 */
let handler = null;

export function setUnauthorizedHandler(fn) {
  handler = fn;
}

export function notifyUnauthorized() {
  if (handler) handler();
}
