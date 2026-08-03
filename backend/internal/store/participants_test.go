package store

import (
	"context"
	"errors"
	"testing"
)

func TestUpsertParticipantCreatesAndReuses(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	p1, err := s.UpsertParticipant(ctx, Participant{ID: 500, EventID: eventID, Email: "ana@x.com", Photo: "foto1"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}
	if p1.ID != 500 || p1.Photo != "foto1" {
		t.Errorf("participante divergente: %+v", p1)
	}

	p2, err := s.UpsertParticipant(ctx, Participant{ID: 501, EventID: eventID, Email: "ana@x.com", Photo: "foto2"})
	if err != nil {
		t.Fatalf("reaproveitar participante: %v", err)
	}
	if p2.ID != p1.ID {
		t.Errorf("deveria reaproveitar o mesmo participante, got id %d, esperado %d", p2.ID, p1.ID)
	}
	if p2.Photo != "foto2" {
		t.Errorf("foto deveria ser atualizada, got %q", p2.Photo)
	}
}

func TestFindParticipantByEventAndEmailNotFound(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	if _, err := s.FindParticipantByEventAndEmail(context.Background(), eventID, "nao-existe@x.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("esperado ErrNotFound, got %v", err)
	}
}

func TestUpsertParticipantScopedPerEvent(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	if _, err := s.CreateEvent(ctx, Event{ID: 11, OwnerID: 1, Title: "Outro evento", PINCode: "654321", Status: "PREPARATION"}); err != nil {
		t.Fatalf("criar segundo evento: %v", err)
	}

	p1, err := s.UpsertParticipant(ctx, Participant{ID: 600, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante evento 1: %v", err)
	}
	p2, err := s.UpsertParticipant(ctx, Participant{ID: 601, EventID: 11, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante evento 2: %v", err)
	}
	if p1.ID == p2.ID {
		t.Errorf("participantes de eventos diferentes não deveriam compartilhar id")
	}
}
