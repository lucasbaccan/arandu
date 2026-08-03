package store

import (
	"context"
	"database/sql"
	"fmt"
)

// Answer representa a resposta de um participante a uma pergunta.
// OptionID é 0 quando a pergunta é de resposta aberta (usa FreeText).
type Answer struct {
	ID         int64
	QuestionID int64
	OptionID   int64
	FreeText   string
}

// ReplaceAnswers grava as respostas do participante, substituindo qualquer
// resposta anterior às mesmas perguntas (permite reenvio).
func (s *Store) ReplaceAnswers(ctx context.Context, participantID int64, answers []Answer) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: iniciar transação: %w", err)
	}
	defer tx.Rollback()

	for _, a := range answers {
		var optionID sql.NullInt64
		if a.OptionID != 0 {
			optionID = sql.NullInt64{Int64: a.OptionID, Valid: true}
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO answers (id, question_id, participant_id, option_id, free_text)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT(question_id, participant_id) DO UPDATE SET
				option_id = excluded.option_id,
				free_text = excluded.free_text`,
			a.ID, a.QuestionID, participantID, optionID, a.FreeText,
		)
		if err != nil {
			return fmt.Errorf("store: gravar resposta: %w", err)
		}
	}
	return tx.Commit()
}

// ListAnswersByParticipant retorna as respostas de um participante, por pergunta.
func (s *Store) ListAnswersByParticipant(ctx context.Context, participantID int64) ([]Answer, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, question_id, COALESCE(option_id, 0), free_text
		 FROM answers WHERE participant_id = ?`,
		participantID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar respostas: %w", err)
	}
	defer rows.Close()

	answers := make([]Answer, 0)
	for rows.Next() {
		var a Answer
		if err := rows.Scan(&a.ID, &a.QuestionID, &a.OptionID, &a.FreeText); err != nil {
			return nil, fmt.Errorf("store: ler resposta: %w", err)
		}
		answers = append(answers, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar respostas: %w", err)
	}
	return answers, nil
}

// UpdateAnswer atualiza a resposta de um participante a uma pergunta (moderação
// pelo organizador). Retorna ErrNotFound se a resposta não existir.
func (s *Store) UpdateAnswer(ctx context.Context, participantID, questionID, optionID int64, freeText string) error {
	var opt sql.NullInt64
	if optionID != 0 {
		opt = sql.NullInt64{Int64: optionID, Valid: true}
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE answers SET option_id = ?, free_text = ? WHERE question_id = ? AND participant_id = ?`,
		opt, freeText, questionID, participantID,
	)
	if err != nil {
		return fmt.Errorf("store: atualizar resposta: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: atualizar resposta: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
