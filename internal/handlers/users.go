package handlers

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"storage-api/internal/database"
)

type UserHandler struct {
	db *pgxpool.Pool
}

func NewUserHandler(db *pgxpool.Pool) *UserHandler {
	return &UserHandler{db: db}
}

// GetMe handles the /v1/me endpoint
// Dev auth: provide X-Dev-User header as external_sub (temporary)
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	sub := r.Header.Get("X-Dev-User")
	if sub == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error": "missing X-Dev-User header (dev mode). Example: X-Dev-User: dev-you",
		})
		return
	}

	u, err := database.GetUserByExternalSub(r.Context(), h.db, sub)
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, database.ErrNotFound) {
			status = http.StatusInternalServerError
		}
		writeJSON(w, status, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}

