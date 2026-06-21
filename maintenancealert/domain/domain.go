package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Status describes the maintenance urgency of a single part.
type Status string

const (
	StatusOK      Status = "ok"       // healthy — comfortably within both thresholds
	StatusDueSoon Status = "due_soon" // approaching a threshold (<=500 km or <=14 days)
	StatusOverdue Status = "overdue"  // a threshold has passed — service immediately
	StatusDueNow  Status = "due_now"  // no maintenance record exists — establish a baseline
)

// Due-soon warning windows.
const (
	KmDueSoonThreshold   int32 = 500 // km remaining <= this → due_soon
	DateDueSoonThreshold int   = 14  // days remaining <= this → due_soon
)

// Alert is the computed maintenance status for one part on one vehicle.
type Alert struct {
	VehicleID uuid.UUID `json:"vehicle_id"`

	PartID       uuid.UUID `json:"part_id"`
	PartName     string    `json:"part_name"`
	PartSlug     string    `json:"part_slug"`
	PartCategory string    `json:"part_category"`

	ScheduleRuleID uuid.UUID `json:"schedule_rule_id"`
	TriggerMode    string    `json:"trigger_mode"`
	IntervalKm     *int32    `json:"interval_km"`
	IntervalDays   *int32    `json:"interval_days"`

	Status Status `json:"status"`

	// KM dimension.
	LastOdometerKm int32  `json:"last_odometer_km"`
	NextKmDue      *int32 `json:"next_km_due"`
	KmRemaining    *int32 `json:"km_remaining"`

	// Date dimension.
	LastPerformedAt *time.Time `json:"last_performed_at"`
	NextDateDue     *time.Time `json:"next_date_due"`
	DaysRemaining   *int       `json:"days_remaining"`

	// Record metadata.
	HasRecord      bool             `json:"has_record"`
	LastRecordCost *decimal.Decimal `json:"last_record_cost"`
	RuleNotes      string           `json:"rule_notes"`
}

// AlertSummary aggregates per-status counts alongside the detailed alerts.
type AlertSummary struct {
	VehicleID           uuid.UUID `json:"vehicle_id"`
	VehicleTypeID       uuid.UUID `json:"vehicle_type_id"`
	CurrentOdometerKm   int32     `json:"current_odometer_km"`
	PlateNumber         string    `json:"plate_number"`
	Total               int       `json:"total"`
	Overdue             int       `json:"overdue"`
	DueNow              int       `json:"due_now"`
	DueSoon             int       `json:"due_soon"`
	OK                  int       `json:"ok"`
	HasUrgent           bool      `json:"has_urgent"` // overdue || due_now > 0
	Alerts              []Alert   `json:"alerts"`
}

// ---- Read-model inputs (repository → usecase) ----

// VehicleInfo is the minimal vehicle data needed for computation.
type VehicleInfo struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	VehicleTypeID     uuid.UUID
	PlateNumber       string
	CurrentOdometerKm int32
	InitialOdometerKm int32
	CreatedAt         time.Time
}

// RuleWithPart is a schedule rule enriched with its maintenance part.
type RuleWithPart struct {
	ID            uuid.UUID
	PartID        uuid.UUID
	VehicleTypeID uuid.UUID
	IntervalKm    *int32
	IntervalDays  *int32
	TriggerMode   string
	Notes         string
	PartName      string
	PartSlug      string
	PartCategory  string
}

// RecordLite is the latest maintenance record for a (vehicle, part) pair.
type RecordLite struct {
	PartID      uuid.UUID
	PerformedAt time.Time
	OdometerKm  int32
	Cost        decimal.Decimal
	Notes       string
}

var (
	ErrVehicleNotFound = errors.New("vehicle not found")
)

// Repository is the read-model interface for alert computation.
type Repository interface {
	GetVehicle(ctx context.Context, vehicleID, userID uuid.UUID) (VehicleInfo, error)
	ListRulesForVehicleType(ctx context.Context, vehicleTypeID uuid.UUID) ([]RuleWithPart, error)
	LatestRecordsPerPart(ctx context.Context, vehicleID, userID uuid.UUID) ([]RecordLite, error)
}
