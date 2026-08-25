package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// LiveQAMessage é uma pergunta/recado que um participante manda pro
// organizador durante a apresentação ao vivo. ParticipantID é 0 quando quem
// mandou é convidado/observador (sem identidade além do token anônimo).
// ClientID é um identificador gerado pelo navegador de quem mandou — permite
// que a própria pessoa remova a mensagem dela (ver RemoverPerguntaAoVivoPorCliente);
// vazio = mensagem sem dono de navegador (bancos antigos).
type LiveQAMessage struct {
	ID            int64
	EventID       int64
	ParticipantID int64
	ClientID      string
	Text          string
	Dismissed     bool
	CreatedAt     time.Time
}

func (s *Store) CriarPerguntaAoVivo(ctx context.Context, m LiveQAMessage) (LiveQAMessage, error) {
	m.CreatedAt = time.Now()
	var participantID sql.NullInt64
	if m.ParticipantID != 0 {
		participantID = sql.NullInt64{Int64: m.ParticipantID, Valid: true}
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO live_qa_messages (id, event_id, participant_id, client_id, text, dismissed, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.EventID, participantID, m.ClientID, m.Text, m.Dismissed, m.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return LiveQAMessage{}, fmt.Errorf("store: criar mensagem de q&a: %w", err)
	}
	return m, nil
}

// ListarPerguntasAoVivoPorEvento retorna as mensagens ainda não dispensadas, da
// mais antiga pra mais nova (mesma ordem de ListarParticipantesPorEvento).
func (s *Store) ListarPerguntasAoVivoPorEvento(ctx context.Context, eventID int64) ([]LiveQAMessage, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, event_id, COALESCE(participant_id, 0), COALESCE(client_id, ''), text, dismissed, created_at
		 FROM live_qa_messages WHERE event_id = ? AND dismissed = 0 ORDER BY created_at`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar mensagens de q&a: %w", err)
	}
	defer rows.Close()

	messages := make([]LiveQAMessage, 0)
	for rows.Next() {
		m, err := scanLiveQAMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar mensagens de q&a: %w", err)
	}
	return messages, nil
}

func (s *Store) BuscarPerguntaAoVivoPorID(ctx context.Context, id int64) (LiveQAMessage, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, event_id, COALESCE(participant_id, 0), COALESCE(client_id, ''), text, dismissed, created_at
		 FROM live_qa_messages WHERE id = ?`,
		id,
	)
	return scanLiveQAMessage(row)
}

func (s *Store) DispensarPerguntaAoVivo(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `UPDATE live_qa_messages SET dismissed = 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("store: dispensar mensagem de q&a: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: dispensar mensagem de q&a: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// RemoverPerguntaAoVivoPorCliente remove (dispensa) a mensagem só se o clientID
// bater — quem perguntou consegue apagar a própria pergunta, ninguém mais.
// Retorna ErrNotFound quando não existe mensagem com (id, clientID).
func (s *Store) RemoverPerguntaAoVivoPorCliente(ctx context.Context, id int64, clientID string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE live_qa_messages SET dismissed = 1 WHERE id = ? AND client_id = ? AND dismissed = 0`,
		id, clientID,
	)
	if err != nil {
		return fmt.Errorf("store: remover mensagem de q&a pelo cliente: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: remover mensagem de q&a pelo cliente: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanLiveQAMessage(row scanner) (LiveQAMessage, error) {
	var m LiveQAMessage
	var createdAt string
	err := row.Scan(&m.ID, &m.EventID, &m.ParticipantID, &m.ClientID, &m.Text, &m.Dismissed, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return LiveQAMessage{}, ErrNotFound
	}
	if err != nil {
		return LiveQAMessage{}, fmt.Errorf("store: ler mensagem de q&a: %w", err)
	}
	m.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return LiveQAMessage{}, fmt.Errorf("store: parse created_at: %w", err)
	}
	return m, nil
}
