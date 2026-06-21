package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/maintenancerecord/domain"
	"github.com/shopspring/decimal"
)

// ErrValidation signals an invalid input (maps to HTTP 422).
var ErrValidation = errors.New("validation error")

// Repository is the persistence port the maintenancerecord usecase needs.
type Repository interface {
	List(ctx context.Context, vehicleID, userID uuid.UUID, limit, offset int32) ([]domain.MaintenanceRecord, error)
	Count(ctx context.Context, vehicleID, userID uuid.UUID) (int64, error)
	ListAll(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, limit, offset int32) ([]domain.MaintenanceRecord, error)
	CountAll(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID) (int64, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*domain.MaintenanceRecord, error)
	Create(ctx context.Context, rec domain.MaintenanceRecord) (*domain.MaintenanceRecord, error)
	Update(ctx context.Context, rec domain.MaintenanceRecord) (*domain.MaintenanceRecord, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	LatestPerPart(ctx context.Context, vehicleID, userID uuid.UUID) ([]domain.MaintenanceRecord, error)
}

// ListInput holds list/pagination params.
type ListInput struct {
	Page    int
	PerPage int
}

// UpsertInput holds create/update fields.
type UpsertInput struct {
	VehicleID   uuid.UUID
	PartID      uuid.UUID
	PerformedAt time.Time
	OdometerKm  int32
	Cost        decimal.Decimal
	Technician  string
	Notes       string
}

// ListResult holds a page of items with total count.
type ListResult struct {
	Items []domain.MaintenanceRecord
	Total int64
}

type Service interface {
	List(ctx context.Context, userID, vehicleID uuid.UUID, in ListInput) (*ListResult, error)
	ListAll(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, in ListInput) (*ListResult, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*domain.MaintenanceRecord, error)
	Create(ctx context.Context, userID uuid.UUID, in UpsertInput) (*domain.MaintenanceRecord, error)
	Update(ctx context.Context, id, userID uuid.UUID, in UpsertInput) (*domain.MaintenanceRecord, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type service struct {
	repo Repository
}

func New(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, userID, vehicleID uuid.UUID, in ListInput) (*ListResult, error) {
	page, perPage := normalizePaging(in.Page, in.PerPage)
	offset := int32((page - 1) * perPage)

	items, err := s.repo.List(ctx, vehicleID, userID, int32(perPage), offset)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.Count(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total}, nil
}

func (s *service) ListAll(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, in ListInput) (*ListResult, error) {
	page, perPage := normalizePaging(in.Page, in.PerPage)
	offset := int32((page - 1) * perPage)

	items, err := s.repo.ListAll(ctx, userID, vehicleID, int32(perPage), offset)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.CountAll(ctx, userID, vehicleID)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total}, nil
}

func (s *service) Get(ctx context.Context, id, userID uuid.UUID) (*domain.MaintenanceRecord, error) {
	return s.repo.Get(ctx, id, userID)
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, in UpsertInput) (*domain.MaintenanceRecord, error) {
	if err := validateUpsert(in); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	rec := domain.MaintenanceRecord{
		VehicleID:   in.VehicleID,
		PartID:      in.PartID,
		UserID:      userID,
		PerformedAt: in.PerformedAt,
		OdometerKm:  in.OdometerKm,
		Cost:        in.Cost,
		Technician:  strings.TrimSpace(in.Technician),
		Notes:       strings.TrimSpace(in.Notes),
	}
	return s.repo.Create(ctx, rec)
}

func (s *service) Update(ctx context.Context, id, userID uuid.UUID, in UpsertInput) (*domain.MaintenanceRecord, error) {
	if err := validateUpsert(in); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	rec := domain.MaintenanceRecord{
		ID:          id,
		VehicleID:   in.VehicleID,
		PartID:      in.PartID,
		UserID:      userID,
		PerformedAt: in.PerformedAt,
		OdometerKm:  in.OdometerKm,
		Cost:        in.Cost,
		Technician:  strings.TrimSpace(in.Technician),
		Notes:       strings.TrimSpace(in.Notes),
	}
	return s.repo.Update(ctx, rec)
}

func (s *service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

func validateUpsert(in UpsertInput) error {
	if in.VehicleID == uuid.Nil {
		return errors.New("vehicle_id is required")
	}
	if in.PartID == uuid.Nil {
		return errors.New("part_id is required")
	}
	if in.PerformedAt.IsZero() {
		return errors.New("performed_at is required")
	}
	if in.OdometerKm < 0 {
		return errors.New("odometer_km cannot be negative")
	}
	return nil
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
