// backend/internal/transport/auth_handler.go
package transport

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"todo_axios_api/backend/internal/models"
)

const (
	authCookieName = "auth_token"
	jwtDuration    = 24 * time.Hour
	bcryptCost     = 12
)

// jwtSecret returns the signing secret from the environment.
// Panics at startup if JWT_SECRET is not set (caught by main).
func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		panic("JWT_SECRET environment variable is not set")
	}
	return []byte(s)
}

// claims is the JWT payload we store.
type claims struct {
	UserID uint   `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// issueJWT creates a signed HS256 token and sets the HttpOnly cookie.
func issueJWT(c *gin.Context, user models.User) error {
	now := time.Now()
	expires := now.Add(jwtDuration)

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	})

	signed, err := tok.SignedString(jwtSecret())
	if err != nil {
		return err
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     authCookieName,
		Value:    signed,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(jwtDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   os.Getenv("SESSION_COOKIE_SECURE") == "true",
	})
	return nil
}

// clearAuthCookie removes the JWT cookie from the browser.
func clearAuthCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /signup
// Body: { "email": "...", "password": "..." }
// ─────────────────────────────────────────────────────────────────────────────
func makeSignup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in struct {
			Email    string `json:"email"    binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		in.Email = strings.ToLower(strings.TrimSpace(in.Email))

		// Duplicate-email check — friendly message.
		var existing models.User
		if err := db.Where("email = ?", in.Email).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}

		user := models.User{Email: in.Email, Password: string(hash)}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
			return
		}

		if err := issueJWT(c, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"user": user})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /login
// Body: { "email": "...", "password": "..." }
// ─────────────────────────────────────────────────────────────────────────────
func makeLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in struct {
			Email    string `json:"email"    binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		in.Email = strings.ToLower(strings.TrimSpace(in.Email))

		// Account lockout check (before any DB work).
		if accountLockout.isLocked(in.Email) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "account temporarily locked due to too many failed attempts — try again in 15 minutes",
			})
			return
		}

		var user models.User
		if err := db.Where("email = ?", in.Email).First(&user).Error; err != nil {
			accountLockout.recordFailure(in.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
			accountLockout.recordFailure(in.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		accountLockout.recordSuccess(in.Email)

		if err := issueJWT(c, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /logout
// ─────────────────────────────────────────────────────────────────────────────
func makeLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		clearAuthCookie(c)
		c.Status(http.StatusNoContent)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /me  (requires valid JWT — returns current user)
// ─────────────────────────────────────────────────────────────────────────────
func makeMe() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("currentUser")
		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}
