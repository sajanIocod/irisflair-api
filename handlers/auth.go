package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	ExpiresAt int64 `json:"expiresAt"`
}

// AuthCookieName is the HttpOnly session cookie the admin app authenticates
// with. It is set by Login, cleared by Logout, and read by AuthMiddleware.
const AuthCookieName = "admin_token"

// setAuthCookie writes the HttpOnly session cookie. Secure is safe for local
// development too: browsers treat http://localhost as a trustworthy origin.
// SameSite=Lax stops cross-site POST/PUT/DELETE from carrying the cookie
// (CSRF), while same-origin admin traffic — proxied via the app's /backend
// route — always includes it.
func setAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ---- Login rate limiting (in-memory, per client IP) ----

type loginAttempt struct {
	count     int
	firstTime time.Time
}

var (
	loginAttempts = make(map[string]*loginAttempt)
	attemptsMu    sync.Mutex
)

const (
	maxLoginAttempts = 5
	lockoutDuration  = 15 * time.Minute
)

func clientIP(r *http.Request) string {
	// Honor reverse-proxy header (Render/Vercel set this)
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if ip, _, ok := splitFirstComma(fwd); ok {
			return ip
		}
		return fwd
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func splitFirstComma(s string) (string, string, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", s != ""
}

func isLockedOut(key string) bool {
	attemptsMu.Lock()
	defer attemptsMu.Unlock()

	a, ok := loginAttempts[key]
	if !ok {
		return false
	}
	if time.Since(a.firstTime) > lockoutDuration {
		delete(loginAttempts, key)
		return false
	}
	return a.count >= maxLoginAttempts
}

func recordFailedAttempt(key string) {
	attemptsMu.Lock()
	defer attemptsMu.Unlock()

	a, ok := loginAttempts[key]
	if !ok || time.Since(a.firstTime) > lockoutDuration {
		loginAttempts[key] = &loginAttempt{count: 1, firstTime: time.Now()}
		return
	}
	a.count++
}

func clearAttempts(key string) {
	attemptsMu.Lock()
	defer attemptsMu.Unlock()
	delete(loginAttempts, key)
}

// credentialsValid checks the supplied credentials using constant-time
// comparison. If ADMIN_PASSWORD_HASH (bcrypt) is set it takes precedence
// over the plain ADMIN_PASSWORD. Env values are whitespace-trimmed:
// dashboard-pasted values can carry invisible trailing newlines that make
// byte-exact comparison impossible to satisfy from a login form.
func credentialsValid(username, password string) bool {
	adminUsername := strings.TrimSpace(os.Getenv("ADMIN_USERNAME"))
	if adminUsername == "" {
		return false
	}

	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(adminUsername)) == 1

	if hash := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD_HASH")); hash != "" {
		passOK := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
		return userOK && passOK
	}

	adminPassword := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	if adminPassword == "" {
		return false
	}
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(adminPassword)) == 1
	return userOK && passOK
}

// Login authenticates an admin and returns a JWT token
func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ip := clientIP(r)
	if isLockedOut(ip) {
		log.Printf("Login blocked (rate limited) from %s", ip)
		http.Error(w, "Too many failed attempts. Try again in 15 minutes.", http.StatusTooManyRequests)
		return
	}

	if !credentialsValid(req.Username, req.Password) {
		recordFailedAttempt(ip)
		log.Printf("Failed login attempt for user %q from %s", req.Username, ip)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	clearAttempts(ip)

	// Generate JWT token
	expiresAt := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Printf("Login: failed to sign token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// The admin app authenticates via this HttpOnly cookie (invisible to JS);
	// the body token remains for curl/API tooling only.
	setAuthCookie(w, tokenString, expiresAt)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
	})
}

// Logout clears the auth cookie.
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

// Me reports the authenticated admin (auth-protected route); the app uses it
// to restore the session from the cookie on page load.
func Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": r.Header.Get("X-Username"),
	})
}

// VerifyToken verifies a JWT token and returns the claims
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Reject tokens signed with an unexpected algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
