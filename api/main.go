package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Addr string
	DSN  string
}

type User struct {
	ID          string `json:"id"`
	HouseholdID string `json:"householdId"`
	ExternalSub string `json:"externalSub"`
	Email       string `json:"email,omitempty"`
	Role        string `json:"role"`
	CreatedAt   string `json:"createdAt"`
}

func main() {
	cfg := Config{
		Addr: getenv("ADDR", ":8080"),
		DSN:  getenv("DATABASE_URL", "postgres://storageapp:change_me_now@localhost:5432/storage_db?sslmode=disable"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("db ping error: %v", err)
	}

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	r.Get("/health/db", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var one int
		if err := db.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status": "db_unhealthy",
				"error":  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "db_ok"})
	})

	// Dev auth endpoint: provide X-Dev-User header as external_sub (temporary)
	r.Get("/v1/me", func(w http.ResponseWriter, r *http.Request) {
		sub := r.Header.Get("X-Dev-User")
		if sub == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": "missing X-Dev-User header (dev mode). Example: X-Dev-User: dev-you",
			})
			return
		}

		u, err := getUserByExternalSub(r.Context(), db, sub)
		if err != nil {
			status := http.StatusUnauthorized
			if !errors.Is(err, errNotFound) {
				status = http.StatusInternalServerError
			}
			writeJSON(w, status, map[string]any{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"user": u})
	})

	log.Printf("API listening on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}

var errNotFound = errors.New("not found")

func getUserByExternalSub(ctx context.Context, db *pgxpool.Pool, sub string) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var u User
	var created time.Time

	err := db.QueryRow(ctx, `
		SELECT id::text,
		       household_id::text,
		       external_sub,
		       COALESCE(email,''),
		       role,
		       created_at
		FROM users
		WHERE external_sub = $1
	`, sub).Scan(&u.ID, &u.HouseholdID, &u.ExternalSub, &u.Email, &u.Role, &created)

	if err != nil {
		// pgx returns an error for no rows; to keep it simple, treat any scan error as not found here.
		// Later we can be more precise using pgx.ErrNoRows.
		return User{}, errNotFound
	}

	u.CreatedAt = created.Format(time.RFC3339)
	return u, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

