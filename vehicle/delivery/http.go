// Package delivery exposes the vehicle feature over HTTP (gin).
package delivery

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/internal/database"
	"github.com/jimmywiraarbaa/transport-api/internal/middleware"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
	"github.com/jimmywiraarbaa/transport-api/vehicle/domain"
	vusecase "github.com/jimmywiraarbaa/transport-api/vehicle/usecase"
)

type handler struct {
	svc vusecase.Service
}

// Register mounts the vehicle CRUD routes on a protected router.
func Register(r gin.IRouter, svc vusecase.Service) {
	h := &handler{svc: svc}
	r.GET("/vehicles", h.list)
	r.GET("/vehicles/:id", h.get)
	r.POST("/vehicles", h.create)
	r.PUT("/vehicles/:id", h.update)
	r.PATCH("/vehicles/:id/odometer", h.updateOdometer)
	r.DELETE("/vehicles/:id", h.delete)
}

type createRequest struct {
	VehicleTypeID     uuid.UUID `json:"vehicle_type_id" binding:"required"`
	PlateNumber       string    `json:"plate_number" binding:"required"`
	Brand             string    `json:"brand"`
	Model             string    `json:"model"`
	Year              *int32    `json:"year"`
	CurrentOdometerKm int32     `json:"current_odometer_km"`
	Notes             string    `json:"notes"`
}

type updateRequest struct {
	VehicleTypeID uuid.UUID `json:"vehicle_type_id" binding:"required"`
	PlateNumber   string    `json:"plate_number" binding:"required"`
	Brand         string    `json:"brand"`
	Model         string    `json:"model"`
	Year          *int32    `json:"year"`
	Notes         string    `json:"notes"`
}

type odometerRequest struct {
	OdometerKm int32 `json:"odometer_km" binding:"required"`
}

func (h *handler) list(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	res, err := h.svc.List(c.Request.Context(), userID, vusecase.ListInput{Page: page, PerPage: perPage})
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
	v, err := h.svc.Get(c.Request.Context(), id, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, v)
}

func (h *handler) create(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "body", Message: err.Error()}})
		return
	}
	v, err := h.svc.Create(c.Request.Context(), userID, vusecase.CreateInput{
		VehicleTypeID:     req.VehicleTypeID,
		PlateNumber:       req.PlateNumber,
		Brand:             req.Brand,
		Model:             req.Model,
		Year:              req.Year,
		CurrentOdometerKm: req.CurrentOdometerKm,
		Notes:             req.Notes,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Created(c, v)
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
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "body", Message: err.Error()}})
		return
	}
	v, err := h.svc.Update(c.Request.Context(), id, userID, vusecase.UpdateInput{
		VehicleTypeID: req.VehicleTypeID,
		PlateNumber:   req.PlateNumber,
		Brand:         req.Brand,
		Model:         req.Model,
		Year:          req.Year,
		Notes:         req.Notes,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, v)
}

func (h *handler) updateOdometer(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	var req odometerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "odometer_km", Message: err.Error()}})
		return
	}
	v, err := h.svc.UpdateOdometer(c.Request.Context(), id, userID, req.OdometerKm)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, v)
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

// mustUserID extracts the authenticated user id; writes 401 and returns Nil on failure.
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
	case errors.Is(err, domain.ErrVehicleNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, vusecase.ErrValidation):
		response.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
	case database.IsUniqueViolation(err):
		response.Conflict(c, "plate number already exists")
	case database.IsForeignKeyViolation(err):
		response.Error(c, http.StatusBadRequest, "INVALID_REFERENCE", "Referenced vehicle type does not exist")
	default:
		response.Internal(c, err)
	}
}
