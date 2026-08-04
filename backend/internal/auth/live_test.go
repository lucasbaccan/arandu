package auth

import "testing"

func TestLiveViewerTokenRoundTripParticipant(t *testing.T) {
	secret := "test-secret"
	token, err := NewLiveViewerToken(secret, 42, 100, LiveRoleParticipant, 6)
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}

	claims, err := VerifyLiveViewerToken(secret, token)
	if err != nil {
		t.Fatalf("verificar token: %v", err)
	}
	if claims.EventID != "42" {
		t.Errorf("eventId esperado 42, got %s", claims.EventID)
	}
	if claims.ParticipantID != "100" {
		t.Errorf("participantId esperado 100, got %s", claims.ParticipantID)
	}
	if claims.Role != LiveRoleParticipant {
		t.Errorf("role esperado participante, got %s", claims.Role)
	}
}

func TestLiveViewerTokenObserverHasNoParticipantID(t *testing.T) {
	secret := "test-secret"
	token, err := NewLiveViewerToken(secret, 42, 0, LiveRoleObserver, 6)
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}

	claims, err := VerifyLiveViewerToken(secret, token)
	if err != nil {
		t.Fatalf("verificar token: %v", err)
	}
	if claims.ParticipantID != "" {
		t.Errorf("observador nao deveria ter participantId, got %s", claims.ParticipantID)
	}
	if claims.Role != LiveRoleObserver {
		t.Errorf("role esperado observador, got %s", claims.Role)
	}
}

func TestLiveViewerTokenWrongSecret(t *testing.T) {
	token, err := NewLiveViewerToken("secret-a", 42, 100, LiveRoleParticipant, 6)
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}
	if _, err := VerifyLiveViewerToken("secret-b", token); err == nil {
		t.Error("token com secret errado deveria falhar")
	}
}

func TestLiveViewerTokenTampered(t *testing.T) {
	token, err := NewLiveViewerToken("secret-a", 42, 100, LiveRoleParticipant, 6)
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}
	if _, err := VerifyLiveViewerToken("secret-a", token+"x"); err == nil {
		t.Error("token adulterado deveria falhar")
	}
}
