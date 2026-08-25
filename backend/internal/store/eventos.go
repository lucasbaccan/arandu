package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrPinTaken = errors.New("store: PIN já está em uso")

type Event struct {
	ID                  int64
	OwnerID             int64
	Title               string
	PINCode             string
	Status              string
	ShowRanking         bool
	AllowEdit           bool
	InteractionsEnabled bool
	CreatedAt           time.Time
}

func (s *Store) CriarEvento(ctx context.Context, e Event) (Event, error) {
	e.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO events (id, owner_id, title, pin_code, status, config_show_ranking, allow_edit, interactions_enabled, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.OwnerID, e.Title, e.PINCode, e.Status, e.ShowRanking, e.AllowEdit, e.InteractionsEnabled,
		e.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return Event{}, ErrPinTaken
		}
		return Event{}, fmt.Errorf("store: criar evento: %w", err)
	}
	return e, nil
}

func (s *Store) ListarEventosPorDono(ctx context.Context, ownerID int64) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, allow_edit, interactions_enabled, created_at
		 FROM events WHERE owner_id = ? ORDER BY created_at DESC`,
		ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar eventos: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar eventos: %w", err)
	}
	return events, nil
}

// EventSummary é um Event com os contadores usados na lista de eventos do
// organizador (quantas perguntas, quantas pessoas já responderam).
type EventSummary struct {
	Event
	QuestionCount    int
	ParticipantCount int
}

func (s *Store) ListarResumosDeEventosPorDono(ctx context.Context, ownerID int64) ([]EventSummary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT e.id, e.owner_id, e.title, e.pin_code, e.status, e.config_show_ranking, e.allow_edit, e.interactions_enabled, e.created_at,
		 (SELECT COUNT(*) FROM questions q WHERE q.event_id = e.id),
		 (SELECT COUNT(*) FROM participants p WHERE p.event_id = e.id)
		 FROM events e WHERE e.owner_id = ? ORDER BY e.created_at DESC`,
		ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar eventos com contadores: %w", err)
	}
	defer rows.Close()

	var summaries []EventSummary
	for rows.Next() {
		var es EventSummary
		var createdAt string
		if err := rows.Scan(
			&es.ID, &es.OwnerID, &es.Title, &es.PINCode, &es.Status, &es.ShowRanking, &es.AllowEdit, &es.InteractionsEnabled, &createdAt,
			&es.QuestionCount, &es.ParticipantCount,
		); err != nil {
			return nil, fmt.Errorf("store: ler evento com contadores: %w", err)
		}
		es.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("store: parse created_at: %w", err)
		}
		summaries = append(summaries, es)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar eventos com contadores: %w", err)
	}
	return summaries, nil
}

// BuscarEventoPorID busca o evento por ID, sem checar dono (uso público/participante).
func (s *Store) BuscarEventoPorID(ctx context.Context, id int64) (Event, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, allow_edit, interactions_enabled, created_at
		 FROM events WHERE id = ?`,
		id,
	)
	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	return e, err
}

// BuscarEventoPorPIN busca o evento pelo PIN (uso público, entrada na live sem
// saber o ID). Prioriza eventos ainda ativos; se o PIN já foi reciclado por
// eventos finalizados, cai pro mais recente.
func (s *Store) BuscarEventoPorPIN(ctx context.Context, pin string) (Event, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, allow_edit, interactions_enabled, created_at
		 FROM events WHERE pin_code = ?
		 ORDER BY CASE WHEN status != 'FINISHED' THEN 0 ELSE 1 END, created_at DESC
		 LIMIT 1`,
		pin,
	)
	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	return e, err
}

func (s *Store) BuscarEventoPorIDEDono(ctx context.Context, id, ownerID int64) (Event, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, allow_edit, interactions_enabled, created_at
		 FROM events WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	return e, err
}

func (s *Store) AtualizarEvento(ctx context.Context, e Event) (Event, error) {
	var createdAt string
	err := s.db.QueryRowContext(ctx,
		`UPDATE events SET title = ?, pin_code = ?, status = ?, config_show_ranking = ?, allow_edit = ?
		 WHERE id = ? AND owner_id = ?
		 RETURNING created_at`,
		e.Title, e.PINCode, e.Status, e.ShowRanking, e.AllowEdit, e.ID, e.OwnerID,
	).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return Event{}, ErrPinTaken
		}
		return Event{}, fmt.Errorf("store: atualizar evento: %w", err)
	}
	e.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return Event{}, fmt.Errorf("store: parse created_at: %w", err)
	}
	return e, nil
}

// DefinirInteracoesHabilitadas liga/desliga emoji e Q&A pro evento. Separado de
// AtualizarEvento de propósito: é um controle da apresentação ao vivo, não uma
// configuração geral do evento, e não deve ser afetado pelo PATCH genérico.
func (s *Store) DefinirInteracoesHabilitadas(ctx context.Context, eventID int64, enabled bool) error {
	res, err := s.db.ExecContext(ctx, `UPDATE events SET interactions_enabled = ? WHERE id = ?`, enabled, eventID)
	if err != nil {
		return fmt.Errorf("store: alternar interações: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: alternar interações: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanEvent(row scanner) (Event, error) {
	var e Event
	var createdAt string
	err := row.Scan(&e.ID, &e.OwnerID, &e.Title, &e.PINCode, &e.Status, &e.ShowRanking, &e.AllowEdit, &e.InteractionsEnabled, &createdAt)
	if err != nil {
		return Event{}, err
	}
	e.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return Event{}, fmt.Errorf("store: parse created_at: %w", err)
	}
	return e, nil
}
