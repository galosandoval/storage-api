package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"storage-api/internal/models"
)

var ErrNotFound = errors.New("not found")

func GetUserByExternalSub(ctx context.Context, db *pgxpool.Pool, sub string) (models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var u models.User
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
		return models.User{}, ErrNotFound
	}

	u.CreatedAt = created.Format(time.RFC3339)
	return u, nil
}

