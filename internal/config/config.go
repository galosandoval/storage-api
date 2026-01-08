package config

import "os"

type Config struct {
	Addr string
	DSN  string
}

func Load() Config {
	return Config{
		Addr: getenv("ADDR", ":8080"),
		DSN:  getenv("DATABASE_URL", "postgres://storageapp:change_me_now@localhost:5432/storage_db?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

