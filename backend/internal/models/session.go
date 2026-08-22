// backend/internal/models/session.go
package models

import "time"

// Session maps a random opaque token to a user.
// Tokens are stored as bcrypt-safe random hex strings.
type Session struct {
	Token     string    `gorm:"primaryKey"  json:"-"`
	UserID    uint      `gorm:"not null"    json:"-"`
	ExpiresAt time.Time `gorm:"not null"    json:"-"`
	CreatedAt time.Time `                   json:"-"`

	// Preloaded on demand — not stored as a column.
	User User `gorm:"foreignKey:UserID" json:"-"`
}
