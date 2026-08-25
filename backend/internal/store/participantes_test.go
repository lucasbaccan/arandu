package store

import (
	"context"
	"errors"
	"testing"
)

func TestUpsertParticipantCreatesAndReuses(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	p1, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 500, EventID: eventID, Email: "ana@x.com", Photo: "foto1"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}
	if p1.ID != 500 || p1.Photo != "foto1" {
		t.Errorf("participante divergente: %+v", p1)
	}

	p2, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 501, EventID: eventID, Email: "ana@x.com", Photo: "foto2"})
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

	p1, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 510, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}
	if p1.EditToken == "" {
		t.Fatal("edit_token não deveria vir vazio após criação")
	}

	p2, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 511, EventID: eventID, Email: "ana@x.com"})
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

	if _, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 520, EventID: eventID, Email: "ana@x.com", Photo: "foto1"}); err != nil {
		t.Fatalf("criar participante: %v", err)
	}

	p2, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 521, EventID: eventID, Email: "ana@x.com", Photo: ""})
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

	p1, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 530, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}

	found, err := s.BuscarParticipantePorEventoEToken(ctx, eventID, p1.EditToken)
	if err != nil {
		t.Fatalf("buscar por token: %v", err)
	}
	if found.ID != p1.ID {
		t.Errorf("participante divergente: got %d, esperado %d", found.ID, p1.ID)
	}

	if _, err := s.BuscarParticipantePorEventoEToken(ctx, eventID, "token-invalido"); !errors.Is(err, ErrNotFound) {
		t.Errorf("token inválido: esperado ErrNotFound, got %v", err)
	}
	if _, err := s.BuscarParticipantePorEventoEToken(ctx, eventID, ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("token vazio: esperado ErrNotFound, got %v", err)
	}
}

func TestUpdateParticipantPhoto(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	p1, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 540, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}

	if err := s.AtualizarFotoDoParticipante(ctx, p1.ID, "nova-foto"); err != nil {
		t.Fatalf("atualizar foto: %v", err)
	}
	updated, err := s.BuscarParticipantePorID(ctx, p1.ID)
	if err != nil {
		t.Fatalf("buscar participante: %v", err)
	}
	if updated.Photo != "nova-foto" {
		t.Errorf("foto não atualizada, got %q", updated.Photo)
	}

	if err := s.AtualizarFotoDoParticipante(ctx, 999999, "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("participante inexistente: esperado ErrNotFound, got %v", err)
	}
}

func TestFindParticipantByEventAndEmailNotFound(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	if _, err := s.BuscarParticipantePorEventoEEmail(context.Background(), eventID, "nao-existe@x.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("esperado ErrNotFound, got %v", err)
	}
}

func TestUpsertParticipantScopedPerEvent(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	if _, err := s.CriarEvento(ctx, Event{ID: 11, OwnerID: 1, Title: "Outro evento", PINCode: "654321", Status: "PREPARATION"}); err != nil {
		t.Fatalf("criar segundo evento: %v", err)
	}

	p1, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 600, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante evento 1: %v", err)
	}
	p2, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 601, EventID: 11, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante evento 2: %v", err)
	}
	if p1.ID == p2.ID {
		t.Errorf("participantes de eventos diferentes não deveriam compartilhar id")
	}
}

// insertParticipantWithoutToken simula uma linha criada antes do link de
// edição existir (edit_token vazio), contornando o InserirOuAtualizarParticipante normal.
func insertParticipantWithoutToken(t *testing.T, s *Store, id, eventID int64, email string) {
	t.Helper()
	_, err := s.db.Exec(
		`INSERT INTO participants (id, event_id, email, photo, edit_token, created_at) VALUES (?, ?, ?, '', '', datetime('now'))`,
		id, eventID, email,
	)
	if err != nil {
		t.Fatalf("inserir participante legado sem token: %v", err)
	}
}

func TestUpsertParticipantBackfillsMissingEditToken(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()
	insertParticipantWithoutToken(t, s, 610, eventID, "legado@x.com")

	p, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 611, EventID: eventID, Email: "legado@x.com"})
	if err != nil {
		t.Fatalf("reenviar participante legado: %v", err)
	}
	if p.ID != 610 {
		t.Errorf("deveria reaproveitar o participante legado, got id %d", p.ID)
	}
	if p.EditToken == "" {
		t.Fatal("edit_token deveria ser preenchido ao reenviar um participante legado sem token")
	}

	found, err := s.BuscarParticipantePorEventoEToken(ctx, eventID, p.EditToken)
	if err != nil {
		t.Fatalf("buscar pelo token recém-gerado: %v", err)
	}
	if found.ID != 610 {
		t.Errorf("token gerado não aponta para o participante legado, got id %d", found.ID)
	}
}

func TestListParticipantsByEventBackfillsMissingEditToken(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()
	insertParticipantWithoutToken(t, s, 620, eventID, "legado2@x.com")

	list, err := s.ListarParticipantesPorEvento(ctx, eventID)
	if err != nil {
		t.Fatalf("listar participantes: %v", err)
	}
	if len(list) != 1 || list[0].EditToken == "" {
		t.Fatalf("edit_token deveria ser preenchido ao listar, got %+v", list)
	}

	found, err := s.BuscarParticipantePorEventoEToken(ctx, eventID, list[0].EditToken)
	if err != nil {
		t.Fatalf("buscar pelo token recém-gerado: %v", err)
	}
	if found.ID != 620 {
		t.Errorf("token gerado não aponta para o participante legado, got id %d", found.ID)
	}
}
