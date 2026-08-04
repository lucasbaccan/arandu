package live

import (
	"sync"
	"testing"
	"time"
)

func TestRevealMarksParticipant(t *testing.T) {
	m := NewManager()
	m.Reveal(1, 10, 100)
	m.Reveal(1, 10, 101)

	state := m.Get(1)
	if !state.Revealed[10][100] || !state.Revealed[10][101] {
		t.Fatalf("esperava 100 e 101 revelados na pergunta 10, got %+v", state.Revealed)
	}
}

func TestRevealAllMarksEveryone(t *testing.T) {
	m := NewManager()
	m.RevealAll(1, 10, []int64{100, 101, 102})

	state := m.Get(1)
	for _, id := range []int64{100, 101, 102} {
		if !state.Revealed[10][id] {
			t.Fatalf("esperava %d revelado, got %+v", id, state.Revealed)
		}
	}
}

func TestResetClearsOnlyThatQuestion(t *testing.T) {
	m := NewManager()
	m.Reveal(1, 10, 100)
	m.Reveal(1, 20, 200)

	m.Reset(1, 10)

	state := m.Get(1)
	if len(state.Revealed[10]) != 0 {
		t.Fatalf("pergunta 10 deveria estar limpa, got %+v", state.Revealed[10])
	}
	if !state.Revealed[20][200] {
		t.Fatal("pergunta 20 nao deveria ter sido afetada pelo reset da pergunta 10")
	}
}

func TestSetCurrentQuestion(t *testing.T) {
	m := NewManager()
	m.SetCurrentQuestion(1, 10)
	m.SetCurrentQuestion(1, 20)

	if got := m.Get(1).CurrentQuestionID; got != 20 {
		t.Fatalf("esperava pergunta atual 20, got %d", got)
	}
}

func TestGetReturnsIndependentCopy(t *testing.T) {
	m := NewManager()
	m.Reveal(1, 10, 100)

	state := m.Get(1)
	state.Revealed[10][999] = true // mutar a cópia não pode afetar o Manager

	fresh := m.Get(1)
	if fresh.Revealed[10][999] {
		t.Fatal("Get deveria retornar uma cópia independente do estado interno")
	}
}

func TestSubscribeNotifiesOnChange(t *testing.T) {
	m := NewManager()
	ch, unsubscribe := m.Subscribe(1)
	defer unsubscribe()

	m.Reveal(1, 10, 100)

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("esperava receber um sinal apos Reveal")
	}
}

func TestSubscribeCoalescesRapidChanges(t *testing.T) {
	m := NewManager()
	ch, unsubscribe := m.Subscribe(1)
	defer unsubscribe()

	for i := int64(0); i < 5; i++ {
		m.Reveal(1, 10, 100+i)
	}

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("esperava pelo menos um sinal apos as mudancas")
	}
	// canal tem buffer 1: mudancas rapidas nao devem travar o Reveal nem
	// exigir que o assinante drene 5 sinais.
	select {
	case <-ch:
		t.Fatal("nao esperava um segundo sinal pendente (deveria coalescer)")
	default:
	}
}

func TestUnsubscribeStopsFurtherSignals(t *testing.T) {
	m := NewManager()
	ch, unsubscribe := m.Subscribe(1)
	unsubscribe()

	m.Reveal(1, 10, 100)

	select {
	case <-ch:
		t.Fatal("nao deveria receber sinal apos unsubscribe")
	case <-time.After(50 * time.Millisecond):
	}
}

// TestConcurrentAccess roda com -race pra garantir que reveals concorrentes
// de varias goroutines (varios cliques quase simultaneos, por ex.) e um
// assinante lendo o estado ao mesmo tempo nao disparam data race.
func TestConcurrentAccess(t *testing.T) {
	m := NewManager()
	var wg sync.WaitGroup

	for i := int64(0); i < 50; i++ {
		wg.Add(1)
		go func(participantID int64) {
			defer wg.Done()
			m.Reveal(1, 10, participantID)
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			m.Get(1)
		}
	}()

	ch, unsubscribe := m.Subscribe(1)
	defer unsubscribe()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			select {
			case <-ch:
			default:
			}
		}
	}()

	wg.Wait()

	state := m.Get(1)
	if len(state.Revealed[10]) != 50 {
		t.Fatalf("esperava 50 participantes revelados, got %d", len(state.Revealed[10]))
	}
}
