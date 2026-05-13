package services

import (
	"errors"
	"time"

	"github.com/devnolife/umkm-api/internal/config"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string          `json:"userId"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func SignToken(userID string, role models.UserRole) (string, error) {
	cfg := config.Get()
	exp := time.Now().Add(time.Duration(cfg.JWTExpiresHours) * time.Hour)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(cfg.JWTSecret))
}

func ParseToken(tokenStr string) (*Claims, error) {
	cfg := config.Get()
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := tok.Claims.(*Claims); ok && tok.Valid {
		return c, nil
	}
	return nil, errors.New("invalid token")
}
