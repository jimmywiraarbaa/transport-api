package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jimmywiraarbaa/transport-api/auth/delivery"
	authrepo "github.com/jimmywiraarbaa/transport-api/auth/repository"
	authusecase "github.com/jimmywiraarbaa/transport-api/auth/usecase"
	mpdelivery "github.com/jimmywiraarbaa/transport-api/maintenancepart/delivery"
	mprepo "github.com/jimmywiraarbaa/transport-api/maintenancepart/repository"
	mpusecase "github.com/jimmywiraarbaa/transport-api/maintenancepart/usecase"
	"github.com/jimmywiraarbaa/transport-api/internal/middleware"
	srdelivery "github.com/jimmywiraarbaa/transport-api/schedulerule/delivery"
	srrepo "github.com/jimmywiraarbaa/transport-api/schedulerule/repository"
	srusecase "github.com/jimmywiraarbaa/transport-api/schedulerule/usecase"
	mrdelivery "github.com/jimmywiraarbaa/transport-api/maintenancerecord/delivery"
	mrrepo "github.com/jimmywiraarbaa/transport-api/maintenancerecord/repository"
	mrusecase "github.com/jimmywiraarbaa/transport-api/maintenancerecord/usecase"
	madelivery "github.com/jimmywiraarbaa/transport-api/maintenancealert/delivery"
	marepo "github.com/jimmywiraarbaa/transport-api/maintenancealert/repository"
	mausecase "github.com/jimmywiraarbaa/transport-api/maintenancealert/usecase"
	vdelivery "github.com/jimmywiraarbaa/transport-api/vehicle/delivery"
	vrepo "github.com/jimmywiraarbaa/transport-api/vehicle/repository"
	vusecase "github.com/jimmywiraarbaa/transport-api/vehicle/usecase"
	vtdelivery "github.com/jimmywiraarbaa/transport-api/vehicletype/delivery"
	vtrepo "github.com/jimmywiraarbaa/transport-api/vehicletype/repository"
	vtusecase "github.com/jimmywiraarbaa/transport-api/vehicletype/usecase"
)

// RegisterRoutes mounts global middleware and all feature routers.
//
// Layout:
//
//	/health                       → liveness probe (public)
//	/api/v1/auth/...              → auth feature (Phase 3)
//	/api/v1/...  (RequireAuth)    → protected feature routes
func (c *Container) RegisterRoutes(r *gin.Engine) {
	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"app":    "transport-api",
		})
	})

	api := r.Group("/api/v1")

	// Auth feature (public)
	authRepo := authrepo.New(c.DB)
	authSvc := authusecase.New(authRepo, c.JWT)
	delivery.Register(api.Group("/auth"), authSvc)

	// Auth feature (protected)
	delivery.RegisterProtected(api.Group("/auth"), authSvc, c.JWT)

	// Protected feature routers
	protected := api.Group("")
	protected.Use(middleware.RequireAuth(c.JWT))

	// Master data CRUD
	vtdelivery.Register(protected, vtusecase.New(vtrepo.New(c.DB)))
	mpdelivery.Register(protected, mpusecase.New(mprepo.New(c.DB)))
	srdelivery.Register(protected, srusecase.New(srrepo.New(c.DB)))

	// Operational: vehicles + maintenance records
	vdelivery.Register(protected, vusecase.New(vrepo.New(c.DB)))
	mrdelivery.Register(protected, mrusecase.New(mrrepo.New(c.DB)))

	// Core: maintenance alerts (read-model computation)
	madelivery.Register(protected, mausecase.New(marepo.New(c.DB)))
}
