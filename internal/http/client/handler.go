package client

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"winsbygroup.com/regserver/internal/activation"
	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/middleware"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/registration"
)

type Handler struct {
	ActivationService   *activation.Service
	RegistrationService *registration.Service
	ProductService      *product.Service
	LicenseService      *license.Service
	MachineService      *machine.Service
	FeatureService      *feature.Service
	FeatureValueService *featurevalue.Service
	CustomerService     *customer.Service
}

func NewHandler(
	a *activation.Service,
	r *registration.Service,
	p *product.Service,
	l *license.Service,
	m *machine.Service,
	f *feature.Service,
	fv *featurevalue.Service,
	c *customer.Service,
) *Handler {
	return &Handler{
		ActivationService:   a,
		RegistrationService: r,
		ProductService:      p,
		LicenseService:      l,
		MachineService:      m,
		FeatureService:      f,
		FeatureValueService: fv,
		CustomerService:     c,
	}
}

// POST /activate
func (h *Handler) Activate(c echo.Context) error {
	var req activation.Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Get customerID and productID from context (set by LicenseKeyAuth middleware)
	lic, ok := c.Get("license").(middleware.LicenseContext)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid license context",
		})
	}

	resp, err := h.ActivationService.Activate(
		c.Request().Context(),
		lic.CustomerID,
		lic.ProductID,
		&req,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// GET /productver/:guid
func (h *Handler) GetProductVersion(c echo.Context) error {
	guid := c.Param("guid")
	if guid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing product guid",
		})
	}

	prod, err := h.ProductService.GetByGUID(c.Request().Context(), guid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ProductGUID":   prod.ProductGUID,
		"LatestVersion": prod.LatestVersion,
		"DownloadURL":   prod.DownloadURL,
	})
}

// LicenseInfoResponse is the response for the license info endpoint
type LicenseInfoResponse struct {
	CustomerName        string         `json:"CustomerName"`
	ProductGUID         string         `json:"ProductGUID"`
	ProductName         string         `json:"ProductName"`
	LicenseCount        int            `json:"LicenseCount"`
	LicensesAvailable   int            `json:"LicensesAvailable"`
	ExpirationDate      string         `json:"ExpirationDate"`
	MaintExpirationDate string         `json:"MaintExpirationDate"`
	MaxProductVersion   string         `json:"MaxProductVersion"`
	LatestVersion       string         `json:"LatestVersion"`
	Features            map[string]any `json:"Features"`
}

// GET /license/:license_key
func (h *Handler) GetLicenseInfo(c echo.Context) error {
	licenseKey := c.Param("license_key")
	if licenseKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing license key",
		})
	}

	ctx := c.Request().Context()

	// Get the license by key
	lic, err := h.LicenseService.GetByLicenseKey(ctx, licenseKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "license not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Get product info
	prod, err := h.ProductService.Get(ctx, lic.ProductID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Get customer info
	cust, err := h.CustomerService.Get(ctx, lic.CustomerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Get active (non-expired) machine registrations count
	activeMachines, err := h.MachineService.GetActiveForLicense(ctx, lic.CustomerID, lic.ProductID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	licensesAvailable := lic.LicenseCount - len(activeMachines)
	if licensesAvailable < 0 {
		licensesAvailable = 0
	}

	// Get feature values (merged with defaults)
	features, err := h.mergeFeatures(ctx, lic.CustomerID, lic.ProductID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, LicenseInfoResponse{
		CustomerName:        cust.CustomerName,
		ProductGUID:         prod.ProductGUID,
		ProductName:         prod.ProductName,
		LicenseCount:        lic.LicenseCount,
		LicensesAvailable:   licensesAvailable,
		ExpirationDate:      lic.ExpirationDate,
		MaintExpirationDate: lic.MaintExpirationDate,
		MaxProductVersion:   lic.MaxProductVersion,
		LatestVersion:       prod.LatestVersion,
		Features:            features,
	})
}

// UpdateLicenseRequest is the request body for updating license/machine info
type UpdateLicenseRequest struct {
	MachineCode      string `json:"machineCode"`
	InstalledVersion string `json:"installedVersion"`
}

// PUT /license/:license_key
func (h *Handler) UpdateLicenseInfo(c echo.Context) error {
	licenseKey := c.Param("license_key")
	if licenseKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing license key",
		})
	}

	var req UpdateLicenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if req.MachineCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "machineCode is required",
		})
	}

	ctx := c.Request().Context()

	// Get the license by key
	lic, err := h.LicenseService.GetByLicenseKey(ctx, licenseKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "license not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Find the machine by code
	machine, err := h.MachineService.GetByCode(ctx, lic.CustomerID, req.MachineCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	if machine == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "machine not found for this license",
		})
	}

	// Update installed version if provided
	if req.InstalledVersion != "" {
		err = h.RegistrationService.UpdateInstalledVersion(ctx, machine.MachineID, lic.ProductID, req.InstalledVersion)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "no registration found for this machine and product",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
	}

	// Return the same response as GET /license/:license_key
	prod, err := h.ProductService.Get(ctx, lic.ProductID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	cust, err := h.CustomerService.Get(ctx, lic.CustomerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	activeMachines, err := h.MachineService.GetActiveForLicense(ctx, lic.CustomerID, lic.ProductID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	licensesAvailable := lic.LicenseCount - len(activeMachines)
	if licensesAvailable < 0 {
		licensesAvailable = 0
	}

	features, err := h.mergeFeatures(ctx, lic.CustomerID, lic.ProductID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, LicenseInfoResponse{
		CustomerName:        cust.CustomerName,
		ProductGUID:         prod.ProductGUID,
		ProductName:         prod.ProductName,
		LicenseCount:        lic.LicenseCount,
		LicensesAvailable:   licensesAvailable,
		ExpirationDate:      lic.ExpirationDate,
		MaintExpirationDate: lic.MaintExpirationDate,
		MaxProductVersion:   lic.MaxProductVersion,
		LatestVersion:       prod.LatestVersion,
		Features:            features,
	})
}

// mergeFeatures returns feature values with customer overrides applied to defaults
func (h *Handler) mergeFeatures(ctx context.Context, customerID, productID int64) (map[string]any, error) {
	defs, err := h.FeatureService.GetForProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	vals, err := h.FeatureValueService.GetFeatureValues(ctx, customerID, productID)
	if err != nil {
		return nil, err
	}

	return feature.MergeWithOverrides(defs, vals), nil
}
