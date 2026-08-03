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

func TestUpsertParticipantGeneratesStableEditToken(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	p1, err := s.UpsertParticipant(ctx, Participant{ID: 510, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}
	if p1.EditToken == "" {
		t.Fatal("edit_token não deveria vir vazio após criação")
	}

	p2, err := s.UpsertParticipant(ctx, Participant{ID: 511, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("reaproveitar participante: %v", err)
	}
	if p2.EditToken != p1.EditToken {
		t.Errorf("edit_token deveria permanecer o mesmo entre reenvios, got %q e %q", p1.EditToken, p2.EditToken)
	}
}

func TestUpsertParticipantPreservesPhotoWhenNotResent(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	if _, err := s.UpsertParticipant(ctx, Participant{ID: 520, EventID: eventID, Email: "ana@x.com", Photo: "foto1"}); err != nil {
		t.Fatalf("criar participante: %v", err)
	}

	p2, err := s.UpsertParticipant(ctx, Participant{ID: 521, EventID: eventID, Email: "ana@x.com", Photo: ""})
	if err != nil {
		t.Fatalf("reenviar sem foto: %v", err)
	}
	if p2.Photo != "foto1" {
		t.Errorf("foto não deveria ser apagada por um reenvio sem foto, got %q", p2.Photo)
	}
}

func TestFindParticipantByEventAndToken(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	p1, err := s.UpsertParticipant(ctx, Participant{ID: 530, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}

	found, err := s.FindParticipantByEventAndToken(ctx, eventID, p1.EditToken)
	if err != nil {
		t.Fatalf("buscar por token: %v", err)
	}
	if found.ID != p1.ID {
		t.Errorf("participante divergente: got %d, esperado %d", found.ID, p1.ID)
	}

	if _, err := s.FindParticipantByEventAndToken(ctx, eventID, "token-invalido"); !errors.Is(err, ErrNotFound) {
		t.Errorf("token inválido: esperado ErrNotFound, got %v", err)
	}
	if _, err := s.FindParticipantByEventAndToken(ctx, eventID, ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("token vazio: esperado ErrNotFound, got %v", err)
	}
}

func TestUpdateParticipantPhoto(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	p1, err := s.UpsertParticipant(ctx, Participant{ID: 540, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}

	if err := s.UpdateParticipantPhoto(ctx, p1.ID, "nova-foto"); err != nil {
		t.Fatalf("atualizar foto: %v", err)
	}
	updated, err := s.FindParticipantByID(ctx, p1.ID)
	if err != nil {
		t.Fatalf("buscar participante: %v", err)
	}
	if updated.Photo != "nova-foto" {
		t.Errorf("foto não atualizada, got %q", updated.Photo)
	}

	if err := s.UpdateParticipantPhoto(ctx, 999999, "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("participante inexistente: esperado ErrNotFound, got %v", err)
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
