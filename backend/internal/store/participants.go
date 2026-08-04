package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type Participant struct {
	ID        int64
	EventID   int64
	Email     string
	Photo     string
	EditToken string
	CreatedAt time.Time
}

// newEditToken gera um token de edição opaco e imprevisível: é o único
// segredo que autoriza um participante a editar suas respostas depois do
// envio (não há login para participantes).
func newEditToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("store: gerar token de edição: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ensureEditToken preenche o token de um participante que ficou sem um (ex:
// criado antes do link de edição existir), persistindo o valor gerado.
func (s *Store) ensureEditToken(ctx context.Context, p *Participant) error {
	if p.EditToken != "" {
		return nil
	}
	token, err := newEditToken()
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE participants SET edit_token = ? WHERE id = ?`, token, p.ID,
	); err != nil {
		return fmt.Errorf("store: gerar token de edição: %w", err)
	}
	p.EditToken = token
	return nil
}

// UpsertParticipant cria o participante ou reaproveita o existente (mesmo
// evento + e-mail). A foto só é sobrescrita quando uma nova é enviada, para
// não apagar uma foto já salva quando o participante reenvia sem trocar a
// foto (ex: editando respostas pelo link de edição).
func (s *Store) UpsertParticipant(ctx context.Context, p Participant) (Participant, error) {
	existing, err := s.FindParticipantByEventAndEmail(ctx, p.EventID, p.Email)
	if err == nil {
		if p.Photo != "" && p.Photo != existing.Photo {
			if _, err := s.db.ExecContext(ctx,
				`UPDATE participants SET photo = ? WHERE id = ?`, p.Photo, existing.ID,
			); err != nil {
				return Participant{}, fmt.Errorf("store: atualizar participante: %w", err)
			}
			existing.Photo = p.Photo
		}
		if err := s.ensureEditToken(ctx, &existing); err != nil {
			return Participant{}, err
		}
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Participant{}, err
	}

	token, err := newEditToken()
	if err != nil {
		return Participant{}, err
	}
	p.EditToken = token
	p.CreatedAt = time.Now()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO participants (id, event_id, email, photo, edit_token, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.EventID, p.Email, p.Photo, p.EditToken, p.CreatedAt.Format(time.RFC3339),
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
		`SELECT id, event_id, email, photo, edit_token, created_at FROM participants WHERE event_id = ? AND email = ?`,
		eventID, email,
	)
	return scanParticipant(row)
}

func (s *Store) FindParticipantByEventAndToken(ctx context.Context, eventID int64, token string) (Participant, error) {
	if token == "" {
		return Participant{}, ErrNotFound
	}
	row := s.db.QueryRowContext(ctx,
		`SELECT id, event_id, email, photo, edit_token, created_at FROM participants WHERE event_id = ? AND edit_token = ?`,
		eventID, token,
	)
	return scanParticipant(row)
}

func (s *Store) FindParticipantByID(ctx context.Context, id int64) (Participant, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, event_id, email, photo, edit_token, created_at FROM participants WHERE id = ?`,
		id,
	)
	return scanParticipant(row)
}

// UpdateParticipantPhoto permite ao organizador definir/corrigir a foto de um
// participante (ex: participante pediu para atualizar por fora do link).
func (s *Store) UpdateParticipantPhoto(ctx context.Context, participantID int64, photo string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE participants SET photo = ? WHERE id = ?`, photo, participantID)
	if err != nil {
		return fmt.Errorf("store: atualizar foto do participante: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: atualizar foto do participante: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListParticipantsByEvent retorna os participantes de um evento, do mais antigo ao mais recente.
func (s *Store) ListParticipantsByEvent(ctx context.Context, eventID int64) ([]Participant, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, event_id, email, photo, edit_token, created_at FROM participants WHERE event_id = ? ORDER BY created_at`,
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
	for i := range participants {
		if err := s.ensureEditToken(ctx, &participants[i]); err != nil {
			return nil, err
		}
	}
	return participants, nil
}

func scanParticipant(row scanner) (Participant, error) {
	var p Participant
	var createdAt string
	err := row.Scan(&p.ID, &p.EventID, &p.Email, &p.Photo, &p.EditToken, &createdAt)
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
