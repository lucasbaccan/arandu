package auth

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("segredo123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "segredo123" {
		t.Error("hash não pode ser texto plano")
	}
	if !CheckPassword(hash, "segredo123") {
		t.Error("senha correta deveria validar")
	}
	if CheckPassword(hash, "senha-errada") {
		t.Error("senha incorreta não deveria validar")
	}
}

func TestSessionTokenRoundTrip(t *testing.T) {
	secret := "test-secret"
	token, err := NewSessionToken(secret, 4242, 24)
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}

	userID, err := VerifySessionToken(secret, token)
	if err != nil {
		t.Fatalf("verificar token: %v", err)
	}
	if userID != 4242 {
		t.Errorf("userID esperado 4242, got %d", userID)
	}
}

func TestSessionTokenWrongSecret(t *testing.T) {
	token, err := NewSessionToken("secret-a", 1, 1)
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}
	if _, err := VerifySessionToken("secret-b", token); err == nil {
		t.Error("token com secret errado deveria falhar")
	}
}

func TestSessionTokenTampered(t *testing.T) {
	token, err := NewSessionToken("secret-a", 1, 1)
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}
	tampered := token + "x"
	if _, err := VerifySessionToken("secret-a", tampered); err == nil {
		t.Error("token adulterado deveria falhar")
	}
}
