package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/internal/database/sqlcgen"
	"github.com/jimmywiraarbaa/transport-api/schedulerule/domain"
)

type postgresRepository struct {
	db *pgxpool.Pool
	q  *sqlcgen.Queries
}

func New(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{db: db, q: sqlcgen.New(db)}
}

func (r *postgresRepository) List(ctx context.Context, limit, offset int32) ([]domain.ScheduleRule, error) {
	rows, err := r.q.ListScheduleRules(ctx, sqlcgen.ListScheduleRulesParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list schedule rules: %w", err)
	}
	out := make([]domain.ScheduleRule, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

func (r *postgresRepository) Count(ctx context.Context) (int64, error) {
	c, err := r.q.CountScheduleRules(ctx)
	if err != nil {
		return 0, fmt.Errorf("count schedule rules: %w", err)
	}
	return c, nil
}

func (r *postgresRepository) Get(ctx context.Context, id uuid.UUID) (*domain.ScheduleRule, error) {
	row, err := r.q.GetScheduleRule(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleRuleNotFound
		}
		return nil, fmt.Errorf("get schedule rule: %w", err)
	}
	s := toDomain(row)
	return &s, nil
}

func (r *postgresRepository) ListByVehicleType(ctx context.Context, vehicleTypeID uuid.UUID) ([]domain.ScheduleRule, error) {
	rows, err := r.q.ListScheduleRulesByVehicleType(ctx, vehicleTypeID)
	if err != nil {
		return nil, fmt.Errorf("list schedule rules by vehicle type: %w", err)
	}
	out := make([]domain.ScheduleRule, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

func (r *postgresRepository) Create(ctx context.Context, s domain.ScheduleRule) (*domain.ScheduleRule, error) {
	row, err := r.q.CreateScheduleRule(ctx, toCreateParams(s))
	if err != nil {
		return nil, fmt.Errorf("create schedule rule: %w", err)
	}
	out := toDomain(row)
	return &out, nil
}

func (r *postgresRepository) Update(ctx context.Context, s domain.ScheduleRule) (*domain.ScheduleRule, error) {
	row, err := r.q.UpdateScheduleRule(ctx, toUpdateParams(s))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleRuleNotFound
		}
		return nil, fmt.Errorf("update schedule rule: %w", err)
	}
	out := toDomain(row)
	return &out, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteScheduleRule(ctx, id); err != nil {
		return fmt.Errorf("delete schedule rule: %w", err)
	}
	return nil
}

func toDomain(s sqlcgen.ScheduleRule) domain.ScheduleRule {
	return domain.ScheduleRule{
		ID:            s.ID,
		PartID:        s.PartID,
		VehicleTypeID: s.VehicleTypeID,
		IntervalKm:    s.IntervalKm,
		IntervalDays:  s.IntervalDays,
		TriggerMode:   domain.TriggerMode(string(s.TriggerMode)),
		Notes:         s.Notes,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func toCreateParams(s domain.ScheduleRule) sqlcgen.CreateScheduleRuleParams {
	return sqlcgen.CreateScheduleRuleParams{
		PartID:        s.PartID,
		VehicleTypeID: s.VehicleTypeID,
		IntervalKm:    s.IntervalKm,
		IntervalDays:  s.IntervalDays,
		TriggerMode:   sqlcgen.TriggerMode(s.TriggerMode),
		Notes:         s.Notes,
	}
}

func toUpdateParams(s domain.ScheduleRule) sqlcgen.UpdateScheduleRuleParams {
	return sqlcgen.UpdateScheduleRuleParams{
		ID:            s.ID,
		PartID:        s.PartID,
		VehicleTypeID: s.VehicleTypeID,
		IntervalKm:    s.IntervalKm,
		IntervalDays:  s.IntervalDays,
		TriggerMode:   sqlcgen.TriggerMode(s.TriggerMode),
		Notes:         s.Notes,
	}
}
