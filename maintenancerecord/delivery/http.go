// Package delivery exposes the maintenance-record feature over HTTP (gin).
package delivery

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/internal/database"
	"github.com/jimmywiraarbaa/transport-api/internal/middleware"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
	"github.com/jimmywiraarbaa/transport-api/maintenancerecord/domain"
	mrusecase "github.com/jimmywiraarbaa/transport-api/maintenancerecord/usecase"
	"github.com/shopspring/decimal"
)

type handler struct {
	svc mrusecase.Service
}

// Register mounts maintenance-record routes on a protected router.
func Register(r gin.IRouter, svc mrusecase.Service) {
	h := &handler{svc: svc}
	r.GET("/vehicles/:id/maintenance-records", h.list)
	r.POST("/vehicles/:id/maintenance-records", h.create)
	r.GET("/maintenance-records", h.listAll)
	r.GET("/maintenance-records/:id", h.get)
	r.PUT("/maintenance-records/:id", h.update)
	r.DELETE("/maintenance-records/:id", h.delete)
}

type upsertRequest struct {
	PartID      uuid.UUID `json:"part_id" binding:"required"`
	PerformedAt string    `json:"performed_at" binding:"required"`
	OdometerKm  int32     `json:"odometer_km"`
	Cost        string    `json:"cost"`
	VehicleID   uuid.UUID `json:"vehicle_id"` // used on update
	Technician  string    `json:"technician"`
	Notes       string    `json:"notes"`
}

func (h *handler) list(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	vehicleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle id")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	res, err := h.svc.List(c.Request.Context(), userID, vehicleID, mrusecase.ListInput{Page: page, PerPage: perPage})
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OKWithMeta(c, res.Items, &response.Meta{Page: page, PerPage: perPage, Total: int(res.Total)})
}

func (h *handler) listAll(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	var vehicleID *uuid.UUID
	if raw := c.Query("vehicle_id"); raw != "" {
		vid, err := uuid.Parse(raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle_id filter")
			return
		}
		vehicleID = &vid
	}

	res, err := h.svc.ListAll(c.Request.Context(), userID, vehicleID, mrusecase.ListInput{Page: page, PerPage: perPage})
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OKWithMeta(c, res.Items, &response.Meta{Page: page, PerPage: perPage, Total: int(res.Total)})
}

func (h *handler) get(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	rec, err := h.svc.Get(c.Request.Context(), id, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, rec)
}

func (h *handler) create(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	vehicleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle id")
		return
	}
	in, ok := parseUpsert(c)
	if !ok {
		return
	}
	rec, err := h.svc.Create(c.Request.Context(), userID, mrusecase.UpsertInput{
		VehicleID: vehicleID, PartID: in.PartID, PerformedAt: in.PerformedAt,
		OdometerKm: in.OdometerKm, Cost: in.Cost, Technician: in.Technician, Notes: in.Notes,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Created(c, rec)
}

func (h *handler) update(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	in, ok := parseUpsert(c)
	if !ok {
		return
	}
	// On update, the vehicle id may come from the body (records aren't moved across vehicles via path).
	vehicleID := in.VehicleID
	if vehicleID == uuid.Nil {
		response.ValidationError(c, []response.FieldError{{Field: "vehicle_id", Message: "vehicle_id is required"}})
		return
	}
	rec, err := h.svc.Update(c.Request.Context(), id, userID, mrusecase.UpsertInput{
		VehicleID: vehicleID, PartID: in.PartID, PerformedAt: in.PerformedAt,
		OdometerKm: in.OdometerKm, Cost: in.Cost, Technician: in.Technician, Notes: in.Notes,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, rec)
}

func (h *handler) delete(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		writeErr(c, err)
		return
	}
	response.NoContent(c)
}

// parsedInput is the validated, typed view of the request body.
type parsedInput struct {
	PartID      uuid.UUID
	VehicleID   uuid.UUID
	PerformedAt time.Time
	OdometerKm  int32
	Cost        decimal.Decimal
	Technician  string
	Notes       string
}

// parseUpsert binds the JSON body and converts performed_at + cost to typed values.
func parseUpsert(c *gin.Context) (*parsedInput, bool) {
	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "body", Message: err.Error()}})
		return nil, false
	}

	performed, ok := parsePerformedAt(c, req.PerformedAt)
	if !ok {
		return nil, false
	}

	cost, err := parseCost(req.Cost)
	if err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "cost", Message: err.Error()}})
		return nil, false
	}

	return &parsedInput{
		PartID:      req.PartID,
		VehicleID:   req.VehicleID,
		PerformedAt: performed,
		OdometerKm:  req.OdometerKm,
		Cost:        cost,
		Technician:  req.Technician,
		Notes:       req.Notes,
	}, true
}

func parsePerformedAt(c *gin.Context, raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}
	response.ValidationError(c, []response.FieldError{{
		Field: "performed_at", Message: "must be a valid date (YYYY-MM-DD) or RFC3339 datetime",
	}})
	return time.Time{}, false
}

func parseCost(raw string) (decimal.Decimal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(raw)
}

func mustUserID(c *gin.Context) uuid.UUID {
	raw := middleware.UserID(c)
	id, err := uuid.Parse(raw)
	if err != nil {
		response.Unauthorized(c, "missing user identity")
		return uuid.Nil
	}
	return id
}

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrMaintenanceRecordNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrVehicleNotOwned):
		response.NotFound(c, "vehicle not found")
	case errors.Is(err, mrusecase.ErrValidation):
		response.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
	case database.IsForeignKeyViolation(err):
		response.Error(c, http.StatusBadRequest, "INVALID_REFERENCE", "Referenced vehicle or part does not exist")
	default:
		response.Internal(c, err)
	}
}
