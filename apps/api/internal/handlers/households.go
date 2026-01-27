package handlers

import (
	"fmt"
	"net/http"

	"storage-api/internal/models"
	"storage-api/internal/service"
)

type HouseholdsHandler struct {
	svc *service.HouseholdService
}

func NewHouseholdsHandler(svc *service.HouseholdService) *HouseholdsHandler {
	return &HouseholdsHandler{svc: svc}
}

// List handles GET /households
// Returns all households (for dev mode / household selection)
func (h *HouseholdsHandler) List(w http.ResponseWriter, r *http.Request) {
	households, err := h.svc.List(r.Context())
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
