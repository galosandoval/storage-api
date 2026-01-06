package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"storage-api/internal/config"
)

func main() {
	var (
		dir     = flag.String("dir", "migrations", "directory with migration files")
		command = flag.String("command", "up", "goose command (up, down, status, version)")
	)
	flag.Parse()

	cfg := config.Load()

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	switch *command {
	case "up":
		if err := goose.Up(db, *dir); err != nil {
			log.Fatalf("migration up failed: %v", err)
		}
		fmt.Println("✓ Migrations completed successfully")
	case "down":
		if err := goose.Down(db, *dir); err != nil {
			log.Fatalf("migration down failed: %v", err)
		}
		fmt.Println("✓ Rolled back last migration")
	case "status":
		if err := goose.Status(db, *dir); err != nil {
			log.Fatalf("migration status failed: %v", err)
		}
	case "version":
		version, err := goose.GetDBVersion(db)
		if err != nil {
			log.Fatalf("failed to get version: %v", err)
		}
		fmt.Printf("Current version: %d\n", version)
	default:
		log.Fatalf("unknown command: %s", *command)
	}
}

