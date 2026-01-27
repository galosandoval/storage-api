package config

import "os"

type Config struct {
	Addr           string
	DSN            string
	ClerkSecretKey string
}

func Load() Config {
	return Config{
		Addr:           getenv("ADDR", ":8080"),
		DSN:            getenv("DATABASE_URL", ""),
		ClerkSecretKey: getenv("CLERK_SECRET_KEY", ""),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

