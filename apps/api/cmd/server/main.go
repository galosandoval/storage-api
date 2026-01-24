package main

import (
	"log"

	"storage-api/internal/config"
	"storage-api/internal/server"
)

func main() {
	cfg := config.Load()

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	defer srv.Close()

	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

