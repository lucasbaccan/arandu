package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// EventLiveState é o retrato persistido do estado ao vivo de um evento —
// espelha live.EventState (menos Revealed, que mora em revealed_answers).
// Existe pra sobreviver a um restart do servidor: live.Manager guarda tudo
// em memória pra velocidade, mas cada mutação também é gravada aqui.
type EventLiveState struct {
	CurrentQuestionID  int64
	Blanked            bool
	Message            string
	AnswersHidden      bool
	NamesHidden        bool
	PresentDensityMode string
}

// GetEventLiveState busca o estado ao vivo persistido do evento. Evento
// nunca apresentado (sem linha em event_live_state) retorna o zero value,
// não erro — equivalente a "nada foi definido ainda".
func (s *Store) GetEventLiveState(ctx context.Context, eventID int64) (EventLiveState, error) {
	var st EventLiveState
	row := s.db.QueryRowContext(ctx,
		`SELECT current_question_id, blanked, message, answers_hidden, names_hidden, present_density_mode
		 FROM event_live_state WHERE event_id = ?`,
		eventID,
	)
	err := row.Scan(
		&st.CurrentQuestionID, &st.Blanked, &st.Message, &st.AnswersHidden, &st.NamesHidden, &st.PresentDensityMode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return EventLiveState{}, nil
	}
	if err != nil {
		return EventLiveState{}, fmt.Errorf("store: buscar estado ao vivo: %w", err)
	}
	return st, nil
}

func (s *Store) SetEventCurrentQuestion(ctx context.Context, eventID, questionID int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO event_live_state (event_id, current_question_id) VALUES (?, ?)
		 ON CONFLICT(event_id) DO UPDATE SET current_question_id = excluded.current_question_id`,
		eventID, questionID,
	)
	if err != nil {
		return fmt.Errorf("store: definir pergunta atual: %w", err)
	}
	return nil
}

func (s *Store) SetEventBlanked(ctx context.Context, eventID int64, blanked bool) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO event_live_state (event_id, blanked) VALUES (?, ?)
		 ON CONFLICT(event_id) DO UPDATE SET blanked = excluded.blanked`,
		eventID, blanked,
	)
	if err != nil {
		return fmt.Errorf("store: definir tela em branco: %w", err)
	}
	return nil
}

func (s *Store) SetEventMessage(ctx context.Context, eventID int64, message string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO event_live_state (event_id, message) VALUES (?, ?)
		 ON CONFLICT(event_id) DO UPDATE SET message = excluded.message`,
		eventID, message,
	)
	if err != nil {
		return fmt.Errorf("store: definir aviso: %w", err)
	}
	return nil
}

func (s *Store) SetEventAnswersHidden(ctx context.Context, eventID int64, hidden bool) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO event_live_state (event_id, answers_hidden) VALUES (?, ?)
		 ON CONFLICT(event_id) DO UPDATE SET answers_hidden = excluded.answers_hidden`,
		eventID, hidden,
	)
	if err != nil {
		return fmt.Errorf("store: definir ocultar respostas: %w", err)
	}
	return nil
}

func (s *Store) SetEventNamesHidden(ctx context.Context, eventID int64, hidden bool) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO event_live_state (event_id, names_hidden) VALUES (?, ?)
		 ON CONFLICT(event_id) DO UPDATE SET names_hidden = excluded.names_hidden`,
		eventID, hidden,
	)
	if err != nil {
		return fmt.Errorf("store: definir ocultar nomes: %w", err)
	}
	return nil
}

func (s *Store) SetEventPresentDensityMode(ctx context.Context, eventID int64, mode string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO event_live_state (event_id, present_density_mode) VALUES (?, ?)
		 ON CONFLICT(event_id) DO UPDATE SET present_density_mode = excluded.present_density_mode`,
		eventID, mode,
	)
	if err != nil {
		return fmt.Errorf("store: definir modo de densidade da apresentação: %w", err)
	}
	return nil
}

// RevealAnswer grava que um participante foi revelado numa pergunta.
// INSERT OR IGNORE: revelar de novo quem já está revelado é um no-op, não erro.
func (s *Store) RevealAnswer(ctx context.Context, eventID, questionID, participantID int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO revealed_answers (event_id, question_id, participant_id) VALUES (?, ?, ?)`,
		eventID, questionID, participantID,
	)
	if err != nil {
		return fmt.Errorf("store: revelar resposta: %w", err)
	}
	return nil
}

func (s *Store) UnrevealAnswer(ctx context.Context, eventID, questionID, participantID int64) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM revealed_answers WHERE event_id = ? AND question_id = ? AND participant_id = ?`,
		eventID, questionID, participantID,
	)
	if err != nil {
		return fmt.Errorf("store: desrevelar resposta: %w", err)
	}
	return nil
}

func (s *Store) RevealAllAnswers(ctx context.Context, eventID, questionID int64, participantIDs []int64) error {
	for _, participantID := range participantIDs {
		if _, err := s.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO revealed_answers (event_id, question_id, participant_id) VALUES (?, ?, ?)`,
			eventID, questionID, participantID,
		); err != nil {
			return fmt.Errorf("store: revelar todos: %w", err)
		}
	}
	return nil
}

func (s *Store) ResetRevealedForQuestion(ctx context.Context, eventID, questionID int64) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM revealed_answers WHERE event_id = ? AND question_id = ?`,
		eventID, questionID,
	)
	if err != nil {
		return fmt.Errorf("store: reiniciar revelação da pergunta: %w", err)
	}
	return nil
}

func (s *Store) ResetRevealedForEvent(ctx context.Context, eventID int64) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM revealed_answers WHERE event_id = ?`,
		eventID,
	)
	if err != nil {
		return fmt.Errorf("store: reiniciar toda a revelação: %w", err)
	}
	return nil
}

// ListRevealedByEvent retorna, pra cada pergunta do evento, o conjunto de
// participantIDs já revelados — usado pra reidratar o estado em memória
// (live.Manager) na primeira vez que o evento é acessado depois de um
// restart do processo.
func (s *Store) ListRevealedByEvent(ctx context.Context, eventID int64) (map[int64]map[int64]bool, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT question_id, participant_id FROM revealed_answers WHERE event_id = ?`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar revelados: %w", err)
	}
	defer rows.Close()

	revealed := make(map[int64]map[int64]bool)
	for rows.Next() {
		var questionID, participantID int64
		if err := rows.Scan(&questionID, &participantID); err != nil {
			return nil, fmt.Errorf("store: ler revelado: %w", err)
		}
		set, ok := revealed[questionID]
		if !ok {
			set = make(map[int64]bool)
			revealed[questionID] = set
		}
		set[participantID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar revelados: %w", err)
	}
	return revealed, nil
}
