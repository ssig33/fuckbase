package server

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/ssig33/fuckbase/internal/config"
	"github.com/ssig33/fuckbase/internal/logger"
)

// AdminAuth handles admin authentication
type AdminAuth struct {
	Config *config.AdminAuthConfig
}

// NewAdminAuth creates a new admin authentication handler
func NewAdminAuth(cfg *config.AdminAuthConfig) *AdminAuth {
	return &AdminAuth{
		Config: cfg,
	}
}

// Authenticate authenticates an admin user
func (a *AdminAuth) Authenticate(username, password string) bool {
	if a.Config == nil || !a.Config.Enabled {
		return true
	}

	return a.Config.Username == username && a.Config.Password == password
}

// RequireAdminAuth is a middleware that requires admin authentication
func (a *AdminAuth) RequireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.Config == nil || !a.Config.Enabled {
			next(w, r)
			return
		}

		// Check for admin auth in header
		authHeader := r.Header.Get("X-Admin-Authorization")
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Basic ") {
				credentials, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
				if err == nil {
					parts := strings.SplitN(string(credentials), ":", 2)
					if len(parts) == 2 {
						username := parts[0]
						password := parts[1]
						if a.Authenticate(username, password) {
							next(w, r)
							return
						}
					}
				}
			}
		}

		// If we get here, authentication failed
		logger.Warn("Admin authentication failed")
		writeErrorResponse(w, http.StatusUnauthorized, "ADMIN_AUTH_REQUIRED", "Admin authentication required")
	}
}

// ExtractDatabaseAuth extracts database authentication from a request
func ExtractDatabaseAuth(r *http.Request) (string, string, bool) {
	// Check for auth in header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Basic ") {
			credentials, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
			if err == nil {
				parts := strings.SplitN(string(credentials), ":", 2)
				if len(parts) == 2 {
					return parts[0], parts[1], true
				}
			}
		}
	}

	return "", "", false
}