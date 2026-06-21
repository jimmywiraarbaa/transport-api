package domain

import (
	"time"

	"github.com/google/uuid"
)

// MaintenancePart is a master catalog entry for a serviceable component
// (oli mesin, filter udara, kampas rem, ...).
type MaintenancePart struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
