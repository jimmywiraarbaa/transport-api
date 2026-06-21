package delivery

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jimmywiraarbaa/transport-api/internal/middleware"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
	"github.com/jimmywiraarbaa/transport-api/maintenancealert/domain"
	ausecase "github.com/jimmywiraarbaa/transport-api/maintenancealert/usecase"
)

type handler struct {
	svc ausecase.Service
}

// Register mounts the alert route on a protected router.
func Register(r gin.IRouter, svc ausecase.Service) {
	h := &handler{svc: svc}
	r.GET("/vehicles/:id/alerts", h.compute)
}

func (h *handler) compute(c *gin.Context) {
	userID := mustUserID(c)
	if userID == uuid.Nil {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid id format")
		return
	}
	summary, err := h.svc.Compute(c.Request.Context(), id, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, summary)
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
	case errors.Is(err, domain.ErrVehicleNotFound):
		response.NotFound(c, err.Error())
	default:
		response.Internal(c, err)
	}
}
