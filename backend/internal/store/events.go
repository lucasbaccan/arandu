package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrPinTaken = errors.New("store: PIN já está em uso")

type Event struct {
	ID          int64
	OwnerID     int64
	Title       string
	PINCode     string
	Status      string
	ShowRanking bool
	CreatedAt   time.Time
}

func (s *Store) CreateEvent(ctx context.Context, e Event) (Event, error) {
	e.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO events (id, owner_id, title, pin_code, status, config_show_ranking, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.OwnerID, e.Title, e.PINCode, e.Status, e.ShowRanking,
		e.CreatedAt.Format("2006-01-02 15:04:05"),
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
		`SELECT id, owner_id, title, pin_code, status, config_show_ranking, created_at
		 FROM events WHERE owner_id = ? ORDER BY created_at DESC`,
		ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar eventos: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var createdAt string
		if err := rows.Scan(&e.ID, &e.OwnerID, &e.Title, &e.PINCode, &e.Status, &e.ShowRanking, &createdAt); err != nil {
			return nil, fmt.Errorf("store: ler evento: %w", err)
		}
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar eventos: %w", err)
	}
	return events, nil
}
