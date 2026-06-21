// Package delivery exposes the auth feature over HTTP (gin).
package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	authusecase "github.com/jimmywiraarbaa/transport-api/auth/usecase"
	"github.com/jimmywiraarbaa/transport-api/internal/middleware"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/jwt"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
)

type handler struct {
	svc authusecase.Service
}

// Register mounts the public auth routes onto the given router group.
func Register(r gin.IRouter, svc authusecase.Service) {
	h := &handler{svc: svc}
	r.POST("/login", h.login)
	r.POST("/refresh", h.refresh)
}

// RegisterProtected mounts auth routes that require a valid access token.
func RegisterProtected(r gin.IRouter, svc authusecase.Service, mgr *jwt.Manager) {
	h := &handler{svc: svc}
	g := r.Group("")
	g.Use(middleware.RequireAuth(mgr))
	g.GET("/me", h.me)
}

// ── request DTOs ──────────────────────────────────────

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ── response DTOs ─────────────────────────────────────

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type tokenResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    string       `json:"expires_at"`
}

// ── handlers ──────────────────────────────────────────

func (h *handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{
			{Field: "body", Message: err.Error()},
		})
		return
	}

	res, err := h.svc.Login(c.Request.Context(), authusecase.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, toTokenResponse(res))
}

func (h *handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, []response.FieldError{
			{Field: "refresh_token", Message: err.Error()},
		})
		return
	}

	res, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, toTokenResponse(res))
}

func (h *handler) me(c *gin.Context) {
	userIDStr := middleware.UserID(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "invalid token identity")
		return
	}

	user, err := h.svc.Current(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.OK(c, userResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format(timeRFC3339),
	})
}

// writeError maps usecase errors to HTTP envelopes.
func (h *handler) writeError(c *gin.Context, err error) {
	switch err {
	case authusecase.ErrInvalidCredential:
		response.Unauthorized(c, err.Error())
	case authusecase.ErrInvalidToken:
		response.Unauthorized(c, err.Error())
	default:
		response.Internal(c, err)
	}
}

func toTokenResponse(r *authusecase.AuthResult) tokenResponse {
	return tokenResponse{
		User: userResponse{
			ID:        r.User.ID.String(),
			Email:     r.User.Email,
			Username:  r.User.Username,
			Name:      r.User.Name,
			CreatedAt: r.User.CreatedAt.Format(timeRFC3339),
		},
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresAt:    r.ExpiresAt.Format(timeRFC3339),
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
