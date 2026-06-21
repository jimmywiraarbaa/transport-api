// Package delivery exposes the vehicle-type master over HTTP (gin).
package delivery

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/internal/database"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
	"github.com/jimmywiraarbaa/transport-api/vehicletype/domain"
	vtusecase "github.com/jimmywiraarbaa/transport-api/vehicletype/usecase"
)

type handler struct {
	svc vtusecase.Service
}

// Register mounts the vehicle-type CRUD routes. The router must already have
// RequireAuth applied (owner-only master management).
func Register(r gin.IRouter, svc vtusecase.Service) {
	h := &handler{svc: svc}
	r.GET("/vehicle-types", h.list)
	r.GET("/vehicle-types/:id", h.get)
	r.POST("/vehicle-types", h.create)
	r.PUT("/vehicle-types/:id", h.update)
	r.DELETE("/vehicle-types/:id", h.delete)
}

type upsertRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug"`
}

// Register mounts at protected group; RequireAuth is applied by the app.

func (h *handler) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	res, err := h.svc.List(c.Request.Context(), vtusecase.ListInput{Page: page, PerPage: perPage})
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OKWithMeta(c, res.Items, &response.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   int(res.Total),
	})
}

func (h *handler) get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	vt, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, vt)
}

func (h *handler) create(c *gin.Context) {
	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "body", Message: err.Error()}})
		return
	}
	vt, err := h.svc.Create(c.Request.Context(), vtusecase.UpsertInput{Name: req.Name, Slug: req.Slug})
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Created(c, vt)
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
	vt, err := h.svc.Update(c.Request.Context(), id, vtusecase.UpsertInput{Name: req.Name, Slug: req.Slug})
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, vt)
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

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrVehicleTypeNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, vtusecase.ErrValidation):
		response.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
	case database.IsUniqueViolation(err):
		response.Conflict(c, "slug or name already exists")
	default:
		response.Internal(c, err)
	}
}
