// Package delivery exposes the maintenance-part master over HTTP (gin).
package delivery

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/internal/database"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
	"github.com/jimmywiraarbaa/transport-api/maintenancepart/domain"
	mpusecase "github.com/jimmywiraarbaa/transport-api/maintenancepart/usecase"
)

type handler struct {
	svc mpusecase.Service
}

// Register mounts the maintenance-part CRUD routes on a protected router.
func Register(r gin.IRouter, svc mpusecase.Service) {
	h := &handler{svc: svc}
	r.GET("/maintenance-parts", h.list)
	r.GET("/maintenance-parts/:id", h.get)
	r.POST("/maintenance-parts", h.create)
	r.PUT("/maintenance-parts/:id", h.update)
	r.DELETE("/maintenance-parts/:id", h.delete)
}

type upsertRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

func (h *handler) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	res, err := h.svc.List(c.Request.Context(), mpusecase.ListInput{Page: page, PerPage: perPage})
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
	p, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, p)
}

func (h *handler) create(c *gin.Context) {
	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{{Field: "body", Message: err.Error()}})
		return
	}
	p, err := h.svc.Create(c.Request.Context(), mpusecase.UpsertInput{
		Name: req.Name, Slug: req.Slug, Category: req.Category, Description: req.Description,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Created(c, p)
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
	p, err := h.svc.Update(c.Request.Context(), id, mpusecase.UpsertInput{
		Name: req.Name, Slug: req.Slug, Category: req.Category, Description: req.Description,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, p)
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
	case errors.Is(err, domain.ErrMaintenancePartNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, mpusecase.ErrValidation):
		response.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
	case database.IsUniqueViolation(err):
		response.Conflict(c, "slug or name already exists")
	default:
		response.Internal(c, err)
	}
}
