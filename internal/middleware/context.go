package middleware

import (
	"context"

	"github.com/labstack/echo/v4"

	"winsbygroup.com/regserver/internal/version"
)

const ThemeCookieName = "theme"

// Context keys
type themeKey struct{}
type versionKey struct{}

// ValidThemes defines allowed theme values
var validThemes = map[string]bool{"light": true, "dark": true}

// Theme reads the theme cookie and adds it to the request context.
// Defaults to "light" if no cookie is set.
func Theme() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			theme := "light" // default
			cookie, err := c.Cookie(ThemeCookieName)
			if err == nil && validThemes[cookie.Value] {
				theme = cookie.Value
			}

			// Add theme to request context
			ctx := context.WithValue(c.Request().Context(), themeKey{}, theme)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// GetTheme retrieves the theme from context. Returns "light" if not set.
func GetTheme(ctx context.Context) string {
	if theme, ok := ctx.Value(themeKey{}).(string); ok {
		return theme
	}
	return "light"
}

// Version adds the app version to the request context.
func Version() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := context.WithValue(c.Request().Context(), versionKey{}, version.Version)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// GetVersion retrieves the version from context.
func GetVersion(ctx context.Context) string {
	if v, ok := ctx.Value(versionKey{}).(string); ok {
		return v
	}
	return version.Version
}

// GetRepoURL returns the project repository URL.
func GetRepoURL() string {
	return version.RepoURL
}
