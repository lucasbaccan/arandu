package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Papéis de quem entra na apresentação pública: participante é quem já
// respondeu o evento (e-mail bate com um participante existente);
// observador tem o PIN certo mas não é ninguém que respondeu.
const (
	LiveRoleParticipant = "participante"
	LiveRoleObserver    = "observador"
)

type LiveClaims struct {
	EventID       string `json:"eventId"`
	ParticipantID string `json:"participantId,omitempty"`
	Role          string `json:"role"`
	jwt.RegisteredClaims
}

// NewLiveViewerToken emite o token de quem entrou na apresentação pública
// (via PIN + e-mail). participantID é 0 pra observador.
func NewLiveViewerToken(secret string, eventID, participantID int64, role string, hours int) (string, error) {
	var pid string
	if participantID != 0 {
		pid = strconv.FormatInt(participantID, 10)
	}
	claims := LiveClaims{
		EventID:       strconv.FormatInt(eventID, 10),
		ParticipantID: pid,
		Role:          role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(hours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyLiveViewerToken(secret, tokenString string) (LiveClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &LiveClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: método de assinatura inesperado: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return LiveClaims{}, fmt.Errorf("auth: token de visitante inválido")
	}
	claims, ok := token.Claims.(*LiveClaims)
	if !ok {
		return LiveClaims{}, fmt.Errorf("auth: claims inválidos")
	}
	return *claims, nil
}
