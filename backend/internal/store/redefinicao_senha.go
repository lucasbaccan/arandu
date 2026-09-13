package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ResetToken é o link temporário que o super admin gera para alguém redefinir
// a própria senha sem o admin precisar saber ou escolher a senha da pessoa.
type ResetToken struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
}

// CriarTokenRedefinicaoSenha apaga eventuais tokens anteriores do usuário (só
// um link válido por vez — gerar outro invalida o anterior) e grava o novo,
// na mesma transação.
func (s *Store) CriarTokenRedefinicaoSenha(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: criar token de redefinição: begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("store: criar token de redefinição: limpar antigos: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("store: criar token de redefinição: %w", err)
	}
	return tx.Commit()
}

func (s *Store) BuscarTokenRedefinicaoSenha(ctx context.Context, token string) (ResetToken, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT token, user_id, expires_at FROM password_reset_tokens WHERE token = ?`, token,
	)
	var rt ResetToken
	var expiresAt string
	err := row.Scan(&rt.Token, &rt.UserID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ResetToken{}, ErrNotFound
	}
	if err != nil {
		return ResetToken{}, fmt.Errorf("store: buscar token de redefinição: %w", err)
	}
	rt.ExpiresAt, err = time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return ResetToken{}, fmt.Errorf("store: parse expires_at: %w", err)
	}
	return rt, nil
}

func (s *Store) ExcluirTokenRedefinicaoSenha(ctx context.Context, token string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE token = ?`, token); err != nil {
		return fmt.Errorf("store: excluir token de redefinição: %w", err)
	}
	return nil
}
