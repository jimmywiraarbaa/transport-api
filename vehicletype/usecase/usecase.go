package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/text"
	"github.com/jimmywiraarbaa/transport-api/vehicletype/domain"
)

// ErrValidation signals an invalid input (maps to HTTP 422).
var ErrValidation = errors.New("validation error")

// Repository is the persistence port the vehicletype usecase needs.
type Repository interface {
	List(ctx context.Context, limit, offset int32) ([]domain.VehicleType, error)
	Count(ctx context.Context) (int64, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.VehicleType, error)
	Create(ctx context.Context, name, slug string) (*domain.VehicleType, error)
	Update(ctx context.Context, id uuid.UUID, name, slug string) (*domain.VehicleType, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ListInput holds list/pagination params.
type ListInput struct {
	Page    int
	PerPage int
}

// UpsertInput holds create/update fields.
type UpsertInput struct {
	Name string
	Slug string // optional; derived from Name when empty
}

// ListResult holds a page of items with total count.
type ListResult struct {
	Items []domain.VehicleType
	Total int64
}

type Service interface {
	List(ctx context.Context, in ListInput) (*ListResult, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.VehicleType, error)
	Create(ctx context.Context, in UpsertInput) (*domain.VehicleType, error)
	Update(ctx context.Context, id uuid.UUID, in UpsertInput) (*domain.VehicleType, error)
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

func (s *service) Get(ctx context.Context, id uuid.UUID) (*domain.VehicleType, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Create(ctx context.Context, in UpsertInput) (*domain.VehicleType, error) {
	if err := validateUpsert(in); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	name := strings.TrimSpace(in.Name)
	slug := text.Slug(in.Slug)
	if slug == "" {
		slug = text.Slug(name)
	}
	return s.repo.Create(ctx, name, slug)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, in UpsertInput) (*domain.VehicleType, error) {
	if err := validateUpsert(in); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	name := strings.TrimSpace(in.Name)
	slug := text.Slug(in.Slug)
	if slug == "" {
		slug = text.Slug(name)
	}
	return s.repo.Update(ctx, id, name, slug)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func validateUpsert(in UpsertInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("name is required")
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
