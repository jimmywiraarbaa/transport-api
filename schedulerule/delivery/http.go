// Package delivery exposes the schedule-rule master over HTTP (gin).
package delivery

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/internal/database"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
	"github.com/jimmywiraarbaa/transport-api/schedulerule/domain"
	srusecase "github.com/jimmywiraarbaa/transport-api/schedulerule/usecase"
)

type handler struct {
	svc srusecase.Service
}

// Register mounts the schedule-rule CRUD routes on a protected router.
func Register(r gin.IRouter, svc srusecase.Service) {
	h := &handler{svc: svc}
	r.GET("/schedule-rules", h.list)
	r.GET("/schedule-rules/:id", h.get)
	r.GET("/vehicle-types/:id/schedule-rules", h.listByVehicleType)
	r.POST("/schedule-rules", h.create)
	r.PUT("/schedule-rules/:id", h.update)
	r.DELETE("/schedule-rules/:id", h.delete)
}

type upsertRequest struct {
	PartID        uuid.UUID `json:"part_id" binding:"required"`
	VehicleTypeID uuid.UUID `json:"vehicle_type_id" binding:"required"`
	IntervalKm    *int32    `json:"interval_km"`
	IntervalDays  *int32    `json:"interval_days"`
	TriggerMode   string    `json:"trigger_mode"`
	Notes         string    `json:"notes"`
}

func (h *handler) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	res, err := h.svc.List(c.Request.Context(), srusecase.ListInput{Page: page, PerPage: perPage})
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OKWithMeta(c, res.Items, &response.Meta{Page: page, PerPage: perPage, Total: int(res.Total)})
}

func (h *handler) get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	s, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, s)
}

func (h *handler) listByVehicleType(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	items, err := h.svc.ListByVehicleType(c.Request.Context(), id)
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OK(c, items)
}

func (h *handler) create(c *gin.Context) {
	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "body", Message: err.Error()}})
		return
	}
	s, err := h.svc.Create(c.Request.Context(), toUpsert(req))
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Created(c, s)
}

func (h *handler) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "body", Message: err.Error()}})
		return
	}
	s, err := h.svc.Update(c.Request.Context(), id, toUpsert(req))
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, s)
}

func (h *handler) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	response.NoContent(c)
}

func toUpsert(req upsertRequest) srusecase.UpsertInput {
	return srusecase.UpsertInput{
		PartID:        req.PartID,
		VehicleTypeID: req.VehicleTypeID,
		IntervalKm:    req.IntervalKm,
		IntervalDays:  req.IntervalDays,
		TriggerMode:   domain.TriggerMode(req.TriggerMode),
		Notes:         req.Notes,
	}
}

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrScheduleRuleNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, srusecase.ErrValidation):
		response.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
	case database.IsUniqueViolation(err):
		response.Conflict(c, "a rule for this part and vehicle type already exists")
	case database.IsForeignKeyViolation(err):
		response.Error(c, http.StatusBadRequest, "INVALID_REFERENCE", "Referenced part or vehicle type does not exist")
	default:
		response.Internal(c, err)
	}
}
