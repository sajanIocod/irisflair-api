package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/irisflair/api/handlers"
)

// AuthMiddleware checks for valid JWT token in Authorization header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Verify token
		claims, err := handlers.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Store claims in request context for later use if needed
		if username, ok := claims["username"].(string); ok {
			r.Header.Set("X-Username", username)
		}

		next.ServeHTTP(w, r)
	})
}

var (
	allowedOriginsOnce sync.Once
	allowedOrigins     map[string]bool
	allowAllOrigins    bool
)

// loadAllowedOrigins parses the ALLOWED_ORIGINS env var (comma-separated).
// If unset, all origins are allowed with a loud warning (avoids breaking
// existing deployments — set ALLOWED_ORIGINS in production!).
func loadAllowedOrigins() {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		allowAllOrigins = true
		log.Println("WARNING: ALLOWED_ORIGINS not set — CORS is open to all origins. Set ALLOWED_ORIGINS=https://yourdomain.com,... in production!")
		return
	}
	allowedOrigins = make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(strings.TrimSuffix(o, "/"))
		if o != "" {
			allowedOrigins[o] = true
		}
	}
	log.Printf("✓ CORS restricted to %d origin(s)", len(allowedOrigins))
}

// CORSMiddleware adds CORS headers, restricted to origins in ALLOWED_ORIGINS.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOriginsOnce.Do(loadAllowedOrigins)

		origin := r.Header.Get("Origin")
		if allowAllOrigins {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24h

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware sets standard security headers on every response.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		// Only meaningful over HTTPS; harmless otherwise
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// ErrorRecoveryMiddleware recovers from panics and returns 500 error
func ErrorRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered on %s %s: %v", r.Method, r.URL.Path, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
