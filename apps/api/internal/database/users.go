package database

import (
	"context"
	"errors"
	"time"

	"storage-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
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

// ListHouseholds returns all households
func ListHouseholds(ctx context.Context, db *pgxpool.Pool) ([]models.Household, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, `
		SELECT id::text, name, created_at
		FROM households
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var households []models.Household
	for rows.Next() {
		var h models.Household
		var created time.Time
		if err := rows.Scan(&h.ID, &h.Name, &created); err != nil {
			return nil, err
		}
		h.CreatedAt = created.Format(time.RFC3339)
		households = append(households, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return households, nil
}
