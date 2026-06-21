package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/internal/database/sqlcgen"
	"github.com/jimmywiraarbaa/transport-api/maintenancealert/domain"
)

type postgresRepository struct {
	db *pgxpool.Pool
	q  *sqlcgen.Queries
}

func New(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{db: db, q: sqlcgen.New(db)}
}

func (r *postgresRepository) GetVehicle(ctx context.Context, vehicleID, userID uuid.UUID) (domain.VehicleInfo, error) {
	v, err := r.q.GetVehicle(ctx, sqlcgen.GetVehicleParams{ID: vehicleID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VehicleInfo{}, domain.ErrVehicleNotFound
		}
		return domain.VehicleInfo{}, fmt.Errorf("get vehicle: %w", err)
	}
	return domain.VehicleInfo{
		ID:                v.ID,
		UserID:            v.UserID,
		VehicleTypeID:     v.VehicleTypeID,
		PlateNumber:       v.PlateNumber,
		CurrentOdometerKm: v.CurrentOdometerKm,
		InitialOdometerKm: v.InitialOdometerKm,
		CreatedAt:         v.CreatedAt,
	}, nil
}

func (r *postgresRepository) ListRulesForVehicleType(ctx context.Context, vehicleTypeID uuid.UUID) ([]domain.RuleWithPart, error) {
	rows, err := r.q.ListRulesForVehicleType(ctx, vehicleTypeID)
	if err != nil {
		return nil, fmt.Errorf("list rules for vehicle type: %w", err)
	}
	out := make([]domain.RuleWithPart, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.RuleWithPart{
			ID:            row.ID,
			PartID:        row.PartID,
			VehicleTypeID: row.VehicleTypeID,
			IntervalKm:    row.IntervalKm,
			IntervalDays:  row.IntervalDays,
			TriggerMode:   string(row.TriggerMode),
			Notes:         row.Notes,
			PartName:      row.PartName,
			PartSlug:      row.PartSlug,
			PartCategory:  row.PartCategory,
		})
	}
	return out, nil
}

func (r *postgresRepository) LatestRecordsPerPart(ctx context.Context, vehicleID, userID uuid.UUID) ([]domain.RecordLite, error) {
	rows, err := r.q.LatestRecordPerPart(ctx, sqlcgen.LatestRecordPerPartParams{
		VehicleID: vehicleID,
		UserID:    userID,
	})
	if err != nil {
		return nil, fmt.Errorf("latest records per part: %w", err)
	}
	out := make([]domain.RecordLite, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.RecordLite{
			PartID:      row.PartID,
			PerformedAt: row.PerformedAt,
			OdometerKm:  row.OdometerKm,
			Cost:        row.Cost,
			Notes:       row.Notes,
		})
	}
	return out, nil
}
