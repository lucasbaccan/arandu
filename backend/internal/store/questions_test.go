package store

import (
	"context"
	"errors"
	"testing"
)

func setupQuestionStore(t *testing.T) (*Store, int64) {
	t.Helper()
	s := newTestStore(t)
	ctx := context.Background()
	createOwner(t, s, 1, "dono@x.com")
	if _, err := s.CreateEvent(ctx, Event{ID: 10, OwnerID: 1, Title: "Evento", PINCode: "123456", Status: "PREPARATION"}); err != nil {
		t.Fatalf("criar evento: %v", err)
	}
	return s, 10
}

func TestCreateQuestionWithOptionsAndOrder(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	q1, err := s.CreateQuestion(ctx, Question{ID: 100, EventID: eventID, Title: "Primeira", Type: "GROUP", LayoutView: "TIMELINE"}, []QuestionOption{
		{ID: 1, TextLabel: "A"},
		{ID: 2, TextLabel: "B"},
	})
	if err != nil {
		t.Fatalf("criar primeira pergunta: %v", err)
	}
	if q1.OrderIndex != 0 {
		t.Errorf("primeira pergunta deveria ter order 0, got %d", q1.OrderIndex)
	}
	if len(q1.Options) != 2 || q1.Options[0].TextLabel != "A" || q1.Options[1].ID != 2 {
		t.Errorf("opções divergentes: %+v", q1.Options)
	}

	q2, err := s.CreateQuestion(ctx, Question{ID: 200, EventID: eventID, Title: "Segunda", Type: "INDIVIDUAL", LayoutView: "CENTER"}, []QuestionOption{
		{ID: 3, TextLabel: "Único"},
	})
	if err != nil {
		t.Fatalf("criar segunda pergunta: %v", err)
	}
	if q2.OrderIndex != 1 {
		t.Errorf("segunda pergunta deveria ter order 1, got %d", q2.OrderIndex)
	}
}

func TestCreateQuestionSkipsInvalidOptions(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	q, err := s.CreateQuestion(ctx, Question{ID: 100, EventID: eventID, Title: "P", Type: "GROUP"}, []QuestionOption{
		{ID: 0, TextLabel: "sem id"},
		{ID: 1, TextLabel: "válida"},
	})
	if err != nil {
		t.Fatalf("criar pergunta: %v", err)
	}
	if len(q.Options) != 1 || q.Options[0].TextLabel != "válida" {
		t.Errorf("opção sem id deveria ser ignorada: %+v", q.Options)
	}
}

func TestListQuestionsByEventWithOptions(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	s.CreateQuestion(ctx, Question{ID: 100, EventID: eventID, Title: "Uma", Type: "GROUP"}, []QuestionOption{
		{ID: 1, TextLabel: "A"},
		{ID: 2, TextLabel: "B"},
	})
	s.CreateQuestion(ctx, Question{ID: 200, EventID: eventID, Title: "Duas", Type: "INDIVIDUAL", LayoutView: "TIMELINE"}, []QuestionOption{
		{ID: 3, TextLabel: "C"},
	})

	questions, err := s.ListQuestionsByEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("listar perguntas: %v", err)
	}
	if len(questions) != 2 {
		t.Fatalf("esperado 2 perguntas, got %d", len(questions))
	}
	if questions[0].Title != "Uma" || questions[1].Title != "Duas" {
		t.Errorf("ordem divergente: %+v", questions)
	}
	if len(questions[0].Options) != 2 {
		t.Errorf("primeira pergunta deveria ter 2 opções, got %+v", questions[0].Options)
	}
	if questions[1].Type != "INDIVIDUAL" || questions[1].LayoutView != "TIMELINE" {
		t.Errorf("layout padrão divergente: %+v", questions[1])
	}

	if others, err := s.ListQuestionsByEvent(ctx, 999); err != nil || len(others) != 0 {
		t.Errorf("evento sem perguntas deveria retornar lista vazia, got %v (%v)", others, err)
	}
}

func TestDeleteQuestionCascadesOptions(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	s.CreateQuestion(ctx, Question{ID: 100, EventID: eventID, Title: "Uma", Type: "GROUP"}, []QuestionOption{
		{ID: 1, TextLabel: "A"},
		{ID: 2, TextLabel: "B"},
	})

	if err := s.DeleteQuestion(ctx, 100, eventID); err != nil {
		t.Fatalf("remover pergunta: %v", err)
	}

	questions, _ := s.ListQuestionsByEvent(ctx, eventID)
	if len(questions) != 0 {
		t.Errorf("pergunta deveria ter sido removida: %+v", questions)
	}

	rows, err := s.db.QueryContext(ctx, `SELECT COUNT(*) FROM question_options WHERE question_id = 100`)
	if err != nil {
		t.Fatalf("contar opções: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("sem linha de contagem")
	}
	var count int
	if err := rows.Scan(&count); err != nil {
		t.Fatalf("ler contagem: %v", err)
	}
	if count != 0 {
		t.Errorf("opções deveriam ter sido removidas em cascata, count=%d", count)
	}
}

func TestDeleteQuestionNotFound(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	err := s.DeleteQuestion(ctx, 999, eventID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("esperado ErrNotFound, got %v", err)
	}
}

func TestUpdateQuestion(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	q, err := s.CreateQuestion(ctx, Question{ID: 100, EventID: eventID, Title: "Antes", Type: "GROUP", LayoutView: "TIMELINE"}, []QuestionOption{
		{ID: 1, TextLabel: "A"},
		{ID: 2, TextLabel: "B"},
	})
	if err != nil {
		t.Fatalf("criar pergunta: %v", err)
	}
	s.CreateQuestion(ctx, Question{ID: 200, EventID: eventID, Title: "Outra", Type: "GROUP"}, []QuestionOption{{ID: 3, TextLabel: "C"}})

	updated, err := s.UpdateQuestion(ctx, Question{
		ID: 100, EventID: eventID, Title: "Depois", Type: "GROUP", LayoutView: "CENTER", OrderIndex: q.OrderIndex,
	}, []QuestionOption{
		{ID: 5, TextLabel: "X"},
		{ID: 6, TextLabel: "Y"},
	})
	if err != nil {
		t.Fatalf("atualizar pergunta: %v", err)
	}
	if updated.Title != "Depois" || updated.LayoutView != "CENTER" {
		t.Errorf("pergunta atualizada divergente: %+v", updated)
	}
	if len(updated.Options) != 2 || updated.Options[0].TextLabel != "X" || updated.Options[1].ID != 6 {
		t.Errorf("opções substituídas divergentes: %+v", updated.Options)
	}

	questions, _ := s.ListQuestionsByEvent(ctx, eventID)
	if questions[0].Title != "Depois" || questions[0].OrderIndex != 0 || questions[0].Type != "GROUP" {
		t.Errorf("alterações não persistidas ou ordem/tipo alterados: %+v", questions[0])
	}
	if len(questions[0].Options) != 2 {
		t.Errorf("opções antigas deveriam ser substituídas: %+v", questions[0].Options)
	}
	if questions[1].Title != "Outra" {
		t.Errorf("outras perguntas não deveriam mudar: %+v", questions[1])
	}
}

func TestUpdateQuestionNotFound(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	s.CreateQuestion(ctx, Question{ID: 100, EventID: eventID, Title: "Minha", Type: "GROUP"}, []QuestionOption{{ID: 1, TextLabel: "A"}})

	createOwner(t, s, 2, "outro@x.com")
	s.CreateEvent(ctx, Event{ID: 20, OwnerID: 2, Title: "Outro evento", PINCode: "654321", Status: "PREPARATION"})
	s.CreateQuestion(ctx, Question{ID: 400, EventID: 20, Title: "De outro evento", Type: "GROUP"}, []QuestionOption{{ID: 4, TextLabel: "D"}})

	cases := []struct {
		name string
		id   int64
	}{
		{"inexistente", 999},
		{"de outro evento", 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.UpdateQuestion(ctx, Question{ID: tc.id, EventID: eventID, Title: "X"}, []QuestionOption{{ID: 9, TextLabel: "N"}})
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("esperado ErrNotFound, got %v", err)
			}
		})
	}
}

func TestReorderQuestions(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	s.CreateQuestion(ctx, Question{ID: 100, EventID: eventID, Title: "Primeira", Type: "GROUP"}, []QuestionOption{{ID: 1, TextLabel: "A"}})
	s.CreateQuestion(ctx, Question{ID: 200, EventID: eventID, Title: "Segunda", Type: "GROUP"}, []QuestionOption{{ID: 2, TextLabel: "B"}})
	s.CreateQuestion(ctx, Question{ID: 300, EventID: eventID, Title: "Terceira", Type: "GROUP"}, []QuestionOption{{ID: 3, TextLabel: "C"}})

	if err := s.ReorderQuestions(ctx, eventID, []int64{300, 100, 200}); err != nil {
		t.Fatalf("reordenar: %v", err)
	}

	questions, err := s.ListQuestionsByEvent(ctx, eventID)
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	got := []string{questions[0].Title, questions[1].Title, questions[2].Title}
	want := []string{"Terceira", "Primeira", "Segunda"}
	if got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("ordem esperada %v, got %v", want, got)
	}
	for i, q := range questions {
		if q.OrderIndex != int64(i) {
			t.Errorf("order_index esperado %d para %s, got %d", i, q.Title, q.OrderIndex)
		}
	}
}

func TestReorderQuestionsScopedToEventAndNotFound(t *testing.T) {
	s, eventID := setupQuestionStore(t)
	ctx := context.Background()

	s.CreateQuestion(ctx, Question{ID: 100, EventID: eventID, Title: "Do evento", Type: "GROUP"}, []QuestionOption{{ID: 1, TextLabel: "A"}})

	createOwner(t, s, 2, "outro@x.com")
	s.CreateEvent(ctx, Event{ID: 20, OwnerID: 2, Title: "Outro evento", PINCode: "654321", Status: "PREPARATION"})
	s.CreateQuestion(ctx, Question{ID: 400, EventID: 20, Title: "De outro evento", Type: "GROUP"}, []QuestionOption{{ID: 4, TextLabel: "D"}})

	err := s.ReorderQuestions(ctx, eventID, []int64{400})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("pergunta de outro evento deveria ser ErrNotFound, got %v", err)
	}

	// ordem do outro evento não deve mudar
	others, _ := s.ListQuestionsByEvent(ctx, 20)
	if len(others) != 1 || others[0].OrderIndex != 0 {
		t.Errorf("pergunta de outro evento não deveria ser alterada: %+v", others)
	}

	if err := s.ReorderQuestions(ctx, eventID, []int64{}); err != nil {
		t.Errorf("reordenação vazia deveria ser no-op, got %v", err)
	}
}
