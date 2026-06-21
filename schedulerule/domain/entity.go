package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TriggerMode controls how km and date intervals combine for a schedule rule.
type TriggerMode string

const (
	// TriggerModeOr fires when EITHER the km or the date interval is reached
	// (whichever comes first). Default for vehicle maintenance.
	TriggerModeOr TriggerMode = "or"
	// TriggerModeAnd fires only when BOTH intervals are reached.
	TriggerModeAnd TriggerMode = "and"
	// TriggerModeKMOnly ignores the date interval entirely.
	TriggerModeKMOnly TriggerMode = "km_only"
	// TriggerModeDateOnly ignores the km interval entirely.
	TriggerModeDateOnly TriggerMode = "date_only"
)

// ValidTriggerModes is the set of accepted trigger modes.
var ValidTriggerModes = []TriggerMode{
	TriggerModeOr, TriggerModeAnd, TriggerModeKMOnly, TriggerModeDateOnly,
}

// IsValid reports whether m is a recognized trigger mode.
func (m TriggerMode) IsValid() bool {
	for _, v := range ValidTriggerModes {
		if m == v {
			return true
		}
	}
	return false
}

// ScheduleRule defines, per (part, vehicle_type), the replacement interval.
type ScheduleRule struct {
	ID            uuid.UUID   `json:"id"`
	PartID        uuid.UUID   `json:"part_id"`
	VehicleTypeID uuid.UUID   `json:"vehicle_type_id"`
	IntervalKm    *int32      `json:"interval_km"`
	IntervalDays  *int32      `json:"interval_days"`
	TriggerMode   TriggerMode `json:"trigger_mode"`
	Notes         string      `json:"notes"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

var (
	ErrScheduleRuleNotFound = errors.New("schedule rule not found")
)
