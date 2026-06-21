// Package usecase contains the auth feature's business rules.
// It declares the Repository port it needs (consumer-side interface, per
// bxcodec/go-clean-arch v4) and a Service interface consumed by delivery.
package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/auth/domain"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors for the auth usecase.
var (
	ErrInvalidCredential = errors.New("invalid username or password")
	ErrInvalidToken      = errors.New("invalid refresh token")
	ErrUserNotFound      = errors.New("user not found")
)

// Repository is the persistence port the auth usecase depends on.
type Repository interface {
	Create(ctx context.Context, email, username, passwordHash, name string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

// LoginInput is the data needed to authenticate.
type LoginInput struct {
	Username string
	Password string
}

// AuthResult is returned on successful authentication.
type AuthResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Service is the auth usecase contract consumed by the delivery layer.
type Service interface {
	Login(ctx context.Context, in LoginInput) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	Current(ctx context.Context, userID uuid.UUID) (*domain.User, error)
}

type service struct {
	repo Repository
	jwt  *jwt.Manager
}

// New wires the auth usecase with its repository and jwt manager.
func New(repo Repository, mgr *jwt.Manager) Service {
	return &service{repo: repo, jwt: mgr}
}

// Login verifies credentials and returns a token pair.
func (s *service) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)

	user, err := s.repo.GetByUsername(ctx, in.Username)
	if err != nil {
		return nil, ErrInvalidCredential
	}
	if !comparePassword(user.PasswordHash, in.Password) {
		return nil, ErrInvalidCredential
	}
	return s.issueTokens(user)
}

// Refresh validates a refresh token and issues a new pair.
func (s *service) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	claims, err := s.jwt.Verify(jwt.RefreshToken, refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return s.issueTokens(user)
}

// Current returns the user for a given id (used by the /me endpoint).
func (s *service) Current(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// issueTokens generates an access+refresh pair for the user.
func (s *service) issueTokens(user *domain.User) (*AuthResult, error) {
	access, expiresAt, err := s.jwt.Generate(jwt.AccessToken, user.ID.String(), user.Username)
	if err != nil {
		return nil, err
	}
	refresh, _, err := s.jwt.Generate(jwt.RefreshToken, user.ID.String(), user.Username)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		User:         user,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt,
	}, nil
}

func comparePassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
