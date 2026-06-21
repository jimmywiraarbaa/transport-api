package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/vehicle/domain"
)

// ErrValidation signals an invalid input (maps to HTTP 422).
var ErrValidation = errors.New("validation error")

// Repository is the persistence port the vehicle usecase needs.
type Repository interface {
	List(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.Vehicle, error)
	Count(ctx context.Context, userID uuid.UUID) (int64, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*domain.Vehicle, error)
	Create(ctx context.Context, v domain.Vehicle) (*domain.Vehicle, error)
	Update(ctx context.Context, v domain.Vehicle) (*domain.Vehicle, error)
	UpdateOdometerIfHigher(ctx context.Context, id, userID uuid.UUID, km int32) (*domain.Vehicle, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

// ListInput holds list/pagination params.
type ListInput struct {
	Page    int
	PerPage int
}

// CreateInput holds the fields for creating a vehicle.
type CreateInput struct {
	VehicleTypeID     uuid.UUID
	PlateNumber       string
	Brand             string
	Model             string
	Year              *int32
	CurrentOdometerKm int32
	Notes             string
}

// UpdateInput holds the mutable fields of a vehicle (odometer updated separately).
type UpdateInput struct {
	VehicleTypeID uuid.UUID
	PlateNumber   string
	Brand         string
	Model         string
	Year          *int32
	Notes         string
}

// ListResult holds a page of items with total count.
type ListResult struct {
	Items []domain.Vehicle
	Total int64
}

type Service interface {
	List(ctx context.Context, userID uuid.UUID, in ListInput) (*ListResult, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*domain.Vehicle, error)
	Create(ctx context.Context, userID uuid.UUID, in CreateInput) (*domain.Vehicle, error)
	Update(ctx context.Context, id, userID uuid.UUID, in UpdateInput) (*domain.Vehicle, error)
	UpdateOdometer(ctx context.Context, id, userID uuid.UUID, km int32) (*domain.Vehicle, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type service struct {
	repo Repository
}

func New(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, userID uuid.UUID, in ListInput) (*ListResult, error) {
	page, perPage := normalizePaging(in.Page, in.PerPage)
	offset := int32((page - 1) * perPage)

	items, err := s.repo.List(ctx, userID, int32(perPage), offset)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.Count(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total}, nil
}

func (s *service) Get(ctx context.Context, id, userID uuid.UUID) (*domain.Vehicle, error) {
	return s.repo.Get(ctx, id, userID)
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, in CreateInput) (*domain.Vehicle, error) {
	if err := validateCreate(in); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	v := domain.Vehicle{
		UserID:            userID,
		VehicleTypeID:     in.VehicleTypeID,
		PlateNumber:       strings.ToUpper(strings.TrimSpace(in.PlateNumber)),
		Brand:             strings.TrimSpace(in.Brand),
		Model:             strings.TrimSpace(in.Model),
		Year:              in.Year,
		CurrentOdometerKm: in.CurrentOdometerKm,
		InitialOdometerKm: in.CurrentOdometerKm,
		Notes:             strings.TrimSpace(in.Notes),
	}
	return s.repo.Create(ctx, v)
}

func (s *service) Update(ctx context.Context, id, userID uuid.UUID, in UpdateInput) (*domain.Vehicle, error) {
	if err := validateUpdate(in); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	v := domain.Vehicle{
		ID:            id,
		UserID:        userID,
		VehicleTypeID: in.VehicleTypeID,
		PlateNumber:   strings.ToUpper(strings.TrimSpace(in.PlateNumber)),
		Brand:         strings.TrimSpace(in.Brand),
		Model:         strings.TrimSpace(in.Model),
		Year:          in.Year,
		Notes:         strings.TrimSpace(in.Notes),
	}
	return s.repo.Update(ctx, v)
}

func (s *service) UpdateOdometer(ctx context.Context, id, userID uuid.UUID, km int32) (*domain.Vehicle, error) {
	if km < 0 {
		return nil, fmt.Errorf("%w: odometer cannot be negative", ErrValidation)
	}
	// Verify the vehicle exists (and is owned) before deciding whether to advance.
	v, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	// No-op when the reading is not higher than the current odometer.
	if km <= v.CurrentOdometerKm {
		return v, nil
	}
	return s.repo.UpdateOdometerIfHigher(ctx, id, userID, km)
}

func (s *service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

func validateCreate(in CreateInput) error {
	if in.VehicleTypeID == uuid.Nil {
		return errors.New("vehicle_type_id is required")
	}
	if strings.TrimSpace(in.PlateNumber) == "" {
		return errors.New("plate_number is required")
	}
	if in.CurrentOdometerKm < 0 {
		return errors.New("current_odometer_km cannot be negative")
	}
	if in.Year != nil && (*in.Year < 1900 || *in.Year > 2100) {
		return errors.New("year is out of range")
	}
	return nil
}

func validateUpdate(in UpdateInput) error {
	if in.VehicleTypeID == uuid.Nil {
		return errors.New("vehicle_type_id is required")
	}
	if strings.TrimSpace(in.PlateNumber) == "" {
		return errors.New("plate_number is required")
	}
	if in.Year != nil && (*in.Year < 1900 || *in.Year > 2100) {
		return errors.New("year is out of range")
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
