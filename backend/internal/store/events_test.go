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
