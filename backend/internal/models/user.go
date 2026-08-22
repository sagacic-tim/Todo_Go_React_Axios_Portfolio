// backend/internal/models/user.go
package models

import "time"

// User represents a registered account.
// The Password field stores a bcrypt hash and is intentionally
// excluded from JSON responses (json:"-").
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"    json:"id"`
	Email     string    `gorm:"uniqueIndex;not null"         json:"email"`
	Password  string    `gorm:"not null"                     json:"-"`
	CreatedAt time.Time `                                    json:"createdAt"`
	UpdatedAt time.Time `                                    json:"updatedAt"`
}
