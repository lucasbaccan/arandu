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
	InteractionsEnabled bool
	CreatedAt           time.Time
}

func (s *Store) CreateEvent(ctx context.Context, e Event) (Event, error) {
	e.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO events (id, owner_id, title, pin_code, status, config_show_ranking, interactions_enabled, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.OwnerID, e.Title, e.PINCode, e.Status, e.ShowRanking, e.InteractionsEnabled,
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

func (s *Store) ListEventsByOwner(ctx context.Context, ownerID int64) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, interactions_enabled, created_at
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

// FindEventByID busca o evento por ID, sem checar dono (uso público/participante).
func (s *Store) FindEventByID(ctx context.Context, id int64) (Event, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, interactions_enabled, created_at
		 FROM events WHERE id = ?`,
		id,
	)
	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	return e, err
}

// FindEventByPIN busca o evento pelo PIN (uso público, entrada na live sem
// saber o ID). Prioriza eventos ainda ativos; se o PIN já foi reciclado por
// eventos finalizados, cai pro mais recente.
func (s *Store) FindEventByPIN(ctx context.Context, pin string) (Event, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, interactions_enabled, created_at
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

func (s *Store) FindEventByIDAndOwner(ctx context.Context, id, ownerID int64) (Event, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, interactions_enabled, created_at
		 FROM events WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	return e, err
}

func (s *Store) UpdateEvent(ctx context.Context, e Event) (Event, error) {
	var createdAt string
	err := s.db.QueryRowContext(ctx,
		`UPDATE events SET title = ?, pin_code = ?, status = ?, config_show_ranking = ?
		 WHERE id = ? AND owner_id = ?
		 RETURNING created_at`,
		e.Title, e.PINCode, e.Status, e.ShowRanking, e.ID, e.OwnerID,
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

// SetInteractionsEnabled liga/desliga emoji e Q&A pro evento. Separado de
// UpdateEvent de propósito: é um controle da apresentação ao vivo, não uma
// configuração geral do evento, e não deve ser afetado pelo PATCH genérico.
func (s *Store) SetInteractionsEnabled(ctx context.Context, eventID int64, enabled bool) error {
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
	err := row.Scan(&e.ID, &e.OwnerID, &e.Title, &e.PINCode, &e.Status, &e.ShowRanking, &e.InteractionsEnabled, &createdAt)
	if err != nil {
		return Event{}, err
	}
	e.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return Event{}, fmt.Errorf("store: parse created_at: %w", err)
	}
	return e, nil
}
