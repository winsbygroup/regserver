package client

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes wires all client-facing endpoints under the given Echo group.
// The licKeyAuth middleware is applied only to endpoints that require license key validation.
func RegisterRoutes(g *echo.Group, h *Handler, licKeyAuth echo.MiddlewareFunc) {

	// Activation endpoint (requires license key)
	g.POST("/activate", h.Activate, licKeyAuth)

	// Product version lookup (public, no auth required)
	g.GET("/productver/:guid", h.GetProductVersion)

	// License info lookup (public, no auth required - license key is in URL)
	g.GET("/license/:license_key", h.GetLicenseInfo)

	// Update machine's installed version (public, no auth required - license key is in URL)
	g.PUT("/license/:license_key", h.UpdateLicenseInfo)
}
