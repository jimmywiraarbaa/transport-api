// Package domain holds the auth feature's entity and value objects.
// It has zero dependencies on infrastructure (no db, no http).
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the authenticated account that owns vehicles.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // never serialized
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
