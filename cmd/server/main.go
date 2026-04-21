package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"lan-topology-mapper/internal/api"
	"lan-topology-mapper/internal/config"
	"lan-topology-mapper/internal/discovery"
	"lan-topology-mapper/internal/store"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer db.Close()

	scanner := discovery.NewScanner(cfg)
	server := api.NewServer(db, scanner, cfg.StaticDir)

	log.Printf("listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, server.Handler()); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
