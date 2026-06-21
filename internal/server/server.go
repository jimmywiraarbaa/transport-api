package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jimmywiraarbaa/transport-api/internal/config"
)

// Server wraps the underlying HTTP server and the gin engine.
type Server struct {
	cfg     *config.Config
	router  *gin.Engine
	httpSrv *http.Server
}

// New creates a configured HTTP server. Routes are registered via the passed
// registrar function so feature packages stay decoupled from bootstrapping.
func New(cfg *config.Config, register func(r *gin.Engine)) *Server {
	if cfg.App.Env != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	register(r)

	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	return &Server{
		cfg:     cfg,
		router:  r,
		httpSrv: srv,
	}
}

// Start boots the HTTP server (blocking).
func (s *Server) Start() error {
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}
