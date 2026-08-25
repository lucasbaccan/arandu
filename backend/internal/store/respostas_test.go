package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func setupAnswerStore(t *testing.T) (s *Store, groupQID, openQID, optA, optB, participantID int64) {
	t.Helper()
	var eventID int64
	s, eventID = setupQuestionStore(t)
	ctx := context.Background()

	q, err := s.CriarPergunta(ctx, Question{ID: 300, EventID: eventID, Title: "Escolha", Type: "GROUP", LayoutView: "TIMELINE"}, []QuestionOption{
		{ID: 301, TextLabel: "A"},
		{ID: 302, TextLabel: "B"},
	})
	if err != nil {
		t.Fatalf("criar pergunta de grupo: %v", err)
	}
	oq, err := s.CriarPergunta(ctx, Question{ID: 310, EventID: eventID, Title: "Aberta", Type: "OPEN_TEXT", LayoutView: "TIMELINE"}, nil)
	if err != nil {
		t.Fatalf("criar pergunta aberta: %v", err)
	}

	p, err := s.InserirOuAtualizarParticipante(ctx, Participant{ID: 700, EventID: eventID, Email: "ana@x.com"})
	if err != nil {
		t.Fatalf("criar participante: %v", err)
	}

	return s, q.ID, oq.ID, q.Options[0].ID, q.Options[1].ID, p.ID
}

func TestReplaceAnswersCreatesAndUpdates(t *testing.T) {
	s, groupQID, openQID, optA, optB, participantID := setupAnswerStore(t)
	ctx := context.Background()

	if err := s.SubstituirRespostas(ctx, participantID, []Answer{
		{ID: 1, QuestionID: groupQID, OptionID: optA},
		{ID: 2, QuestionID: openQID, FreeText: "Minha resposta"},
	}); err != nil {
		t.Fatalf("gravar respostas: %v", err)
	}

	// reenviar com resposta diferente para a pergunta de grupo deve substituir, não duplicar
	if err := s.SubstituirRespostas(ctx, participantID, []Answer{
		{ID: 3, QuestionID: groupQID, OptionID: optB},
	}); err != nil {
		t.Fatalf("substituir resposta: %v", err)
	}

	var count int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM answers WHERE participant_id = ? AND question_id = ?`, participantID, groupQID,
	).Scan(&count); err != nil {
		t.Fatalf("contar respostas: %v", err)
	}
	if count != 1 {
		t.Errorf("esperado 1 resposta após substituição, got %d", count)
	}

	var optionID int64
	if err := s.db.QueryRow(
		`SELECT option_id FROM answers WHERE participant_id = ? AND question_id = ?`, participantID, groupQID,
	).Scan(&optionID); err != nil {
		t.Fatalf("ler option_id: %v", err)
	}
	if optionID != optB {
		t.Errorf("option_id esperado %d, got %d", optB, optionID)
	}
}

func TestReplaceAnswersFreeTextHasNullOption(t *testing.T) {
	s, _, openQID, _, _, participantID := setupAnswerStore(t)
	ctx := context.Background()

	if err := s.SubstituirRespostas(ctx, participantID, []Answer{
		{ID: 1, QuestionID: openQID, FreeText: "Texto livre"},
	}); err != nil {
		t.Fatalf("gravar resposta aberta: %v", err)
	}

	var optionID sql.NullInt64
	var freeText string
	if err := s.db.QueryRow(
		`SELECT option_id, free_text FROM answers WHERE participant_id = ? AND question_id = ?`, participantID, openQID,
	).Scan(&optionID, &freeText); err != nil {
		t.Fatalf("ler resposta: %v", err)
	}
	if optionID.Valid {
		t.Errorf("option_id deveria ser NULL para resposta aberta, got %v", optionID)
	}
	if freeText != "Texto livre" {
		t.Errorf("free_text esperado 'Texto livre', got %q", freeText)
	}
}

func TestListAnswersByParticipant(t *testing.T) {
	s, groupQID, openQID, optA, _, participantID := setupAnswerStore(t)
	ctx := context.Background()

	if err := s.SubstituirRespostas(ctx, participantID, []Answer{
		{ID: 1, QuestionID: groupQID, OptionID: optA},
		{ID: 2, QuestionID: openQID, FreeText: "Pizza"},
	}); err != nil {
		t.Fatalf("gravar respostas: %v", err)
	}

	answers, err := s.ListarRespostasPorParticipante(ctx, participantID)
	if err != nil {
		t.Fatalf("listar respostas: %v", err)
	}
	if len(answers) != 2 {
		t.Fatalf("esperado 2 respostas, got %d", len(answers))
	}
	byQuestion := map[int64]Answer{}
	for _, a := range answers {
		byQuestion[a.QuestionID] = a
	}
	if byQuestion[groupQID].OptionID != optA {
		t.Errorf("resposta de grupo divergente: %+v", byQuestion[groupQID])
	}
	if byQuestion[openQID].FreeText != "Pizza" || byQuestion[openQID].OptionID != 0 {
		t.Errorf("resposta aberta divergente: %+v", byQuestion[openQID])
	}
}

func TestUpdateAnswer(t *testing.T) {
	s, groupQID, openQID, optA, optB, participantID := setupAnswerStore(t)
	ctx := context.Background()

	if err := s.SubstituirRespostas(ctx, participantID, []Answer{
		{ID: 1, QuestionID: groupQID, OptionID: optA},
		{ID: 2, QuestionID: openQID, FreeText: "Original"},
	}); err != nil {
		t.Fatalf("gravar respostas: %v", err)
	}

	if err := s.AtualizarResposta(ctx, participantID, groupQID, optB, ""); err != nil {
		t.Fatalf("atualizar resposta de grupo: %v", err)
	}
	if err := s.AtualizarResposta(ctx, participantID, openQID, 0, "[removido pelo organizador]"); err != nil {
		t.Fatalf("atualizar resposta aberta: %v", err)
	}

	answers, err := s.ListarRespostasPorParticipante(ctx, participantID)
	if err != nil {
		t.Fatalf("listar respostas: %v", err)
	}
	byQuestion := map[int64]Answer{}
	for _, a := range answers {
		byQuestion[a.QuestionID] = a
	}
	if byQuestion[groupQID].OptionID != optB {
		t.Errorf("option_id não atualizado: %+v", byQuestion[groupQID])
	}
	if byQuestion[openQID].FreeText != "[removido pelo organizador]" {
		t.Errorf("free_text não atualizado: %+v", byQuestion[openQID])
	}
}

func TestUpdateAnswerNotFound(t *testing.T) {
	s, groupQID, _, optA, _, participantID := setupAnswerStore(t)
	ctx := context.Background()

	// nunca respondida -> nada para atualizar
	if err := s.AtualizarResposta(ctx, participantID, groupQID, optA, ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("esperado ErrNotFound, got %v", err)
	}
}
