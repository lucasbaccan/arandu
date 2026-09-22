package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const chaveRegistroHabilitado = "registration_enabled"

// RegistroHabilitado indica se novas contas de organizador podem ser criadas.
// Sem linha salva (banco novo, ou nunca alterado pelo super admin), o padrão
// é permitido.
func (s *Store) RegistroHabilitado(ctx context.Context) (bool, error) {
	var v string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM app_settings WHERE key = ?`, chaveRegistroHabilitado,
	).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("store: ler configuração de registro: %w", err)
	}
	return v == "true", nil
}

func (s *Store) DefinirRegistroHabilitado(ctx context.Context, enabled bool) error {
	v := "false"
	if enabled {
		v = "true"
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO app_settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		chaveRegistroHabilitado, v,
	)
	if err != nil {
		return fmt.Errorf("store: salvar configuração de registro: %w", err)
	}
	return nil
}

const chaveEmailHabilitado = "email_enabled"

// EmailHabilitado indica se o e-mail dos participantes é mostrado no painel
// de respostas do organizador e pedido na tela da plateia (que sem ele já
// cai direto no fluxo de convidado). Sem linha salva, o padrão é mostrado —
// mesmo comportamento de sempre, pra quem nunca mexeu nessa configuração.
func (s *Store) EmailHabilitado(ctx context.Context) (bool, error) {
	var v string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM app_settings WHERE key = ?`, chaveEmailHabilitado,
	).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("store: ler configuração de e-mail: %w", err)
	}
	return v == "true", nil
}

func (s *Store) DefinirEmailHabilitado(ctx context.Context, enabled bool) error {
	v := "false"
	if enabled {
		v = "true"
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO app_settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		chaveEmailHabilitado, v,
	)
	if err != nil {
		return fmt.Errorf("store: salvar configuração de e-mail: %w", err)
	}
	return nil
}
