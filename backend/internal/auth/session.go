package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SessionCookieName = "session"

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

func NewSessionToken(secret string, userID int64, hours int) (string, error) {
	claims := Claims{
		UserID: strconv.FormatInt(userID, 10),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(hours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.FormatInt(userID, 10),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifySessionToken(secret, tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: método de assinatura inesperado: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("auth: token inválido")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, fmt.Errorf("auth: claims inválidos")
	}
	userID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("auth: user id inválido")
	}
	return userID, nil
}
