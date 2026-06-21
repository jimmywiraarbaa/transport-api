package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MaintenanceRecord is a log of an actual maintenance/service performed.
type MaintenanceRecord struct {
	ID          uuid.UUID       `json:"id"`
	VehicleID   uuid.UUID       `json:"vehicle_id"`
	PartID      uuid.UUID       `json:"part_id"`
	UserID      uuid.UUID       `json:"user_id"`
	PerformedAt time.Time       `json:"performed_at"`
	OdometerKm  int32           `json:"odometer_km"`
	Cost        decimal.Decimal `json:"cost"`
	Technician  string          `json:"technician"`
	Notes       string          `json:"notes"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

var (
	ErrMaintenanceRecordNotFound = errors.New("maintenance record not found")
	ErrVehicleNotOwned           = errors.New("vehicle is not owned by the user")
)
