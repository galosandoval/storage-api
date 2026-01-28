package middleware

import (
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

// ClerkAuth returns middleware that verifies Clerk JWT tokens.
// It extracts the Bearer token from the Authorization header,
// verifies it, and makes claims available via context.
// Returns 403 Forbidden if the token is missing or invalid.
func ClerkAuth() func(http.Handler) http.Handler {
	return clerkhttp.RequireHeaderAuthorization()
}

// GetClerkUserID extracts the Clerk user ID (subject) from verified claims.
// Returns empty string and false if claims are not present in context.
func GetClerkUserID(r *http.Request) (string, bool) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		return "", false
	}
	return claims.Subject, true
}
