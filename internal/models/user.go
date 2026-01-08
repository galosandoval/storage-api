package models

type User struct {
	ID          string `json:"id"`
	HouseholdID string `json:"householdId"`
	ExternalSub string `json:"externalSub"`
	Email       string `json:"email,omitempty"`
	Role        string `json:"role"`
	CreatedAt   string `json:"createdAt"`
}

