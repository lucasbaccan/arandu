package store

import (
	"context"
	"errors"
	"testing"
)

func TestCreateAndListLiveQAMessages(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	m1, err := s.CreateLiveQAMessage(ctx, LiveQAMessage{ID: 900, EventID: eventID, Text: "primeira"})
	if err != nil {
		t.Fatalf("criar mensagem 1: %v", err)
	}
	if m1.ID != 900 || m1.Text != "primeira" {
		t.Errorf("mensagem divergente: %+v", m1)
	}

	p, err := s.UpsertParticipant(ctx, Participant{ID: 800, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}
	m2, err := s.CreateLiveQAMessage(ctx, LiveQAMessage{ID: 901, EventID: eventID, ParticipantID: p.ID, Text: "segunda"})
	if err != nil {
		t.Fatalf("criar mensagem 2: %v", err)
	}
	if m2.ParticipantID != p.ID {
		t.Errorf("participantId divergente: %+v", m2)
	}

	list, err := s.ListLiveQAMessagesByEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("listar mensagens: %v", err)
	}
	if len(list) != 2 || list[0].ID != 900 || list[1].ID != 901 {
		t.Errorf("lista divergente (esperado ordem cronológica): %+v", list)
	}
}

func TestListLiveQAMessagesExcludesDismissed(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	if _, err := s.CreateLiveQAMessage(ctx, LiveQAMessage{ID: 910, EventID: eventID, Text: "fica"}); err != nil {
		t.Fatalf("criar mensagem 1: %v", err)
	}
	if _, err := s.CreateLiveQAMessage(ctx, LiveQAMessage{ID: 911, EventID: eventID, Text: "some"}); err != nil {
		t.Fatalf("criar mensagem 2: %v", err)
	}
	if err := s.DismissLiveQAMessage(ctx, 911); err != nil {
		t.Fatalf("dispensar: %v", err)
	}

	list, err := s.ListLiveQAMessagesByEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("listar mensagens: %v", err)
	}
	if len(list) != 1 || list[0].ID != 910 {
		t.Errorf("esperava só a mensagem não dispensada, got %+v", list)
	}
}

func TestDismissLiveQAMessageNotFound(t *testing.T) {
	s, _ := setupQuestionStore(t)
	ctx := context.Background()

	err := s.DismissLiveQAMessage(ctx, 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("esperava ErrNotFound, got %v", err)
	}
}

func TestListLiveQAMessagesScopedToEvent(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	if _, err := s.CreateEvent(ctx, Event{ID: 20, OwnerID: 1, Title: "Outro evento", PINCode: "654321", Status: "PREPARATION"}); err != nil {
		t.Fatalf("criar outro evento: %v", err)
	}

	if _, err := s.CreateLiveQAMessage(ctx, LiveQAMessage{ID: 920, EventID: eventID, Text: "evento 1"}); err != nil {
		t.Fatalf("criar mensagem evento 1: %v", err)
	}
	if _, err := s.CreateLiveQAMessage(ctx, LiveQAMessage{ID: 921, EventID: 20, Text: "evento 2"}); err != nil {
		t.Fatalf("criar mensagem evento 2: %v", err)
	}

	list, err := s.ListLiveQAMessagesByEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("listar mensagens: %v", err)
	}
	if len(list) != 1 || list[0].ID != 920 {
		t.Errorf("esperava só a mensagem do evento 1, got %+v", list)
	}
}
