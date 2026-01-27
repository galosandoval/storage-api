package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HouseholdID uuid.UUID `gorm:"type:uuid;not null;index" json:"householdId"`
	ExternalSub string    `gorm:"size:255;not null;uniqueIndex" json:"externalSub"`
	Email       string    `gorm:"size:255" json:"email,omitempty"`
	FirstName   string    `gorm:"size:255" json:"firstName,omitempty"`
	LastName    string    `gorm:"size:255" json:"lastName,omitempty"`
	ImageURL    string    `gorm:"size:500" json:"imageUrl,omitempty"`
	Role        string    `gorm:"size:50;not null" json:"role"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

type Household struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName specifies the table name for GORM
func (Household) TableName() string {
	return "households"
}
