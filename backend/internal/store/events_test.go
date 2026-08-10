package store

import (
	"context"
	"errors"
	"testing"
)

func createOwner(t *testing.T, s *Store, id int64, email string) {
	t.Helper()
	if _, err := s.CreateUser(context.Background(), User{
		ID: id, Email: email, Name: "Dono", PasswordHash: "h", AuthProvider: "email",
	}); err != nil {
		t.Fatalf("criar dono: %v", err)
	}
}

func TestCreateAndListEvents(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1001, "dono1@x.com")
	createOwner(t, s, 1002, "dono2@x.com")

	ev, err := s.CreateEvent(ctx, Event{
		ID: 2001, OwnerID: 1001, Title: "Conecta DevOps", PINCode: "123456", Status: "PREPARATION",
	})
	if err != nil {
		t.Fatalf("criar evento: %v", err)
	}
	if ev.PINCode != "123456" || ev.Status != "PREPARATION" || ev.ShowRanking {
		t.Errorf("evento divergente: %+v", ev)
	}

	s.CreateEvent(ctx, Event{ID: 2002, OwnerID: 1001, Title: "Outro", PINCode: "654321", Status: "PREPARATION"})
	s.CreateEvent(ctx, Event{ID: 2003, OwnerID: 1002, Title: "De outro dono", PINCode: "111222", Status: "PREPARATION"})

	events, err := s.ListEventsByOwner(ctx, 1001)
	if err != nil {
		t.Fatalf("listar eventos: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("esperado 2 eventos do dono 1001, got %d", len(events))
	}
	for _, e := range events {
		if e.OwnerID != 1001 {
			t.Errorf("evento de outro dono listado: %+v", e)
		}
	}
}

func TestCreateEventPinTaken(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1, "a@x.com")
	createOwner(t, s, 2, "b@x.com")

	_, err := s.CreateEvent(ctx, Event{ID: 1, OwnerID: 1, Title: "A", PINCode: "abc", Status: "PREPARATION"})
	if err != nil {
		t.Fatalf("criar primeiro evento: %v", err)
	}
	_, err = s.CreateEvent(ctx, Event{ID: 2, OwnerID: 2, Title: "B", PINCode: "abc", Status: "PREPARATION"})
	if !errors.Is(err, ErrPinTaken) {
		t.Errorf("esperado ErrPinTaken, got %v", err)
	}
}

func TestCreateEventPinReusableAfterFinish(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1, "a@x.com")
	createOwner(t, s, 2, "b@x.com")

	_, err := s.CreateEvent(ctx, Event{ID: 1, OwnerID: 1, Title: "A", PINCode: "abc", Status: "FINISHED"})
	if err != nil {
		t.Fatalf("criar primeiro evento: %v", err)
	}
	_, err = s.CreateEvent(ctx, Event{ID: 2, OwnerID: 2, Title: "B", PINCode: "abc", Status: "PREPARATION"})
	if err != nil {
		t.Errorf("PIN de evento FINISHED deveria ser reciclável, got %v", err)
	}
}

func TestFindEventByIDAndOwner(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1, "a@x.com")
	createOwner(t, s, 2, "b@x.com")

	s.CreateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "Meu", PINCode: "111", Status: "PREPARATION"})

	ev, err := s.FindEventByIDAndOwner(ctx, 10, 1)
	if err != nil {
		t.Fatalf("buscar evento do dono: %v", err)
	}
	if ev.Title != "Meu" {
		t.Errorf("título esperado Meu, got %s", ev.Title)
	}

	if _, err := s.FindEventByIDAndOwner(ctx, 10, 2); !errors.Is(err, ErrNotFound) {
		t.Errorf("outro dono deveria ser ErrNotFound, got %v", err)
	}
	if _, err := s.FindEventByIDAndOwner(ctx, 999, 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("inexistente deveria ser ErrNotFound, got %v", err)
	}
}

func TestUpdateEvent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1, "a@x.com")

	_, err := s.CreateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "Antes", PINCode: "111", Status: "PREPARATION"})
	if err != nil {
		t.Fatalf("criar evento: %v", err)
	}

	ev, err := s.UpdateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "Depois", PINCode: "222", Status: "PREPARATION", ShowRanking: true})
	if err != nil {
		t.Fatalf("atualizar evento: %v", err)
	}
	if ev.Title != "Depois" || ev.PINCode != "222" || !ev.ShowRanking {
		t.Errorf("evento atualizado divergente: %+v", ev)
	}

	got, err := s.FindEventByIDAndOwner(ctx, 10, 1)
	if err != nil {
		t.Fatalf("buscar evento atualizado: %v", err)
	}
	if got.Title != "Depois" || got.PINCode != "222" || !got.ShowRanking {
		t.Errorf("alterações não persistidas: %+v", got)
	}
}

func TestSetInteractionsEnabled(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1, "a@x.com")
	if _, err := s.CreateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "Evento", PINCode: "111", Status: "PREPARATION", InteractionsEnabled: true}); err != nil {
		t.Fatalf("criar evento: %v", err)
	}

	if err := s.SetInteractionsEnabled(ctx, 10, false); err != nil {
		t.Fatalf("desligar interações: %v", err)
	}
	got, err := s.FindEventByID(ctx, 10)
	if err != nil {
		t.Fatalf("buscar evento: %v", err)
	}
	if got.InteractionsEnabled {
		t.Errorf("esperava interações desligadas, got %+v", got)
	}

	if err := s.SetInteractionsEnabled(ctx, 10, true); err != nil {
		t.Fatalf("ligar interações: %v", err)
	}
	got, err = s.FindEventByID(ctx, 10)
	if err != nil {
		t.Fatalf("buscar evento: %v", err)
	}
	if !got.InteractionsEnabled {
		t.Errorf("esperava interações ligadas, got %+v", got)
	}
}

func TestUpdateEventDoesNotTouchInteractionsEnabled(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1, "a@x.com")
	if _, err := s.CreateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "Evento", PINCode: "111", Status: "PREPARATION", InteractionsEnabled: true}); err != nil {
		t.Fatalf("criar evento: %v", err)
	}
	if err := s.SetInteractionsEnabled(ctx, 10, false); err != nil {
		t.Fatalf("desligar interações: %v", err)
	}

	if _, err := s.UpdateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "Depois", PINCode: "222", Status: "PREPARATION"}); err != nil {
		t.Fatalf("atualizar evento: %v", err)
	}

	got, err := s.FindEventByID(ctx, 10)
	if err != nil {
		t.Fatalf("buscar evento: %v", err)
	}
	if got.InteractionsEnabled {
		t.Errorf("UpdateEvent não deveria reverter interactions_enabled, got %+v", got)
	}
}

func TestUpdateEventNotFoundAndPinTaken(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1, "a@x.com")
	createOwner(t, s, 2, "b@x.com")

	s.CreateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "A", PINCode: "111", Status: "PREPARATION"})
	s.CreateEvent(ctx, Event{ID: 20, OwnerID: 2, Title: "B", PINCode: "222", Status: "PREPARATION"})

	_, err := s.UpdateEvent(ctx, Event{ID: 10, OwnerID: 2, Title: "X", PINCode: "333", Status: "PREPARATION"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("outro dono deveria ser ErrNotFound, got %v", err)
	}

	_, err = s.UpdateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "A", PINCode: "222", Status: "PREPARATION"})
	if !errors.Is(err, ErrPinTaken) {
		t.Errorf("PIN duplicado deveria ser ErrPinTaken, got %v", err)
	}
}
