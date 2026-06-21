package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/internal/database/sqlcgen"
	"github.com/jimmywiraarbaa/transport-api/maintenancepart/domain"
)

type postgresRepository struct {
	db *pgxpool.Pool
	q  *sqlcgen.Queries
}

func New(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{db: db, q: sqlcgen.New(db)}
}

func (r *postgresRepository) List(ctx context.Context, limit, offset int32) ([]domain.MaintenancePart, error) {
	rows, err := r.q.ListMaintenanceParts(ctx, sqlcgen.ListMaintenancePartsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list maintenance parts: %w", err)
	}
	out := make([]domain.MaintenancePart, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

func (r *postgresRepository) Count(ctx context.Context) (int64, error) {
	c, err := r.q.CountMaintenanceParts(ctx)
	if err != nil {
		return 0, fmt.Errorf("count maintenance parts: %w", err)
	}
	return c, nil
}

func (r *postgresRepository) Get(ctx context.Context, id uuid.UUID) (*domain.MaintenancePart, error) {
	row, err := r.q.GetMaintenancePart(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMaintenancePartNotFound
		}
		return nil, fmt.Errorf("get maintenance part: %w", err)
	}
	p := toDomain(row)
	return &p, nil
}

func (r *postgresRepository) Create(ctx context.Context, p domain.MaintenancePart) (*domain.MaintenancePart, error) {
	row, err := r.q.CreateMaintenancePart(ctx, sqlcgen.CreateMaintenancePartParams{
		Name: p.Name, Slug: p.Slug, Category: p.Category, Description: p.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("create maintenance part: %w", err)
	}
	out := toDomain(row)
	return &out, nil
}

func (r *postgresRepository) Update(ctx context.Context, p domain.MaintenancePart) (*domain.MaintenancePart, error) {
	row, err := r.q.UpdateMaintenancePart(ctx, sqlcgen.UpdateMaintenancePartParams{
		ID: p.ID, Name: p.Name, Slug: p.Slug, Category: p.Category, Description: p.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMaintenancePartNotFound
		}
		return nil, fmt.Errorf("update maintenance part: %w", err)
	}
	out := toDomain(row)
	return &out, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteMaintenancePart(ctx, id); err != nil {
		return fmt.Errorf("delete maintenance part: %w", err)
	}
	return nil
}

func toDomain(p sqlcgen.MaintenancePart) domain.MaintenancePart {
	return domain.MaintenancePart{
		ID:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Category:    p.Category,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
