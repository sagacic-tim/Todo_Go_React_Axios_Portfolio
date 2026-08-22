// backend/internal/transport/security_headers.go
//
// securityHeaders sets defensive HTTP response headers on every response and
// enforces a maximum request-body size to guard against payload-flooding attacks.

package transport

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// maxBodyBytes is the largest request body we will read (32 KB).
// Callers sending larger payloads receive 413 immediately.
const maxBodyBytes = 32 * 1024 // 32 KB

// SecurityHeaders returns a Gin middleware that:
//  1. Wraps the request body with an io.LimitedReader so we never read more
//     than maxBodyBytes from any single request.
//  2. Sets a comprehensive set of security-oriented response headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ── 1. Limit request body size ────────────────────────────────────────
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		}

		// ── 2. Security response headers ──────────────────────────────────────

		h := c.Writer.Header()

		// Prevent the browser from sniffing the MIME type.
		h.Set("X-Content-Type-Options", "nosniff")

		// Deny framing entirely (clickjacking protection).
		h.Set("X-Frame-Options", "DENY")

		// Only send the origin as the referrer — no path/query leakage.
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Disable all browser features not needed by this app.
		h.Set("Permissions-Policy",
			"camera=(), microphone=(), geolocation=(), payment=(), usb=()")

		// Content-Security-Policy:
		//   default-src 'self'           – only load resources from this origin
		//   script-src  'self'           – no inline JS, no third-party scripts
		//   style-src   'self' 'unsafe-inline' – CSS modules need inline styles
		//   img-src     'self' data:     – allow data-URIs (favicons, etc.)
		//   font-src    'self'
		//   connect-src 'self'           – XHR/fetch only to this origin
		//   frame-ancestors 'none'       – belt-and-suspenders for clickjacking
		//   base-uri    'self'
		//   form-action 'self'
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"font-src 'self'; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'")

		// HSTS – tell browsers to use HTTPS for the next year.
		// (Only effective over TLS; harmless over HTTP in dev.)
		h.Set("Strict-Transport-Security",
			fmt.Sprintf("max-age=%d; includeSubDomains", 365*24*60*60))

		// Cross-Origin policies – prevent data leakage via cross-origin reads.
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")

		// Remove the Server header so we don't advertise the stack.
		h.Del("Server")

		c.Next()

		// If the body limit was exceeded, Gin's ShouldBindJSON returns a
		// "http: request body too large" error.  We surface it as 413 here
		// in case a downstream handler didn't catch it.
		if c.Writer.Status() == http.StatusOK {
			for _, e := range c.Errors {
				if e.Err == io.ErrUnexpectedEOF || isMaxBytesError(e.Err) {
					c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
						"error": "request body too large",
					})
					return
				}
			}
		}
	}
}

// isMaxBytesError detects the stdlib "http: request body too large" sentinel.
func isMaxBytesError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*http.MaxBytesError)
	return ok
}
