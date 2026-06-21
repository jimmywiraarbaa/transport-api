package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Vehicle is a vehicle owned by a user.
type Vehicle struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	VehicleTypeID     uuid.UUID `json:"vehicle_type_id"`
	PlateNumber       string    `json:"plate_number"`
	Brand             string    `json:"brand"`
	Model             string    `json:"model"`
	Year              *int32    `json:"year"`
	CurrentOdometerKm int32     `json:"current_odometer_km"`
	InitialOdometerKm int32     `json:"initial_odometer_km"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

var (
	ErrVehicleNotFound = errors.New("vehicle not found")
)
