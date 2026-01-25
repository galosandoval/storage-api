package handlers

import (
	"net/http"
	"time"

	"storage-api/internal/db"

	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(database *gorm.DB) *HealthHandler {
	return &HealthHandler{db: database}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *HealthHandler) HealthDB(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(h.db); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "db_unhealthy",
			"error":  err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "db_ok"})
}
