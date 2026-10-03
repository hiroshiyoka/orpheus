package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/api"
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

	go collector.Start(ctx, db, projects, collector.Ping, cfg.FailureThreshold, cfg.TelegramBotToken, cfg.TelegramChatID, nil)
	go collector.StartCloudflareCollector(ctx, db, cfg.CloudflareAPIToken, time.Hour, cfg.TelegramBotToken, cfg.TelegramChatID, nil)

	router := api.NewRouter(db)
	srv := &http.Server{Addr: ":8080", Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
