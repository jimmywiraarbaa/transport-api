package domain

import (
	"time"

	"github.com/google/uuid"
)

// VehicleType is a master category of vehicle (motor, mobil, truk, ...).
type VehicleType struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
