package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

const SessionCookieName = "regadmin_session"

// LicenseContext holds customer and product IDs extracted from license key
type LicenseContext struct {
	CustomerID int64 `db:"customer_id"`
	ProductID  int64 `db:"product_id"`
}

// LicenseKeyAuth validates the X-License-Key header against
// license records. Used for CLIENT endpoints.
func LicenseKeyAuth(db *sqlx.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			licKey := c.Request().Header.Get("X-License-Key")
			if licKey == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing license key")
			}

			var lic LicenseContext
			err := db.Get(&lic, `
                SELECT customer_id, product_id
                FROM license
                WHERE license_key = ?
            `, licKey)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid license key")
			}

			// Attach to context
			c.Set("license", lic)
			return next(c)
		}
	}
}

// AdminAPIKeyAuth validates the X-API-Key header against ADMIN_API_KEY.
// Used for ADMIN API endpoints. Returns 401 if authentication fails.
func AdminAPIKeyAuth() echo.MiddlewareFunc {
	adminKey := os.Getenv("ADMIN_API_KEY")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if adminKey == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "ADMIN_API_KEY environment variable not configured")
			}

			key := c.Request().Header.Get("X-API-Key")
			if key == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing admin API key")
			}

			if key != adminKey {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid admin API key")
			}

			return next(c)
		}
	}
}

// WebAuth validates requests via X-API-Key header OR session cookie.
// Used for WEB UI endpoints. Redirects to login if authentication fails.
func WebAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()

			// Public routes: login page and static assets
			if path == "/web/login" ||
				strings.HasPrefix(path, "/web/static/") {
				return next(c)
			}

			// Check X-API-Key header first (for programmatic access)
			if key := c.Request().Header.Get("X-API-Key"); key != "" && ValidateAdminKey(key) {
				return next(c)
			}

			// Check session cookie
			if cookie, err := c.Cookie(SessionCookieName); err == nil {
				if sessionID := cookie.Value; sessionID != "" {
					if _, ok := GetSession(sessionID); ok {
						return next(c)
					}
				}
			}

			// Not authenticated - redirect to login
			return c.Redirect(http.StatusFound, "/web/login")
		}
	}
}

// ValidateAdminKey checks if the provided key matches ADMIN_API_KEY
// using constant-time comparison to prevent timing attacks.
func ValidateAdminKey(key string) bool {
	adminKey := os.Getenv("ADMIN_API_KEY")
	if adminKey == "" {
		return false
	}
	return constantEqual(adminKey, key)
}

// constantEqual provides constant-time string equality to avoid timing attacks.
func constantEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
