package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jimmywiraarbaa/transport-api/app"
	"github.com/jimmywiraarbaa/transport-api/internal/config"
	"github.com/jimmywiraarbaa/transport-api/internal/database"
	"github.com/jimmywiraarbaa/transport-api/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	log.Println("running migrations...")
	if err := database.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	container := app.NewContainer(cfg, pool)
	srv := server.New(cfg, container.RegisterRoutes)

	go func() {
		log.Printf("transport-api listening on :%s", cfg.App.Port)
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down transport-api...")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("transport-api stopped")
}
