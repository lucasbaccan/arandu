/*
 * liveConnect é um wrapper leve de EventSource pras telas ao vivo
 * (/palco, /palco/:id/apresentar, /plateia/:id): o SSE é o caminho rápido e
 * barato, mas nas redes corporativas o proxy costuma "silenciar" o stream sem
 * derrubá-lo — o socket vira um half-open que ninguém detecta, e a tela
 * congela no último snapshot. Então além do SSE, este helper re-busca o
 * snapshot periodicamente (poll) QUANDO o stream está mudo: é esse re-fetch
 * (uma request curta, não stream) que atravessa o proxy e garante que a tela
 * sempre converge pro estado do servidor. Enquanto o SSE está entregando, o
 * poll nem dispara (sem custo extra).
 *
 * Status: o SSE envia um snapshot a cada mudança E no heartbeat do servidor
 * (~15s), então receber um data frame periodicamente é prova de conexão viva.
 * Quando o notebook deixa de entregar frames por mais que `offlineAfter`, a
 * tela é avisada via onStatus('offline') — mesmo que o poll continue trazendo
 * conteúdo (é o modo degradado, não a tempo real). Se o SSE voltar, volta
 * onStatus('online').
 */
export function liveConnect({
  url,
  onSnapshot,
  onError,
  onReaction,
  poll,
  staleAfter = 8000,
  pollEvery = 4000,
  offlineAfter = 40000,
  onStatus
}) {
  // lastFrame = último data frame que chegou VIA SSE (base do status offline).
  // lastSync   = última vez que a tela teve dados atualizados (SSE ou poll).
  // Separá-los é o que evita mostrar "sem conexão" durante uma pausa normal
  // da apresentação, quando o stream está vivo mas ocioso: o heartbeat do
  // servidor renova lastFrame mesmo sem mudança de estado.
  let lastFrame = Date.now();
  let lastSync = Date.now();
  let offline = false;

  const setOffline = (value) => {
    if (offline === value) return;
    offline = value;
    if (onStatus) onStatus(value ? 'offline' : 'online');
  };

  const isFrameStale = () => Date.now() - lastFrame >= offlineAfter;
  const isSyncStale = () => Date.now() - lastSync >= staleAfter;

  const es = new EventSource(url);
  es.onmessage = (e) => {
    lastFrame = Date.now();
    lastSync = Date.now();
    setOffline(false);
    onSnapshot(JSON.parse(e.data));
  };
  es.onerror = () => {
    setOffline(true);
    if (onError) onError();
  };
  if (onReaction) {
    es.addEventListener('reaction', (e) => onReaction(JSON.parse(e.data).emoji));
  }

  const timer = setInterval(async () => {
    if (isFrameStale()) setOffline(true);
    if (!isSyncStale()) return;
    try {
      const snapshot = await poll();
      lastSync = Date.now();
      // O poll NÃO marca online: se o stream segue mudo, seguir mostrando
      // "sem conexão" é o aviso honesto de que não se está recebendo em
      // tempo real. A restauração pro modo normal é o SSE voltar a entregar.
      onSnapshot(snapshot);
    } catch {
      // Sem sinal: mantém offline e tenta de novo no próximo tick.
    }
  }, pollEvery);

  return {
    close() {
      es.close();
      clearInterval(timer);
    }
  };
}
