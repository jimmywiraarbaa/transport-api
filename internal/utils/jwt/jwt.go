// Package jwt handles access & refresh token generation and verification.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jimmywiraarbaa/transport-api/internal/config"
)

// TokenType distinguishes access tokens from refresh tokens.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Errors returned by this package.
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims embedded in every issued token.
type Claims struct {
	UserID string    `json:"user_id"`
	Email  string    `json:"email"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

// Manager signs and verifies JWTs using the secrets from config.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// New builds a Manager from app configuration.
func New(cfg config.JWTConfig) *Manager {
	return &Manager{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

// Generate creates a signed token of the given type for a user.
func (m *Manager) Generate(t TokenType, userID, email string) (string, time.Time, error) {
	var secret []byte
	var ttl time.Duration
	switch t {
	case AccessToken:
		secret, ttl = m.accessSecret, m.accessTTL
	case RefreshToken:
		secret, ttl = m.refreshSecret, m.refreshTTL
	default:
		return "", time.Time{}, fmt.Errorf("unknown token type: %s", t)
	}

	expiresAt := time.Now().Add(ttl)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Type:   t,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

// Verify validates a token string and returns its claims.
func (m *Manager) Verify(t TokenType, tokenStr string) (*Claims, error) {
	var secret []byte
	switch t {
	case AccessToken:
		secret = m.accessSecret
	case RefreshToken:
		secret = m.refreshSecret
	default:
		return nil, fmt.Errorf("unknown token type: %s", t)
	}

	var claims Claims
	parser := jwt.Parser{}
	tok, err := parser.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if !tok.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != t {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}
