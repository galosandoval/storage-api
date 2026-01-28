package handlers

import (
	"errors"
	"net/http"

	"storage-api/internal/models"
	"storage-api/internal/service"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetMe handles the /me endpoint
// Requires Clerk JWT authentication (handled by middleware)
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get Clerk claims from context (set by middleware)
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error": "no authentication claims found",
		})
		return
	}

	// Fetch user details from Clerk to get email
	clerkUser, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to fetch user from Clerk",
		})
		return
	}

	// Get primary email from Clerk user
	var email string
	for _, e := range clerkUser.EmailAddresses {
		if e.ID == *clerkUser.PrimaryEmailAddressID {
			email = e.EmailAddress
			break
		}
	}

	if email == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error": "no email address found for user",
		})
		return
	}

	// Look up user in our database by email
	u, err := h.svc.GetByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			// User not in household - return 401
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": "user not authorized for this household",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}
