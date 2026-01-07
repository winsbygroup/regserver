package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
)

// Context key for CSRF token
type csrfKey struct{}

// CSRF copies the CSRF token from Echo context to request context.
// This must run AFTER Echo's CSRF middleware.
func CSRF() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Echo's CSRF middleware stores token at key "csrf"
			if token, ok := c.Get("csrf").(string); ok {
				ctx := context.WithValue(c.Request().Context(), csrfKey{}, token)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	}
}

// GetCSRF retrieves the CSRF token from context.
func GetCSRF(ctx context.Context) string {
	if token, ok := ctx.Value(csrfKey{}).(string); ok {
		return token
	}
	return ""
}
