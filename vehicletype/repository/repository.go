package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/internal/database/sqlcgen"
	"github.com/jimmywiraarbaa/transport-api/vehicletype/domain"
)

type postgresRepository struct {
	db *pgxpool.Pool
	q  *sqlcgen.Queries
}

func New(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{db: db, q: sqlcgen.New(db)}
}

func (r *postgresRepository) List(ctx context.Context, limit, offset int32) ([]domain.VehicleType, error) {
	rows, err := r.q.ListVehicleTypes(ctx, sqlcgen.ListVehicleTypesParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list vehicle types: %w", err)
	}
	out := make([]domain.VehicleType, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

func (r *postgresRepository) Count(ctx context.Context) (int64, error) {
	c, err := r.q.CountVehicleTypes(ctx)
	if err != nil {
		return 0, fmt.Errorf("count vehicle types: %w", err)
	}
	return c, nil
}

func (r *postgresRepository) Get(ctx context.Context, id uuid.UUID) (*domain.VehicleType, error) {
	row, err := r.q.GetVehicleType(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVehicleTypeNotFound
		}
		return nil, fmt.Errorf("get vehicle type: %w", err)
	}
	vt := toDomain(row)
	return &vt, nil
}

func (r *postgresRepository) Create(ctx context.Context, name, slug string) (*domain.VehicleType, error) {
	row, err := r.q.CreateVehicleType(ctx, sqlcgen.CreateVehicleTypeParams{Name: name, Slug: slug})
	if err != nil {
		return nil, fmt.Errorf("create vehicle type: %w", err)
	}
	vt := toDomain(row)
	return &vt, nil
}

func (r *postgresRepository) Update(ctx context.Context, id uuid.UUID, name, slug string) (*domain.VehicleType, error) {
	row, err := r.q.UpdateVehicleType(ctx, sqlcgen.UpdateVehicleTypeParams{ID: id, Name: name, Slug: slug})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVehicleTypeNotFound
		}
		return nil, fmt.Errorf("update vehicle type: %w", err)
	}
	vt := toDomain(row)
	return &vt, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteVehicleType(ctx, id); err != nil {
		return fmt.Errorf("delete vehicle type: %w", err)
	}
	return nil
}

func toDomain(v sqlcgen.VehicleType) domain.VehicleType {
	return domain.VehicleType{
		ID:        v.ID,
		Name:      v.Name,
		Slug:      v.Slug,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}
