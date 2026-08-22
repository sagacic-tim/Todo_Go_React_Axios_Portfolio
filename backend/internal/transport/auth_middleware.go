// backend/internal/transport/auth_middleware.go
package transport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"todo_axios_api/backend/internal/models"
)

// requireAuth validates the JWT cookie and injects "currentUser" (models.User)
// into the Gin context.  Any problem results in a 401 and the request is aborted.
func requireAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(authCookieName)
		if err != nil || raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		cl := &claims{}
		tok, err := jwt.ParseWithClaims(raw, cl, func(t *jwt.Token) (interface{}, error) {
			// Reject non-HMAC algorithms to prevent algorithm-substitution attacks.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret(), nil
		})

		if err != nil || !tok.Valid {
			clearAuthCookie(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired or invalid"})
			return
		}

		// Fetch the user directly from the DB (ensures the account still exists).
		var user models.User
		if err := db.First(&user, cl.UserID).Error; err != nil {
			clearAuthCookie(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}
