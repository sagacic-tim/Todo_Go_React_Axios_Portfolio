// backend/internal/transport/rate_limiter.go
//
// Provides two independent in-memory mechanisms (no external libraries):
//
//  1. IP-based sliding-window rate limiter
//     – Global routes:  120 requests / 60 s per IP
//     – Auth routes:    10  requests / 60 s per IP  (brute-force deterrent)
//
//  2. Email-based login lockout
//     – After 5 consecutive failures the account is locked for 15 minutes.
//     – A successful login resets the counter.

package transport

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ─── IP rate limiter ──────────────────────────────────────────────────────────

type windowEntry struct {
	timestamps []time.Time
}

type ipRateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	maxReqs  int
	entries  map[string]*windowEntry
	lastPrune time.Time
}

func newIPRateLimiter(maxReqs int, window time.Duration) *ipRateLimiter {
	return &ipRateLimiter{
		window:    window,
		maxReqs:   maxReqs,
		entries:   make(map[string]*windowEntry),
		lastPrune: time.Now(),
	}
}

// allow returns true when the request is within the rate limit.
func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Periodic cleanup so the map doesn't grow forever.
	if now.Sub(rl.lastPrune) > 5*time.Minute {
		for k, e := range rl.entries {
			if len(e.timestamps) == 0 || e.timestamps[len(e.timestamps)-1].Before(cutoff) {
				delete(rl.entries, k)
			}
		}
		rl.lastPrune = now
	}

	e, ok := rl.entries[ip]
	if !ok {
		e = &windowEntry{}
		rl.entries[ip] = e
	}

	// Slide the window: drop timestamps older than the cutoff.
	valid := e.timestamps[:0]
	for _, t := range e.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	e.timestamps = append(valid, now)

	return len(e.timestamps) <= rl.maxReqs
}

// realIP extracts the client IP, honouring X-Forwarded-For when present.
func realIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first (leftmost) address – that is the original client.
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	ip := c.ClientIP()
	return ip
}

// Middleware returns a Gin middleware that enforces this rate limiter.
func (rl *ipRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := realIP(c)
		if !rl.allow(ip) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests — please slow down and try again shortly",
			})
			return
		}
		c.Next()
	}
}

// ─── Shared limiter instances (created once in NewRouter) ─────────────────────

var (
	globalLimiter *ipRateLimiter // 120 req / 60 s — all routes
	authLimiter   *ipRateLimiter // 10  req / 60 s — auth routes only
)

func initRateLimiters() {
	globalLimiter = newIPRateLimiter(120, 60*time.Second)
	authLimiter   = newIPRateLimiter(10,  60*time.Second)
}

// ─── Account lockout ──────────────────────────────────────────────────────────

const (
	maxLoginAttempts  = 5
	lockoutDuration   = 15 * time.Minute
)

type lockoutEntry struct {
	failures  int
	lockedAt  time.Time
	locked    bool
}

type loginLockout struct {
	mu      sync.Mutex
	entries map[string]*lockoutEntry
}

var accountLockout = &loginLockout{
	entries: make(map[string]*lockoutEntry),
}

// isLocked returns true when the email is currently locked out.
func (l *loginLockout) isLocked(email string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[email]
	if !ok {
		return false
	}
	if e.locked && time.Since(e.lockedAt) >= lockoutDuration {
		// Lockout expired — reset automatically.
		delete(l.entries, email)
		return false
	}
	return e.locked
}

// recordFailure increments the failure counter and locks the account if the
// threshold is reached.
func (l *loginLockout) recordFailure(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[email]
	if !ok {
		e = &lockoutEntry{}
		l.entries[email] = e
	}
	e.failures++
	if e.failures >= maxLoginAttempts {
		e.locked   = true
		e.lockedAt = time.Now()
	}
}

// recordSuccess clears failure state on a good login.
func (l *loginLockout) recordSuccess(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, email)
}
