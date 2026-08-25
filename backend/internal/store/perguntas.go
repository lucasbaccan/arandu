package store

import (
	"context"
	"fmt"
)

type Question struct {
	ID         int64
	EventID    int64
	Title      string
	Type       string
	LayoutView string
	OrderIndex int64
}

type QuestionOption struct {
	ID         int64
	QuestionID int64
	TextLabel  string
}

type QuestionWithOptions struct {
	Question
	Options []QuestionOption
}

// CriarPergunta insere uma pergunta com suas opções, posicionada ao final do evento.
func (s *Store) CriarPergunta(ctx context.Context, q Question, options []QuestionOption) (QuestionWithOptions, error) {
	var nextOrder int64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(order_index), -1) + 1 FROM questions WHERE event_id = ?`,
		q.EventID,
	).Scan(&nextOrder)
	if err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: próximo índice da pergunta: %w", err)
	}
	q.OrderIndex = nextOrder

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: iniciar transação: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO questions (id, event_id, title, type, layout_view, order_index)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		q.ID, q.EventID, q.Title, q.Type, q.LayoutView, q.OrderIndex,
	)
	if err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: criar pergunta: %w", err)
	}

	opts := make([]QuestionOption, 0, len(options))
	for _, opt := range options {
		if opt.ID == 0 || opt.TextLabel == "" {
			continue
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO question_options (id, question_id, text_label) VALUES (?, ?, ?)`,
			opt.ID, q.ID, opt.TextLabel,
		)
		if err != nil {
			return QuestionWithOptions{}, fmt.Errorf("store: criar opção: %w", err)
		}
		opts = append(opts, QuestionOption{ID: opt.ID, QuestionID: q.ID, TextLabel: opt.TextLabel})
	}

	if err := tx.Commit(); err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: confirmar pergunta: %w", err)
	}
	return QuestionWithOptions{Question: q, Options: opts}, nil
}

// ListarPerguntasPorEvento retorna as perguntas do evento na ordem definida, com opções.
func (s *Store) ListarPerguntasPorEvento(ctx context.Context, eventID int64) ([]QuestionWithOptions, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, event_id, title, type, layout_view, order_index
		 FROM questions WHERE event_id = ? ORDER BY order_index, id`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar perguntas: %w", err)
	}
	defer rows.Close()

	questions := make([]QuestionWithOptions, 0)
	for rows.Next() {
		var q Question
		if err := rows.Scan(&q.ID, &q.EventID, &q.Title, &q.Type, &q.LayoutView, &q.OrderIndex); err != nil {
			return nil, fmt.Errorf("store: ler pergunta: %w", err)
		}
		questions = append(questions, QuestionWithOptions{Question: q})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar perguntas: %w", err)
	}

	for i := range questions {
		opts, err := s.listOptionsByQuestion(ctx, questions[i].ID)
		if err != nil {
			return nil, err
		}
		questions[i].Options = opts
	}
	return questions, nil
}

func (s *Store) listOptionsByQuestion(ctx context.Context, questionID int64) ([]QuestionOption, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, question_id, text_label FROM question_options WHERE question_id = ? ORDER BY id`,
		questionID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar opções: %w", err)
	}
	defer rows.Close()

	opts := make([]QuestionOption, 0)
	for rows.Next() {
		var o QuestionOption
		if err := rows.Scan(&o.ID, &o.QuestionID, &o.TextLabel); err != nil {
			return nil, fmt.Errorf("store: ler opção: %w", err)
		}
		opts = append(opts, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar opções: %w", err)
	}
	return opts, nil
}

// AtualizarPergunta atualiza o texto e substitui as opções de uma pergunta do evento.
// Tipo, layout e ordem são preservados.
func (s *Store) AtualizarPergunta(ctx context.Context, q Question, options []QuestionOption) (QuestionWithOptions, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: iniciar transação: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE questions SET title = ?, layout_view = ? WHERE id = ? AND event_id = ?`,
		q.Title, q.LayoutView, q.ID, q.EventID,
	)
	if err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: atualizar pergunta: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: atualizar pergunta: %w", err)
	}
	if n == 0 {
		return QuestionWithOptions{}, ErrNotFound
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM question_options WHERE question_id = ?`, q.ID,
	); err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: substituir opções: %w", err)
	}

	opts := make([]QuestionOption, 0, len(options))
	for _, opt := range options {
		if opt.ID == 0 || opt.TextLabel == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO question_options (id, question_id, text_label) VALUES (?, ?, ?)`,
			opt.ID, q.ID, opt.TextLabel,
		); err != nil {
			return QuestionWithOptions{}, fmt.Errorf("store: criar opção: %w", err)
		}
		opts = append(opts, QuestionOption{ID: opt.ID, QuestionID: q.ID, TextLabel: opt.TextLabel})
	}

	if err := tx.Commit(); err != nil {
		return QuestionWithOptions{}, fmt.Errorf("store: confirmar pergunta: %w", err)
	}
	return QuestionWithOptions{Question: q, Options: opts}, nil
}

// ReordenarPerguntas redefine a ordem das perguntas de um evento na sequência dada.
func (s *Store) ReordenarPerguntas(ctx context.Context, eventID int64, ids []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: iniciar transação: %w", err)
	}
	defer tx.Rollback()

	for i, id := range ids {
		res, err := tx.ExecContext(ctx,
			`UPDATE questions SET order_index = ? WHERE id = ? AND event_id = ?`,
			int64(i), id, eventID,
		)
		if err != nil {
			return fmt.Errorf("store: reordenar pergunta: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("store: reordenar pergunta: %w", err)
		}
		if n == 0 {
			return ErrNotFound
		}
	}
	return tx.Commit()
}

// RemoverPergunta remove a pergunta (e suas opções, via cascade) de um evento do dono.
func (s *Store) RemoverPergunta(ctx context.Context, questionID, eventID int64) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM questions WHERE id = ? AND event_id = ?`,
		questionID, eventID,
	)
	if err != nil {
		return fmt.Errorf("store: remover pergunta: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: remover pergunta: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
