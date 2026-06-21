package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/internal/database/sqlcgen"
	"github.com/jimmywiraarbaa/transport-api/maintenancerecord/domain"
)

type postgresRepository struct {
	db *pgxpool.Pool
	q  *sqlcgen.Queries
}

func New(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{db: db, q: sqlcgen.New(db)}
}

func (r *postgresRepository) List(ctx context.Context, vehicleID, userID uuid.UUID, limit, offset int32) ([]domain.MaintenanceRecord, error) {
	rows, err := r.q.ListMaintenanceRecords(ctx, sqlcgen.ListMaintenanceRecordsParams{
		VehicleID: vehicleID, UserID: userID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list maintenance records: %w", err)
	}
	out := make([]domain.MaintenanceRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

func (r *postgresRepository) Count(ctx context.Context, vehicleID, userID uuid.UUID) (int64, error) {
	c, err := r.q.CountMaintenanceRecords(ctx, sqlcgen.CountMaintenanceRecordsParams{
		VehicleID: vehicleID, UserID: userID,
	})
	if err != nil {
		return 0, fmt.Errorf("count maintenance records: %w", err)
	}
	return c, nil
}

func (r *postgresRepository) ListAll(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, limit, offset int32) ([]domain.MaintenanceRecord, error) {
	var vid pgtype.UUID
	if vehicleID != nil {
		vid = pgtype.UUID{Bytes: *vehicleID, Valid: true}
	}
	rows, err := r.q.ListAllMaintenanceRecords(ctx, sqlcgen.ListAllMaintenanceRecordsParams{
		UserID:    userID,
		Limit:     limit,
		Offset:    offset,
		VehicleID: vid,
	})
	if err != nil {
		return nil, fmt.Errorf("list all maintenance records: %w", err)
	}
	out := make([]domain.MaintenanceRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

func (r *postgresRepository) CountAll(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID) (int64, error) {
	var vid pgtype.UUID
	if vehicleID != nil {
		vid = pgtype.UUID{Bytes: *vehicleID, Valid: true}
	}
	c, err := r.q.CountAllMaintenanceRecords(ctx, sqlcgen.CountAllMaintenanceRecordsParams{
		UserID:    userID,
		VehicleID: vid,
	})
	if err != nil {
		return 0, fmt.Errorf("count all maintenance records: %w", err)
	}
	return c, nil
}

func (r *postgresRepository) Get(ctx context.Context, id, userID uuid.UUID) (*domain.MaintenanceRecord, error) {
	row, err := r.q.GetMaintenanceRecord(ctx, sqlcgen.GetMaintenanceRecordParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMaintenanceRecordNotFound
		}
		return nil, fmt.Errorf("get maintenance record: %w", err)
	}
	rec := toDomain(row)
	return &rec, nil
}

// Create inserts a record and, in the same transaction, advances the owning
// vehicle's odometer when the service reading is higher. Ownership is verified
// inside the tx (the vehicle must belong to the user).
func (r *postgresRepository) Create(ctx context.Context, rec domain.MaintenanceRecord) (*domain.MaintenanceRecord, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := sqlcgen.New(tx)

	// Verify ownership (vehicle must exist and belong to the user).
	if _, err := q.GetVehicle(ctx, sqlcgen.GetVehicleParams{ID: rec.VehicleID, UserID: rec.UserID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVehicleNotOwned
		}
		return nil, fmt.Errorf("verify vehicle ownership: %w", err)
	}

	row, err := q.CreateMaintenanceRecord(ctx, sqlcgen.CreateMaintenanceRecordParams{
		VehicleID:   rec.VehicleID,
		PartID:      rec.PartID,
		UserID:      rec.UserID,
		PerformedAt: rec.PerformedAt,
		OdometerKm:  rec.OdometerKm,
		Cost:        rec.Cost,
		Technician:  rec.Technician,
		Notes:       rec.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("create maintenance record: %w", err)
	}

	// Advance the vehicle's odometer if this service reading is higher.
	// ErrNoRows here just means the reading was not higher — safe to ignore.
	if _, err := q.UpdateVehicleOdometerIfHigher(ctx, sqlcgen.UpdateVehicleOdometerIfHigherParams{
		ID: rec.VehicleID, UserID: rec.UserID, CurrentOdometerKm: rec.OdometerKm,
	}); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("update vehicle odometer: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	out := toDomain(row)
	return &out, nil
}

func (r *postgresRepository) Update(ctx context.Context, rec domain.MaintenanceRecord) (*domain.MaintenanceRecord, error) {
	row, err := r.q.UpdateMaintenanceRecord(ctx, sqlcgen.UpdateMaintenanceRecordParams{
		ID: rec.ID, UserID: rec.UserID, PartID: rec.PartID,
		PerformedAt: rec.PerformedAt, OdometerKm: rec.OdometerKm, Cost: rec.Cost,
		Technician: rec.Technician, Notes: rec.Notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMaintenanceRecordNotFound
		}
		return nil, fmt.Errorf("update maintenance record: %w", err)
	}
	out := toDomain(row)
	return &out, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.DeleteMaintenanceRecord(ctx, sqlcgen.DeleteMaintenanceRecordParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("delete maintenance record: %w", err)
	}
	return nil
}

// LatestPerPart returns the most recent record for each part on a vehicle.
func (r *postgresRepository) LatestPerPart(ctx context.Context, vehicleID, userID uuid.UUID) ([]domain.MaintenanceRecord, error) {
	rows, err := r.q.LatestRecordPerPart(ctx, sqlcgen.LatestRecordPerPartParams{VehicleID: vehicleID, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("latest record per part: %w", err)
	}
	out := make([]domain.MaintenanceRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.MaintenanceRecord{
			ID:          row.ID,
			VehicleID:   row.VehicleID,
			PartID:      row.PartID,
			UserID:      row.UserID,
			PerformedAt: row.PerformedAt,
			OdometerKm:  row.OdometerKm,
			Cost:        row.Cost,
			Technician:  row.Technician,
			Notes:       row.Notes,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return out, nil
}

func toDomain(r sqlcgen.MaintenanceRecord) domain.MaintenanceRecord {
	return domain.MaintenanceRecord{
		ID:          r.ID,
		VehicleID:   r.VehicleID,
		PartID:      r.PartID,
		UserID:      r.UserID,
		PerformedAt: r.PerformedAt,
		OdometerKm:  r.OdometerKm,
		Cost:        r.Cost,
		Technician:  r.Technician,
		Notes:       r.Notes,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
