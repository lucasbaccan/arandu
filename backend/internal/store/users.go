package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
	AuthProvider string
	CreatedAt    time.Time
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, u User) (User, error) {
	u.CreatedAt = time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, email, name, password_hash, auth_provider, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.Name, u.PasswordHash, u.AuthProvider, u.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("store: criar usuário: %w", err)
	}
	_ = res
	return u, nil
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, auth_provider, created_at FROM users WHERE email = ?`,
		email,
	)
	return scanUser(row)
}

func (s *Store) FindUserByID(ctx context.Context, id int64) (User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, auth_provider, created_at FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (User, error) {
	var u User
	var createdAt string
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.AuthProvider, &createdAt)
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
