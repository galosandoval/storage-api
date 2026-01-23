package handlers

import (
	"fmt"
	"net/http"

	"storage-api/internal/database"
	"storage-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HouseholdsHandler struct {
	db *pgxpool.Pool
}

func NewHouseholdsHandler(db *pgxpool.Pool) *HouseholdsHandler {
	return &HouseholdsHandler{db: db}
}

// List handles GET /v1/households
// Returns all households (for dev mode / household selection)
func (h *HouseholdsHandler) List(w http.ResponseWriter, r *http.Request) {
	households, err := database.ListHouseholds(r.Context(), h.db)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to list households: %v", err),
		})
		return
	}

	if households == nil {
		households = []models.Household{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"households": households,
	})
}
