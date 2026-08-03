package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Participant struct {
	ID        int64
	EventID   int64
	Email     string
	Photo     string
	CreatedAt time.Time
}

// UpsertParticipant cria o participante ou reaproveita o existente (mesmo
// evento + e-mail), atualizando a foto se uma nova foi enviada.
func (s *Store) UpsertParticipant(ctx context.Context, p Participant) (Participant, error) {
	existing, err := s.FindParticipantByEventAndEmail(ctx, p.EventID, p.Email)
	if err == nil {
		if _, err := s.db.ExecContext(ctx,
			`UPDATE participants SET photo = ? WHERE id = ?`, p.Photo, existing.ID,
		); err != nil {
			return Participant{}, fmt.Errorf("store: atualizar participante: %w", err)
		}
		existing.Photo = p.Photo
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Participant{}, err
	}

	p.CreatedAt = time.Now()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO participants (id, event_id, email, photo, created_at) VALUES (?, ?, ?, ?, ?)`,
		p.ID, p.EventID, p.Email, p.Photo, p.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		if isUniqueViolation(err) {
			// condição de corrida rara: outra requisição criou o participante
			// entre a busca e a inserção acima.
			return s.FindParticipantByEventAndEmail(ctx, p.EventID, p.Email)
		}
		return Participant{}, fmt.Errorf("store: criar participante: %w", err)
	}
	return p, nil
}

func (s *Store) FindParticipantByEventAndEmail(ctx context.Context, eventID int64, email string) (Participant, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, event_id, email, photo, created_at FROM participants WHERE event_id = ? AND email = ?`,
		eventID, email,
	)
	return scanParticipant(row)
}

func (s *Store) FindParticipantByID(ctx context.Context, id int64) (Participant, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, event_id, email, photo, created_at FROM participants WHERE id = ?`,
		id,
	)
	return scanParticipant(row)
}

// ListParticipantsByEvent retorna os participantes de um evento, do mais antigo ao mais recente.
func (s *Store) ListParticipantsByEvent(ctx context.Context, eventID int64) ([]Participant, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, event_id, email, photo, created_at FROM participants WHERE event_id = ? ORDER BY created_at`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar participantes: %w", err)
	}
	defer rows.Close()

	participants := make([]Participant, 0)
	for rows.Next() {
		p, err := scanParticipant(rows)
		if err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar participantes: %w", err)
	}
	return participants, nil
}

func scanParticipant(row scanner) (Participant, error) {
	var p Participant
	var createdAt string
	err := row.Scan(&p.ID, &p.EventID, &p.Email, &p.Photo, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Participant{}, ErrNotFound
	}
	if err != nil {
		return Participant{}, fmt.Errorf("store: ler participante: %w", err)
	}
	p.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return Participant{}, fmt.Errorf("store: parse created_at: %w", err)
	}
	return p, nil
}
