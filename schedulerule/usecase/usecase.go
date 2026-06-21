package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/schedulerule/domain"
)

// ErrValidation signals an invalid input (maps to HTTP 422).
var ErrValidation = errors.New("validation error")

// Repository is the persistence port the schedulerule usecase needs.
type Repository interface {
	List(ctx context.Context, limit, offset int32) ([]domain.ScheduleRule, error)
	Count(ctx context.Context) (int64, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.ScheduleRule, error)
	ListByVehicleType(ctx context.Context, vehicleTypeID uuid.UUID) ([]domain.ScheduleRule, error)
	Create(ctx context.Context, s domain.ScheduleRule) (*domain.ScheduleRule, error)
	Update(ctx context.Context, s domain.ScheduleRule) (*domain.ScheduleRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ListInput holds list/pagination params.
type ListInput struct {
	Page    int
	PerPage int
}

// UpsertInput holds create/update fields.
type UpsertInput struct {
	PartID        uuid.UUID
	VehicleTypeID uuid.UUID
	IntervalKm    *int32
	IntervalDays  *int32
	TriggerMode   domain.TriggerMode
	Notes         string
}

// ListResult holds a page of items with total count.
type ListResult struct {
	Items []domain.ScheduleRule
	Total int64
}

type Service interface {
	List(ctx context.Context, in ListInput) (*ListResult, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.ScheduleRule, error)
	ListByVehicleType(ctx context.Context, vehicleTypeID uuid.UUID) ([]domain.ScheduleRule, error)
	Create(ctx context.Context, in UpsertInput) (*domain.ScheduleRule, error)
	Update(ctx context.Context, id uuid.UUID, in UpsertInput) (*domain.ScheduleRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func New(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, in ListInput) (*ListResult, error) {
	page, perPage := normalizePaging(in.Page, in.PerPage)
	offset := int32((page - 1) * perPage)

	items, err := s.repo.List(ctx, int32(perPage), offset)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total}, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*domain.ScheduleRule, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) ListByVehicleType(ctx context.Context, vehicleTypeID uuid.UUID) ([]domain.ScheduleRule, error) {
	return s.repo.ListByVehicleType(ctx, vehicleTypeID)
}

func (s *service) Create(ctx context.Context, in UpsertInput) (*domain.ScheduleRule, error) {
	if err := validateUpsert(in); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return s.repo.Create(ctx, fromInput(in, uuid.Nil))
}

func (s *service) Update(ctx context.Context, id uuid.UUID, in UpsertInput) (*domain.ScheduleRule, error) {
	if err := validateUpsert(in); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return s.repo.Update(ctx, fromInput(in, id))
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func validateUpsert(in UpsertInput) error {
	if in.PartID == uuid.Nil {
		return errors.New("part_id is required")
	}
	if in.VehicleTypeID == uuid.Nil {
		return errors.New("vehicle_type_id is required")
	}
	if in.IntervalKm == nil && in.IntervalDays == nil {
		return errors.New("at least one of interval_km or interval_days is required")
	}
	if in.IntervalKm != nil && *in.IntervalKm <= 0 {
		return errors.New("interval_km must be greater than 0")
	}
	if in.IntervalDays != nil && *in.IntervalDays <= 0 {
		return errors.New("interval_days must be greater than 0")
	}
	mode := in.TriggerMode
	if mode == "" {
		mode = domain.TriggerModeOr // default applied in fromInput too
	}
	if !mode.IsValid() {
		return errors.New("invalid trigger_mode")
	}
	return nil
}

func fromInput(in UpsertInput, id uuid.UUID) domain.ScheduleRule {
	mode := in.TriggerMode
	if mode == "" {
		mode = domain.TriggerModeOr
	}
	return domain.ScheduleRule{
		ID:            id,
		PartID:        in.PartID,
		VehicleTypeID: in.VehicleTypeID,
		IntervalKm:    in.IntervalKm,
		IntervalDays:  in.IntervalDays,
		TriggerMode:   mode,
		Notes:         strings.TrimSpace(in.Notes),
	}
}

func normalizePaging(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}
