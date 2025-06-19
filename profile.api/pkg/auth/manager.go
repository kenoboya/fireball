package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrSecretKeyIsEmpty = errors.New("secret key is empty")
	ErrUnauthorized     = errors.New("token is invalid")
)

type TokenManager interface {
	ParseToken(token string) (*Claims, error)
}

type Manager struct {
	secretAccessKey string
}

type Claims struct {
	UserID string `json:"id"`
	jwt.RegisteredClaims
}

func NewManager(secretAccessKey string) (*Manager, error) {
	if secretAccessKey == "" {
		return nil, ErrSecretKeyIsEmpty
	}
	return &Manager{
		secretAccessKey: secretAccessKey,
	}, nil
}

func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secretKey := []byte(m.secretAccessKey)
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrUnauthorized
}
