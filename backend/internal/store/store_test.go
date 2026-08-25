package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("abrir store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := Migrate(db); err != nil {
		t.Fatalf("migrar store: %v", err)
	}
	return New(db)
}

func TestCreateAndFindUser(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	u, err := s.CriarUsuario(ctx, User{
		ID:           1001,
		Email:        "ana@exemplo.com",
		Name:         "Ana",
		PasswordHash: "hash",
		AuthProvider: "email",
	})
	if err != nil {
		t.Fatalf("criar usuário: %v", err)
	}
	if u.ID != 1001 {
		t.Errorf("id esperado 1001, got %d", u.ID)
	}

	got, err := s.BuscarUsuarioPorEmail(ctx, "ana@exemplo.com")
	if err != nil {
		t.Fatalf("buscar por email: %v", err)
	}
	if got.Name != "Ana" || got.PasswordHash != "hash" || got.AuthProvider != "email" {
		t.Errorf("usuário divergente: %+v", got)
	}

	got, err = s.BuscarUsuarioPorID(ctx, 1001)
	if err != nil {
		t.Fatalf("buscar por id: %v", err)
	}
	if got.Email != "ana@exemplo.com" {
		t.Errorf("email esperado ana@exemplo.com, got %s", got.Email)
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.CriarUsuario(ctx, User{ID: 1, Email: "a@b.com", Name: "A", PasswordHash: "h", AuthProvider: "email"})
	if err != nil {
		t.Fatalf("criar primeiro usuário: %v", err)
	}
	_, err = s.CriarUsuario(ctx, User{ID: 2, Email: "a@b.com", Name: "B", PasswordHash: "h", AuthProvider: "email"})
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("esperado ErrEmailTaken, got %v", err)
	}
}

func TestFindUserNotFound(t *testing.T) {
	s := newTestStore(t)

	if _, err := s.BuscarUsuarioPorEmail(context.Background(), "nada@x.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("esperado ErrNotFound, got %v", err)
	}
	if _, err := s.BuscarUsuarioPorID(context.Background(), 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("esperado ErrNotFound, got %v", err)
	}
}
