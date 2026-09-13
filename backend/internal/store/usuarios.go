package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	RoleSuperAdmin = "super_admin"
	RoleOrganizer  = "organizer"
)

type User struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
	AuthProvider string
	Role         string
	CreatedAt    time.Time
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// CriarUsuario decide o papel do usuário na hora de gravar: quem se cadastra
// num banco ainda sem ninguém vira super admin automaticamente (não existe
// tela de convite/promoção — o bootstrap tem que ser implícito). A contagem e
// o insert rodam na mesma transação pra dois cadastros simultâneos num banco
// vazio não conseguirem virar super admin os dois.
func (s *Store) CriarUsuario(ctx context.Context, u User) (User, error) {
	u.CreatedAt = time.Now()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("store: criar usuário: begin tx: %w", err)
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return User{}, fmt.Errorf("store: criar usuário: contar: %w", err)
	}
	u.Role = RoleOrganizer
	if count == 0 {
		u.Role = RoleSuperAdmin
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO users (id, email, name, password_hash, auth_provider, role, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.Name, u.PasswordHash, u.AuthProvider, u.Role, u.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("store: criar usuário: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("store: criar usuário: commit: %w", err)
	}
	return u, nil
}

func (s *Store) BuscarUsuarioPorEmail(ctx context.Context, email string) (User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, auth_provider, role, created_at FROM users WHERE email = ?`,
		email,
	)
	return scanUser(row)
}

func (s *Store) BuscarUsuarioPorID(ctx context.Context, id int64) (User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, auth_provider, role, created_at FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

// UserSummary é um User com a contagem de eventos, usado na tela de
// gerenciamento de usuários do super admin.
type UserSummary struct {
	User
	EventCount int
}

func (s *Store) ListarUsuariosComContagemDeEventos(ctx context.Context) ([]UserSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.email, u.name, u.password_hash, u.auth_provider, u.role, u.created_at,
		       (SELECT COUNT(*) FROM events e WHERE e.owner_id = u.id)
		FROM users u ORDER BY u.created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("store: listar usuários: %w", err)
	}
	defer rows.Close()

	var summaries []UserSummary
	for rows.Next() {
		var us UserSummary
		var createdAt string
		if err := rows.Scan(
			&us.ID, &us.Email, &us.Name, &us.PasswordHash, &us.AuthProvider, &us.Role, &createdAt,
			&us.EventCount,
		); err != nil {
			return nil, fmt.Errorf("store: ler usuário: %w", err)
		}
		us.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("store: parse created_at: %w", err)
		}
		summaries = append(summaries, us)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterar usuários: %w", err)
	}
	return summaries, nil
}

// ExcluirUsuario apaga a conta e, antes, todos os eventos que ela possui —
// events.owner_id não tem ON DELETE CASCADE (o dono normalmente não é
// apagável; aqui é o caso excepcional do super admin removendo a conta de
// outra pessoa). Apagar os eventos primeiro arrasta perguntas, participantes,
// respostas etc. via cascade já existente em cada uma dessas tabelas.
func (s *Store) ExcluirUsuario(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: excluir usuário: begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM events WHERE owner_id = ?`, id); err != nil {
		return fmt.Errorf("store: excluir usuário: eventos: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("store: excluir usuário: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: excluir usuário: rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}

// AtualizarSenha troca o hash de senha de um usuário (fluxo "trocar senha
// logado": o usuário já está autenticado via requireAuth).
func (s *Store) AtualizarSenha(ctx context.Context, id int64, hash string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`, hash, id,
	)
	if err != nil {
		return fmt.Errorf("store: atualizar senha: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: atualizar senha: rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (User, error) {
	var u User
	var createdAt string
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.AuthProvider, &u.Role, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("store: ler usuário: %w", err)
	}
	u.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return User{}, fmt.Errorf("store: parse created_at: %w", err)
	}
	return u, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
