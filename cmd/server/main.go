package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/hiroshiyoka/orpheus/internal/collector"
	"github.com/hiroshiyoka/orpheus/internal/config"
	"github.com/hiroshiyoka/orpheus/internal/storage"
)

func main() {
	cfg := config.Load()
	db, err := storage.Open(cfg.DBPath, "migrations/0001_init.sql")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	projects, err := storage.ListProjects(db)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	collector.Start(ctx, db, projects, collector.Ping)
}
