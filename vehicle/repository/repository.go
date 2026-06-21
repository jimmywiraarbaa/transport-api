package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/internal/database/sqlcgen"
	"github.com/jimmywiraarbaa/transport-api/vehicle/domain"
)

type postgresRepository struct {
	db *pgxpool.Pool
	q  *sqlcgen.Queries
}

func New(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{db: db, q: sqlcgen.New(db)}
}

func (r *postgresRepository) List(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.Vehicle, error) {
	rows, err := r.q.ListVehicles(ctx, sqlcgen.ListVehiclesParams{UserID: userID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list vehicles: %w", err)
	}
	out := make([]domain.Vehicle, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

func (r *postgresRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	c, err := r.q.CountVehicles(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count vehicles: %w", err)
	}
	return c, nil
}

func (r *postgresRepository) Get(ctx context.Context, id, userID uuid.UUID) (*domain.Vehicle, error) {
	row, err := r.q.GetVehicle(ctx, sqlcgen.GetVehicleParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle: %w", err)
	}
	v := toDomain(row)
	return &v, nil
}

func (r *postgresRepository) Create(ctx context.Context, v domain.Vehicle) (*domain.Vehicle, error) {
	row, err := r.q.CreateVehicle(ctx, sqlcgen.CreateVehicleParams{
		UserID: v.UserID, VehicleTypeID: v.VehicleTypeID, PlateNumber: v.PlateNumber,
		Brand: v.Brand, Model: v.Model, Year: v.Year,
		CurrentOdometerKm: v.CurrentOdometerKm, InitialOdometerKm: v.InitialOdometerKm, Notes: v.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("create vehicle: %w", err)
	}
	out := toDomain(row)
	return &out, nil
}

func (r *postgresRepository) Update(ctx context.Context, v domain.Vehicle) (*domain.Vehicle, error) {
	row, err := r.q.UpdateVehicle(ctx, sqlcgen.UpdateVehicleParams{
		ID: v.ID, UserID: v.UserID, VehicleTypeID: v.VehicleTypeID, PlateNumber: v.PlateNumber,
		Brand: v.Brand, Model: v.Model, Year: v.Year, Notes: v.Notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVehicleNotFound
		}
		return nil, fmt.Errorf("update vehicle: %w", err)
	}
	out := toDomain(row)
	return &out, nil
}

// UpdateOdometerIfHigher advances the odometer only when km exceeds the current value.
// Returns ErrVehicleNotFound when no row is updated (vehicle missing or km not higher).
func (r *postgresRepository) UpdateOdometerIfHigher(ctx context.Context, id, userID uuid.UUID, km int32) (*domain.Vehicle, error) {
	row, err := r.q.UpdateVehicleOdometerIfHigher(ctx, sqlcgen.UpdateVehicleOdometerIfHigherParams{
		ID: id, UserID: userID, CurrentOdometerKm: km,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVehicleNotFound
		}
		return nil, fmt.Errorf("update odometer: %w", err)
	}
	v := toDomain(row)
	return &v, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.DeleteVehicle(ctx, sqlcgen.DeleteVehicleParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("delete vehicle: %w", err)
	}
	return nil
}

func toDomain(v sqlcgen.Vehicle) domain.Vehicle {
	return domain.Vehicle{
		ID:                v.ID,
		UserID:            v.UserID,
		VehicleTypeID:     v.VehicleTypeID,
		PlateNumber:       v.PlateNumber,
		Brand:             v.Brand,
		Model:             v.Model,
		Year:              v.Year,
		CurrentOdometerKm: v.CurrentOdometerKm,
		InitialOdometerKm: v.InitialOdometerKm,
		Notes:             v.Notes,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}
}
