package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/internal/config"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/jwt"
)

// Container holds shared application dependencies wired once at boot and
// handed to each feature's delivery layer. Keeping it in app/ lets feature
// packages stay decoupled from bootstrap concerns.
type Container struct {
	Config *config.Config
	DB     *pgxpool.Pool
	JWT    *jwt.Manager
}

// NewContainer builds the dependency container from configuration + the db pool.
func NewContainer(cfg *config.Config, db *pgxpool.Pool) *Container {
	return &Container{
		Config: cfg,
		DB:     db,
		JWT:    jwt.New(cfg.JWT),
	}
}
