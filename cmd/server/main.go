package main

import (
	"log"

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
}
