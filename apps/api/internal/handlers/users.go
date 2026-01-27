package handlers

import (
	"errors"
	"net/http"

	"storage-api/internal/models"
	"storage-api/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetMe handles the /me endpoint
// Dev auth: provide X-Dev-User header as external_sub (temporary)
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	sub := r.Header.Get("X-Dev-User")
	if sub == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error": "missing X-Dev-User header (dev mode). Example: X-Dev-User: dev-you",
		})
		return
	}

	u, err := h.svc.GetByExternalSub(r.Context(), sub)
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, models.ErrNotFound) {
			status = http.StatusInternalServerError
		}
		writeJSON(w, status, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}
