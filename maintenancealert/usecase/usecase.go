package usecase

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/maintenancealert/domain"
)

// Service is the alert computation interface consumed by the delivery layer.
type Service interface {
	Compute(ctx context.Context, vehicleID, userID uuid.UUID) (*domain.AlertSummary, error)
}

type service struct {
	repo domain.Repository
	now  func() time.Time
}

func New(repo domain.Repository) Service {
	return &service{repo: repo, now: time.Now}
}

// Compute returns the full alert summary for one vehicle.
func (s *service) Compute(ctx context.Context, vehicleID, userID uuid.UUID) (*domain.AlertSummary, error) {
	vehicle, err := s.repo.GetVehicle(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}

	rules, err := s.repo.ListRulesForVehicleType(ctx, vehicle.VehicleTypeID)
	if err != nil {
		return nil, err
	}

	records, err := s.repo.LatestRecordsPerPart(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}
	recordByPart := make(map[uuid.UUID]domain.RecordLite, len(records))
	for _, rec := range records {
		recordByPart[rec.PartID] = rec
	}

	now := s.now()
	alerts := make([]domain.Alert, 0, len(rules))
	for _, rule := range rules {
		alerts = append(alerts, s.computeAlert(rule, recordByPart, vehicle, now))
	}

	sortByUrgency(alerts)

	summary := &domain.AlertSummary{
		VehicleID:         vehicle.ID,
		VehicleTypeID:     vehicle.VehicleTypeID,
		CurrentOdometerKm: vehicle.CurrentOdometerKm,
		PlateNumber:       vehicle.PlateNumber,
		Total:             len(alerts),
		Alerts:            alerts,
	}
	for _, a := range alerts {
		switch a.Status {
		case domain.StatusOverdue:
			summary.Overdue++
		case domain.StatusDueNow:
			summary.DueNow++
		case domain.StatusDueSoon:
			summary.DueSoon++
		case domain.StatusOK:
			summary.OK++
		}
	}
	summary.HasUrgent = summary.Overdue > 0 || summary.DueNow > 0
	return summary, nil
}

// computeAlert derives the status for a single rule.
func (s *service) computeAlert(
	rule domain.RuleWithPart,
	records map[uuid.UUID]domain.RecordLite,
	vehicle domain.VehicleInfo,
	now time.Time,
) domain.Alert {
	alert := domain.Alert{
		VehicleID:      vehicle.ID,
		PartID:         rule.PartID,
		PartName:       rule.PartName,
		PartSlug:       rule.PartSlug,
		PartCategory:   rule.PartCategory,
		ScheduleRuleID: rule.ID,
		TriggerMode:    rule.TriggerMode,
		IntervalKm:     rule.IntervalKm,
		IntervalDays:   rule.IntervalDays,
		RuleNotes:      rule.Notes,
	}

	rec, hasRecord := records[rule.PartID]
	alert.HasRecord = hasRecord

	// When no maintenance record exists, the baseline is the vehicle's
	// initial odometer (captured at creation time) and creation date.
	// This means vehicle creation is the starting point for all sparepart
	// condition calculations.
	baselineKm := vehicle.InitialOdometerKm
	baselineDate := vehicle.CreatedAt
	if hasRecord {
		baselineKm = rec.OdometerKm
		baselineDate = rec.PerformedAt
		alert.LastOdometerKm = rec.OdometerKm
		alert.LastPerformedAt = &rec.PerformedAt
		c := rec.Cost
		alert.LastRecordCost = &c
	}

	var kmStatus, dateStatus *domain.Status

	if rule.IntervalKm != nil {
		nextKm := baselineKm + *rule.IntervalKm
		remaining := nextKm - vehicle.CurrentOdometerKm
		alert.NextKmDue = &nextKm
		alert.KmRemaining = &remaining
		s := kmStatusOf(remaining)
		kmStatus = &s
	}

	if rule.IntervalDays != nil {
		nextDate := baselineDate.AddDate(0, 0, int(*rule.IntervalDays))
		days := daysRemaining(now, nextDate)
		alert.NextDateDue = &nextDate
		alert.DaysRemaining = &days
		s := dateStatusOf(days)
		dateStatus = &s
	}

	alert.Status = combineStatuses(rule.TriggerMode, kmStatus, dateStatus)

	return alert
}

// ---- dimension helpers ----

func kmStatusOf(kmRemaining int32) domain.Status {
	switch {
	case kmRemaining < 0:
		return domain.StatusOverdue
	case kmRemaining <= domain.KmDueSoonThreshold:
		return domain.StatusDueSoon
	default:
		return domain.StatusOK
	}
}

func dateStatusOf(daysRemaining int) domain.Status {
	switch {
	case daysRemaining < 0:
		return domain.StatusOverdue
	case daysRemaining <= domain.DateDueSoonThreshold:
		return domain.StatusDueSoon
	default:
		return domain.StatusOK
	}
}

// combineStatuses merges the km and date dimensions based on trigger_mode.
func combineStatuses(triggerMode string, km, date *domain.Status) domain.Status {
	switch triggerMode {
	case "km_only":
		if km != nil {
			return *km
		}
		if date != nil {
			return *date
		}
	case "date_only":
		if date != nil {
			return *date
		}
		if km != nil {
			return *km
		}
	case "and":
		if km != nil && date != nil {
			return lessUrgent(*km, *date)
		}
	}
	// default / "or": most urgent dimension wins
	if km != nil && date != nil {
		return moreUrgent(*km, *date)
	}
	if km != nil {
		return *km
	}
	if date != nil {
		return *date
	}
	return domain.StatusOK
}

// urgencyRank maps a status to a numeric urgency (higher = more urgent).
func urgencyRank(s domain.Status) int {
	switch s {
	case domain.StatusOverdue:
		return 3
	case domain.StatusDueSoon:
		return 2
	case domain.StatusOK:
		return 1
	default:
		return 0
	}
}

func moreUrgent(a, b domain.Status) domain.Status {
	if urgencyRank(a) >= urgencyRank(b) {
		return a
	}
	return b
}

func lessUrgent(a, b domain.Status) domain.Status {
	if urgencyRank(a) <= urgencyRank(b) {
		return a
	}
	return b
}

// daysRemaining returns whole days until nextDate (negative if already passed).
func daysRemaining(now, nextDate time.Time) int {
	return int(math.Floor(nextDate.Sub(now).Hours() / 24))
}

// sortByUrgency orders alerts from most to least urgent.
func sortByUrgency(alerts []domain.Alert) {
	sort.SliceStable(alerts, func(i, j int) bool {
		return urgencyRank(alerts[i].Status) > urgencyRank(alerts[j].Status)
	})
}
