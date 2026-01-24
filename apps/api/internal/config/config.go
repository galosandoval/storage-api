package config

import "os"

type Config struct {
	Addr string
	DSN  string
}

func Load() Config {
	return Config{
		Addr: getenv("ADDR", ":8080"),
		DSN:  getenv("DATABASE_URL", ""),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

