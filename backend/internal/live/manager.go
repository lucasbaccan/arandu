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
}

func newEventState() *EventState {
	return &EventState{Revealed: make(map[int64]map[int64]bool)}
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
	return EventState{CurrentQuestionID: s.CurrentQuestionID, Revealed: revealed}
}

type eventEntry struct {
	mu    sync.Mutex
	state *EventState
	subs  map[chan struct{}]struct{}
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
}

func NewManager() *Manager {
	return &Manager{events: make(map[int64]*eventEntry)}
}

func (m *Manager) entry(eventID int64) *eventEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.events[eventID]
	if !ok {
		e = &eventEntry{state: newEventState(), subs: make(map[chan struct{}]struct{})}
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
