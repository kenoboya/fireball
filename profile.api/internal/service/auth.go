package service

import (
	"profile-api/pkg/auth"
)

type AuthService struct {
	tokenManeger auth.TokenManager
}

func NewAuthService(tkManager auth.TokenManager) *AuthService {
	return &AuthService{
		tokenManeger: tkManager,
	}
}

func (s *AuthService) ValidateToken(token string) (string, error) {
	claims, err := s.tokenManeger.ParseToken(token)
	if err != nil {
		return "", err
	}
	return claims.UserID, err
}
