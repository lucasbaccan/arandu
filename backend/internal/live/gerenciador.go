// Package live guarda, em memória, o estado da apresentação ao vivo de cada
// evento (qual pergunta está em foco e quem já foi revelado) e notifica
// assinantes (as conexões SSE da tela pública) quando esse estado muda.
//
// Não persiste em disco: se o servidor reiniciar no meio de uma
// apresentação, o estado zera. É uma troca deliberada — ver o plano da
// feature — já que é um único processo e o estado é efêmero por natureza.
package live

import "sync"

// EventState é um retrato do estado ao vivo de um evento.
type EventState struct {
	CurrentQuestionID int64
	// Revealed mapeia questionID -> conjunto de participantID já revelados
	// nessa pergunta.
	Revealed map[int64]map[int64]bool
	// Blanked, quando true, pede pra tela dos participantes esconder a
	// pergunta atual (ex: intervalo).
	Blanked bool
	// Message é um aviso/recado que o organizador transmite pra tela dos
	// participantes, sobrepondo a pergunta atual enquanto não-vazio.
	Message string
	// AnswersHidden, quando true, esconde só as zonas de resposta (opções e
	// quem já foi revelado nelas) da tela pública — a pergunta e a fila de
	// pendentes continuam visíveis. Diferente de Blanked, que some com tudo.
	AnswersHidden bool
	// NamesHidden, quando true, esconde a legenda de nome sob cada rosto na
	// janela de apresentação e na tela da plateia — o painel do próprio
	// organizador (/stage) sempre mostra os nomes, independente disso.
	NamesHidden bool
	// PresentDensityMode escolhe como o placar de respostas (fitDensity, em
	// PresentationStage.svelte) decide colunas × escala das pílulas:
	// "modo_amplo" (valor padrão de um evento novo) = Amplo, testa todas as
	// colunas e fica com a maior escala, "modo_compacto" = Compacto (o
	// automático antigo, menor nº de colunas que cabe), "1".."4" = força esse
	// nº de colunas. Escolhido pelos botões no rodapé de /stage e refletido em
	// tempo real tanto na janela de apresentação (/stage/{id}/present) quanto
	// na tela pública (/audience/{id}) — as duas usam o mesmo snapshot
	// (buildLiveSnapshot). "" pode chegar do banco (evento nunca configurado)
	// e os clientes o interpretam como o padrão (modo_amplo) — ver o fallback
	// em Palco.svelte e PalcoApresentar.svelte.
	PresentDensityMode string
}

func newEventState() *EventState {
	return &EventState{
		Revealed:           make(map[int64]map[int64]bool),
		PresentDensityMode: "modo_amplo",
	}
}

func (s *EventState) clone() EventState {
	revealed := make(map[int64]map[int64]bool, len(s.Revealed))
	for questionID, set := range s.Revealed {
		copySet := make(map[int64]bool, len(set))
		for participantID := range set {
			copySet[participantID] = true
		}
		revealed[questionID] = copySet
	}
	return EventState{
		CurrentQuestionID:  s.CurrentQuestionID,
		Revealed:           revealed,
		Blanked:            s.Blanked,
		Message:            s.Message,
		AnswersHidden:      s.AnswersHidden,
		NamesHidden:        s.NamesHidden,
		PresentDensityMode: s.PresentDensityMode,
	}
}

// ReactionEvent é uma reação de emoji individual, entregue a quem estiver
// assinando via SubscribeReactions. Diferente do sinal coalescente de
// notify() (que só avisa "algo mudou, busque o snapshot de novo"), cada
// reação precisa chegar sozinha aos assinantes: 10 corações mandados devem
// aparecer como 10 corações, não virar um único refetch.
type ReactionEvent struct {
	Emoji string
}

// reactionBufferSize limita quantas reações ficam em fila por assinante.
// Perder reação em buffer cheio é aceitável — é um recurso cosmético, sem
// requisito de correção (ao contrário do resto deste pacote).
const reactionBufferSize = 32

type eventEntry struct {
	mu           sync.Mutex
	state        *EventState
	subs         map[chan struct{}]struct{}
	reactionSubs map[chan ReactionEvent]struct{}
}

func (e *eventEntry) notify() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for ch := range e.subs {
		select {
		case ch <- struct{}{}:
		default:
			// assinante já tem um aviso pendente; ele vai buscar o snapshot
			// mais recente quando processar, não precisa empilhar sinais.
		}
	}
}

// Manager guarda o EventState de cada evento, protegido por um mutex por
// evento (mesmo padrão do ids.Generator).
type Manager struct {
	mu     sync.Mutex
	events map[int64]*eventEntry
	// loader, se definido, busca o estado persistido (banco) na primeira vez
	// que um evento é acessado neste processo — é o que permite retomar
	// revelação/pergunta atual depois de um restart do servidor, em vez de
	// sempre começar com newEventState() em branco. Ver SetLoader.
	loader func(eventID int64) EventState
}

func NewManager() *Manager {
	return &Manager{events: make(map[int64]*eventEntry)}
}

// SetLoader liga a persistência: chamada uma vez, normalmente logo após
// NewManager, com uma função que busca o estado salvo no banco. Fica fora do
// construtor de propósito — mantém este pacote sem depender de internal/store
// (quem liga os dois é a camada de API) e os testes deste pacote continuam
// funcionando sem loader (estado em branco, como sempre foi).
func (m *Manager) SetLoader(loader func(eventID int64) EventState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loader = loader
}

func (m *Manager) entry(eventID int64) *eventEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.events[eventID]
	if !ok {
		state := newEventState()
		if m.loader != nil {
			loaded := m.loader(eventID)
			if loaded.Revealed == nil {
				loaded.Revealed = make(map[int64]map[int64]bool)
			}
			state = &loaded
		}
		e = &eventEntry{
			state:        state,
			subs:         make(map[chan struct{}]struct{}),
			reactionSubs: make(map[chan ReactionEvent]struct{}),
		}
		m.events[eventID] = e
	}
	return e
}

// Get retorna uma cópia do estado atual do evento (mapas independentes, seguro
// pra ler sem lock externo).
func (m *Manager) Get(eventID int64) EventState {
	e := m.entry(eventID)
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state.clone()
}

func (m *Manager) SetCurrentQuestion(eventID, questionID int64) {
	e := m.entry(eventID)
	e.mu.Lock()
	e.state.CurrentQuestionID = questionID
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) Reveal(eventID, questionID, participantID int64) {
	e := m.entry(eventID)
	e.mu.Lock()
	set, ok := e.state.Revealed[questionID]
	if !ok {
		set = make(map[int64]bool)
		e.state.Revealed[questionID] = set
	}
	set[participantID] = true
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) Unreveal(eventID, questionID, participantID int64) {
	e := m.entry(eventID)
	e.mu.Lock()
	if set, ok := e.state.Revealed[questionID]; ok {
		delete(set, participantID)
	}
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) RevealAll(eventID, questionID int64, participantIDs []int64) {
	e := m.entry(eventID)
	e.mu.Lock()
	set, ok := e.state.Revealed[questionID]
	if !ok {
		set = make(map[int64]bool)
		e.state.Revealed[questionID] = set
	}
	for _, participantID := range participantIDs {
		set[participantID] = true
	}
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) Reset(eventID, questionID int64) {
	e := m.entry(eventID)
	e.mu.Lock()
	delete(e.state.Revealed, questionID)
	e.mu.Unlock()
	e.notify()
}

// ResetAll limpa a revelação de todas as perguntas do evento de uma vez —
// usado pelo botão "Reiniciar tudo" em /stage, pra recomeçar a apresentação
// do zero sem precisar zerar pergunta por pergunta.
func (m *Manager) ResetAll(eventID int64) {
	e := m.entry(eventID)
	e.mu.Lock()
	e.state.Revealed = make(map[int64]map[int64]bool)
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) SetBlanked(eventID int64, blanked bool) {
	e := m.entry(eventID)
	e.mu.Lock()
	e.state.Blanked = blanked
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) SetAnswersHidden(eventID int64, hidden bool) {
	e := m.entry(eventID)
	e.mu.Lock()
	e.state.AnswersHidden = hidden
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) SetNamesHidden(eventID int64, hidden bool) {
	e := m.entry(eventID)
	e.mu.Lock()
	e.state.NamesHidden = hidden
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) SetPresentDensityMode(eventID int64, mode string) {
	e := m.entry(eventID)
	e.mu.Lock()
	e.state.PresentDensityMode = mode
	e.mu.Unlock()
	e.notify()
}

func (m *Manager) SetMessage(eventID int64, message string) {
	e := m.entry(eventID)
	e.mu.Lock()
	e.state.Message = message
	e.mu.Unlock()
	e.notify()
}

// Touch dispara o sinal de "algo mudou" sem alterar o EventState — usado
// quando o que mudou vive fora deste pacote (ex: interactions_enabled no
// banco, ou uma mensagem de Q&A persistida), pra avisar quem está assinando
// via Subscribe que vale a pena buscar um snapshot novo.
func (m *Manager) Touch(eventID int64) {
	e := m.entry(eventID)
	e.notify()
}

// BroadcastReaction entrega uma reação de emoji a todo mundo assinando via
// SubscribeReactions nesse evento. Envio não-bloqueante: assinante lento ou
// com buffer cheio simplesmente perde a reação, sem travar quem mandou.
func (m *Manager) BroadcastReaction(eventID int64, emoji string) {
	e := m.entry(eventID)
	e.mu.Lock()
	defer e.mu.Unlock()
	for ch := range e.reactionSubs {
		select {
		case ch <- ReactionEvent{Emoji: emoji}:
		default:
		}
	}
}

// SubscribeReactions registra um canal que recebe cada reação individual
// (não-coalescente, ao contrário de Subscribe). Chamar unsubscribe ao
// desconectar.
func (m *Manager) SubscribeReactions(eventID int64) (ch chan ReactionEvent, unsubscribe func()) {
	e := m.entry(eventID)
	ch = make(chan ReactionEvent, reactionBufferSize)
	e.mu.Lock()
	e.reactionSubs[ch] = struct{}{}
	e.mu.Unlock()
	unsubscribe = func() {
		e.mu.Lock()
		delete(e.reactionSubs, ch)
		e.mu.Unlock()
	}
	return ch, unsubscribe
}

// Subscribe registra um canal de aviso (buffer 1, envio não-bloqueante):
// cada mudança no estado do evento manda um sinal, e quem assina deve buscar
// o snapshot atualizado via Get. Chamar unsubscribe ao desconectar.
func (m *Manager) Subscribe(eventID int64) (ch chan struct{}, unsubscribe func()) {
	e := m.entry(eventID)
	ch = make(chan struct{}, 1)
	e.mu.Lock()
	e.subs[ch] = struct{}{}
	e.mu.Unlock()
	unsubscribe = func() {
		e.mu.Lock()
		delete(e.subs, ch)
		e.mu.Unlock()
	}
	return ch, unsubscribe
}
