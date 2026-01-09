package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"winsbygroup.com/regserver/internal/middleware"

	"winsbygroup.com/regserver/internal/activation"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/http/admin"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/registration"
	"winsbygroup.com/regserver/internal/sqlite"
	vm "winsbygroup.com/regserver/internal/viewmodels"
	"winsbygroup.com/regserver/templates/components"
	"winsbygroup.com/regserver/templates/pages"
)

// errInvalidVersion is a sentinel error for invalid version format
var errInvalidVersion = fmt.Errorf("invalid version format")

// errInvalidSubscriptionTerm is a sentinel error for subscription licenses without a valid term
var errInvalidSubscriptionTerm = fmt.Errorf("subscription requires term > 0")

// Handler handles web UI requests
type Handler struct {
	svc           *admin.Service
	productSvc    *product.Service
	featureSvc    *feature.Service
	featureValSvc *featurevalue.Service
	machineSvc    *machine.Service
	regSvc        *registration.Service
	activationSvc *activation.Service
}

// NewHandler creates a new web handler
func NewHandler(
	svc *admin.Service,
	productSvc *product.Service,
	featureSvc *feature.Service,
	featureValSvc *featurevalue.Service,
	machineSvc *machine.Service,
	regSvc *registration.Service,
	activationSvc *activation.Service,
) *Handler {
	return &Handler{
		svc:           svc,
		productSvc:    productSvc,
		featureSvc:    featureSvc,
		featureValSvc: featureValSvc,
		machineSvc:    machineSvc,
		regSvc:        regSvc,
		activationSvc: activationSvc,
	}
}

// getCustomerName returns the customer name for the given ID, or empty string if not found
func (h *Handler) getCustomerName(ctx context.Context, customerID int64) string {
	cust, err := h.svc.GetCustomer(ctx, customerID)
	if err != nil || cust == nil {
		return ""
	}
	return cust.CustomerName
}

// --------------------------
// Authentication
// --------------------------

// LoginPage renders the login form
func (h *Handler) LoginPage(c echo.Context) error {
	return pages.Login("").Render(c.Request().Context(), c.Response())
}

// Login handles login form submission
func (h *Handler) Login(c echo.Context) error {
	apiKey := c.FormValue("api_key")

	if !middleware.ValidateAdminKey(apiKey) {
		return pages.Login("Invalid API key").Render(c.Request().Context(), c.Response())
	}

	// Create a new session (does NOT store the admin key in the cookie)
	sessionID := middleware.CreateSession()

	cookie := &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,                    // always true behind Caddy
		SameSite: http.SameSiteStrictMode, // admin-only
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	}
	c.SetCookie(cookie)

	return c.Redirect(http.StatusFound, "/web/")
}

// Logout clears the session cookie and deletes the server-side session
func (h *Handler) Logout(c echo.Context) error {
	// Delete server-side session
	if cookie, err := c.Cookie(middleware.SessionCookieName); err == nil {
		if sessionID := cookie.Value; sessionID != "" {
			middleware.DeleteSession(sessionID)
		}
	}

	// Overwrite cookie with expired one
	cookie := &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	c.SetCookie(cookie)

	return c.Redirect(http.StatusFound, "/web/login")
}

// --------------------------
// Customers
// --------------------------

func (h *Handler) ListCustomers(c echo.Context) error {
	ctx := c.Request().Context()
	customers, err := h.svc.GetCustomers(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewCustomers := FromDomainCustomers(customers)
	if isHTMX(c) {
		return components.CustomersTable(viewCustomers).Render(ctx, c.Response())
	}
	return pages.Customers(viewCustomers).Render(ctx, c.Response())
}

func (h *Handler) NewCustomerForm(c echo.Context) error {
	return components.CustomerForm(nil).Render(c.Request().Context(), c.Response())
}

func (h *Handler) EditCustomerForm(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}

	cust, err := h.svc.GetCustomer(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Customer not found")
	}

	viewCustomer := FromDomainCustomer(*cust)
	return components.CustomerForm(&viewCustomer).Render(ctx, c.Response())
}

func (h *Handler) CreateCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	req := &admin.CreateCustomerRequest{
		CustomerName: strings.TrimSpace(c.FormValue("customer_name")),
		ContactName:  strings.TrimSpace(c.FormValue("contact_name")),
		Phone:        strings.TrimSpace(c.FormValue("phone")),
		Email:        strings.TrimSpace(c.FormValue("email")),
		Notes:        strings.TrimSpace(c.FormValue("notes")),
	}

	if _, err := h.svc.CreateCustomer(ctx, req); err != nil {
		customer := &vm.Customer{
			CustomerName: req.CustomerName,
			ContactName:  req.ContactName,
			Phone:        req.Phone,
			Email:        req.Email,
			Notes:        req.Notes,
		}
		return h.renderCustomerFormWithError(c, ctx, customer, err)
	}

	customers, err := h.svc.GetCustomers(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"closeModal": true, "showToast": {"message": "Customer created successfully", "type": "success"}}`)
	return components.CustomersTable(FromDomainCustomers(customers)).Render(ctx, c.Response())
}

func (h *Handler) UpdateCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}

	req := &admin.UpdateCustomerRequest{
		CustomerName: strings.TrimSpace(c.FormValue("customer_name")),
		ContactName:  strings.TrimSpace(c.FormValue("contact_name")),
		Phone:        strings.TrimSpace(c.FormValue("phone")),
		Email:        strings.TrimSpace(c.FormValue("email")),
		Notes:        strings.TrimSpace(c.FormValue("notes")),
	}

	if err := h.svc.UpdateCustomer(ctx, id, req); err != nil {
		customer := &vm.Customer{
			CustomerID:   id,
			CustomerName: req.CustomerName,
			ContactName:  req.ContactName,
			Phone:        req.Phone,
			Email:        req.Email,
			Notes:        req.Notes,
		}
		return h.renderCustomerFormWithError(c, ctx, customer, err)
	}

	customers, err := h.svc.GetCustomers(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"closeModal": true, "showToast": {"message": "Customer updated successfully", "type": "success"}}`)
	return components.CustomersTable(FromDomainCustomers(customers)).Render(ctx, c.Response())
}

// renderCustomerFormWithError re-renders the customer form with appropriate field errors
func (h *Handler) renderCustomerFormWithError(c echo.Context, ctx context.Context, customer *vm.Customer, err error) error {
	errors := make(map[string]string)

	switch {
	case sqlite.IsUniqueConstraintError(err):
		errors["customer_name"] = "A customer with this name already exists"
	default:
		// Unknown error - show toast instead
		setTriggerWithData(c, fmt.Sprintf(`{"showToast": {"message": %q, "type": "error"}}`, "Failed to save customer"))
		return c.String(http.StatusUnprocessableEntity, "")
	}

	formData := components.CustomerFormData{Customer: customer, Errors: errors}
	c.Response().Header().Set("HX-Retarget", "#modal-content")
	c.Response().Header().Set("HX-Reswap", "innerHTML")
	return components.CustomerFormWithErrors(formData).Render(ctx, c.Response())
}

func (h *Handler) DeleteCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}

	if err := h.svc.DeleteCustomer(ctx, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	customers, err := h.svc.GetCustomers(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"showToast": {"message": "Customer deleted successfully", "type": "success"}}`)
	return components.CustomersTable(FromDomainCustomers(customers)).Render(ctx, c.Response())
}

// --------------------------
// Products
// --------------------------

func (h *Handler) ListProducts(c echo.Context) error {
	ctx := c.Request().Context()
	products, err := h.svc.GetProducts(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewProducts := FromDomainProducts(products)
	if isHTMX(c) {
		return components.ProductsTable(viewProducts).Render(ctx, c.Response())
	}
	return pages.Products(viewProducts).Render(ctx, c.Response())
}

func (h *Handler) NewProductForm(c echo.Context) error {
	data := components.ProductFormData{
		DefaultGUID: uuid.NewString(),
	}
	return components.ProductFormWithErrors(data).Render(c.Request().Context(), c.Response())
}

func (h *Handler) EditProductForm(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	prod, err := h.svc.GetProduct(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Product not found")
	}

	viewProduct := FromDomainProduct(*prod)
	return components.ProductForm(&viewProduct).Render(ctx, c.Response())
}

func (h *Handler) CreateProduct(c echo.Context) error {
	ctx := c.Request().Context()
	req := &admin.CreateProductRequest{
		Name:          strings.TrimSpace(c.FormValue("product_name")),
		Guid:          strings.TrimSpace(c.FormValue("product_guid")),
		LatestVersion: strings.TrimSpace(c.FormValue("latest_version")),
		DownloadURL:   strings.TrimSpace(c.FormValue("download_url")),
	}

	if _, err := h.svc.CreateProduct(ctx, req); err != nil {
		product := &vm.Product{
			ProductName:   req.Name,
			ProductGUID:   req.Guid,
			LatestVersion: req.LatestVersion,
			DownloadURL:   req.DownloadURL,
		}
		return h.renderProductFormWithError(c, ctx, product, err)
	}

	products, err := h.svc.GetProducts(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"closeModal": true, "showToast": {"message": "Product created successfully", "type": "success"}}`)
	return components.ProductsTable(FromDomainProducts(products)).Render(ctx, c.Response())
}

func (h *Handler) UpdateProduct(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	req := &admin.UpdateProductRequest{
		Name:          strings.TrimSpace(c.FormValue("product_name")),
		Guid:          strings.TrimSpace(c.FormValue("product_guid")),
		LatestVersion: strings.TrimSpace(c.FormValue("latest_version")),
		DownloadURL:   strings.TrimSpace(c.FormValue("download_url")),
	}

	if err := h.svc.UpdateProduct(ctx, id, req); err != nil {
		product := &vm.Product{
			ProductID:     id,
			ProductName:   req.Name,
			ProductGUID:   req.Guid,
			LatestVersion: req.LatestVersion,
			DownloadURL:   req.DownloadURL,
		}
		return h.renderProductFormWithError(c, ctx, product, err)
	}

	products, err := h.svc.GetProducts(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"closeModal": true, "showToast": {"message": "Product updated successfully", "type": "success"}}`)
	return components.ProductsTable(FromDomainProducts(products)).Render(ctx, c.Response())
}

// renderProductFormWithError re-renders the product form with appropriate field errors
func (h *Handler) renderProductFormWithError(c echo.Context, ctx context.Context, product *vm.Product, err error) error {
	errors := make(map[string]string)

	switch {
	case strings.Contains(err.Error(), "latest version"):
		errors["latest_version"] = "Must be empty or in #.#.# format (e.g., 1.0.0)"
	case sqlite.IsUniqueConstraintError(err):
		errors["product_guid"] = "A product with this GUID already exists"
	default:
		// Unknown error - show toast instead
		setTriggerWithData(c, fmt.Sprintf(`{"showToast": {"message": %q, "type": "error"}}`, "Failed to save product"))
		return c.String(http.StatusUnprocessableEntity, "")
	}

	formData := components.ProductFormData{Product: product, Errors: errors}
	c.Response().Header().Set("HX-Retarget", "#modal-content")
	c.Response().Header().Set("HX-Reswap", "innerHTML")
	return components.ProductFormWithErrors(formData).Render(ctx, c.Response())
}

func (h *Handler) DeleteProduct(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	if err := h.svc.DeleteProduct(ctx, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	products, err := h.svc.GetProducts(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"showToast": {"message": "Product deleted successfully", "type": "success"}}`)
	return components.ProductsTable(FromDomainProducts(products)).Render(ctx, c.Response())
}

// --------------------------
// Product Features (definitions)
// --------------------------

func (h *Handler) ProductFeaturesManager(c echo.Context) error {
	ctx := c.Request().Context()
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	prod, err := h.svc.GetProduct(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Product not found")
	}

	features, err := h.svc.GetFeatures(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewProduct := FromDomainProduct(*prod)
	viewFeatures := FromDomainFeatures(features)
	return components.ProductFeaturesManager(&viewProduct, viewFeatures).Render(ctx, c.Response())
}

func (h *Handler) NewFeatureForm(c echo.Context) error {
	ctx := c.Request().Context()
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	return components.FeatureForm(nil, productID).Render(ctx, c.Response())
}

func (h *Handler) EditFeatureForm(c echo.Context) error {
	ctx := c.Request().Context()
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}
	featureID, err := strconv.ParseInt(c.Param("featureId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid feature ID")
	}

	features, err := h.svc.GetFeatures(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	for _, f := range features {
		if f.FeatureID == featureID {
			viewFeature := FromDomainFeature(f)
			return components.FeatureForm(&viewFeature, productID).Render(ctx, c.Response())
		}
	}

	return echo.NewHTTPError(http.StatusNotFound, "Feature not found")
}

func (h *Handler) CreateFeature(c echo.Context) error {
	ctx := c.Request().Context()
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	req := &admin.CreateFeatureRequest{
		FeatureName:   c.FormValue("feature_name"),
		FeatureType:   c.FormValue("feature_type"),
		AllowedValues: c.FormValue("allowed_values"),
		DefaultValue:  c.FormValue("default_value"),
	}

	// Validate allowed_values for Values type
	if req.FeatureType == "values" && !strings.Contains(req.AllowedValues, "|") {
		feature := &vm.Feature{
			FeatureName:   req.FeatureName,
			FeatureType:   featureTypeFromString(req.FeatureType),
			AllowedValues: req.AllowedValues,
			DefaultValue:  req.DefaultValue,
		}
		return h.renderFeatureFormWithError(c, ctx, feature, productID, "allowed_values", "Values type requires at least two pipe-separated options (e.g., Yes|No)")
	}

	if _, err := h.svc.CreateFeature(ctx, productID, req); err != nil {
		feature := &vm.Feature{
			FeatureName:   req.FeatureName,
			FeatureType:   featureTypeFromString(req.FeatureType),
			AllowedValues: req.AllowedValues,
			DefaultValue:  req.DefaultValue,
		}
		if sqlite.IsUniqueConstraintError(err) {
			return h.renderFeatureFormWithError(c, ctx, feature, productID, "feature_name", "A feature with this name already exists for this product")
		}
		return h.renderFeatureFormWithError(c, ctx, feature, productID, "", err.Error())
	}

	// Return updated features manager
	return h.ProductFeaturesManager(c)
}

func (h *Handler) UpdateFeature(c echo.Context) error {
	ctx := c.Request().Context()
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}
	featureID, err := strconv.ParseInt(c.Param("featureId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid feature ID")
	}

	req := &admin.UpdateFeatureRequest{
		FeatureName:   c.FormValue("feature_name"),
		FeatureType:   c.FormValue("feature_type"),
		AllowedValues: c.FormValue("allowed_values"),
		DefaultValue:  c.FormValue("default_value"),
	}

	// Validate allowed_values for Values type
	if req.FeatureType == "values" && !strings.Contains(req.AllowedValues, "|") {
		feature := &vm.Feature{
			FeatureID:     featureID,
			FeatureName:   req.FeatureName,
			FeatureType:   featureTypeFromString(req.FeatureType),
			AllowedValues: req.AllowedValues,
			DefaultValue:  req.DefaultValue,
		}
		return h.renderFeatureFormWithError(c, ctx, feature, productID, "allowed_values", "Values type requires at least two pipe-separated options (e.g., Yes|No)")
	}

	if err := h.svc.UpdateFeature(ctx, featureID, req); err != nil {
		feature := &vm.Feature{
			FeatureID:     featureID,
			FeatureName:   req.FeatureName,
			FeatureType:   featureTypeFromString(req.FeatureType),
			AllowedValues: req.AllowedValues,
			DefaultValue:  req.DefaultValue,
		}
		if sqlite.IsUniqueConstraintError(err) {
			return h.renderFeatureFormWithError(c, ctx, feature, productID, "feature_name", "A feature with this name already exists for this product")
		}
		return h.renderFeatureFormWithError(c, ctx, feature, productID, "", err.Error())
	}

	// Return updated features manager
	prod, err := h.svc.GetProduct(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Product not found")
	}

	features, err := h.svc.GetFeatures(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewProduct := FromDomainProduct(*prod)
	viewFeatures := FromDomainFeatures(features)
	return components.ProductFeaturesManager(&viewProduct, viewFeatures).Render(ctx, c.Response())
}

// featureTypeFromString converts string to vm.FeatureType
func featureTypeFromString(s string) vm.FeatureType {
	switch s {
	case "integer":
		return vm.FeatureTypeInteger
	case "string":
		return vm.FeatureTypeString
	case "values":
		return vm.FeatureTypeValues
	default:
		return vm.FeatureTypeInteger
	}
}

// renderFeatureFormWithError re-renders the feature form with appropriate field errors
func (h *Handler) renderFeatureFormWithError(c echo.Context, ctx context.Context, feature *vm.Feature, productID int64, field, message string) error {
	errors := make(map[string]string)
	if field != "" {
		errors[field] = message
	} else {
		// Unknown error - show toast instead
		setTriggerWithData(c, fmt.Sprintf(`{"showToast": {"message": %q, "type": "error"}}`, message))
		return c.String(http.StatusUnprocessableEntity, "")
	}

	formData := components.FeatureFormData{
		Feature:   feature,
		ProductID: productID,
		Errors:    errors,
	}
	return components.FeatureFormWithErrors(formData).Render(ctx, c.Response())
}

func (h *Handler) DeleteFeature(c echo.Context) error {
	ctx := c.Request().Context()
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}
	featureID, err := strconv.ParseInt(c.Param("featureId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid feature ID")
	}

	if err := h.svc.DeleteFeature(ctx, featureID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Return updated features manager
	prod, err := h.svc.GetProduct(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Product not found")
	}

	features, err := h.svc.GetFeatures(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewProduct := FromDomainProduct(*prod)
	viewFeatures := FromDomainFeatures(features)
	return components.ProductFeaturesManager(&viewProduct, viewFeatures).Render(ctx, c.Response())
}

// --------------------------
// Licenses
// --------------------------

func (h *Handler) GetLicenses(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}

	lics, err := h.svc.GetLicenses(ctx, customerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewLics := h.convertLicenses(ctx, lics)
	return components.LicensesTable(customerID, h.getCustomerName(ctx, customerID), viewLics).Render(ctx, c.Response())
}

func (h *Handler) NewLicenseForm(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}

	products, err := h.svc.GetUnlicensedProducts(ctx, customerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return components.LicenseForm(nil, customerID, FromDomainProducts(products)).Render(ctx, c.Response())
}

func (h *Handler) EditLicenseForm(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	lics, err := h.svc.GetLicenses(ctx, customerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var lic *license.License
	for i, item := range lics {
		if item.ProductID == productID {
			lic = &lics[i]
			break
		}
	}

	if lic == nil {
		return echo.NewHTTPError(http.StatusNotFound, "License not found")
	}

	prod, _ := h.svc.GetProduct(ctx, productID)
	products := []product.Product{}
	if prod != nil {
		products = append(products, *prod)
	}

	viewLic := FromDomainLicense(*lic, prod.ProductName)
	return components.LicenseForm(&viewLic, customerID, FromDomainProducts(products)).Render(ctx, c.Response())
}

func (h *Handler) CreateLicense(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}

	productID, _ := strconv.ParseInt(c.FormValue("product_id"), 10, 64)
	licenseCount, _ := strconv.Atoi(c.FormValue("license_count"))
	licenseTerm, _ := strconv.Atoi(c.FormValue("license_term"))
	isSubscription := c.FormValue("license_type") == "subscription"

	req := &admin.CreateLicenseRequest{
		ProductID:           productID,
		LicenseCount:        licenseCount,
		IsSubscription:      isSubscription,
		LicenseTerm:         licenseTerm,
		StartDate:           c.FormValue("start_date"),
		ExpirationDate:      c.FormValue("expiration_date"),
		MaintExpirationDate: c.FormValue("maint_expiration_date"),
		MaxProductVersion:   strings.TrimSpace(c.FormValue("max_product_version")),
	}

	// Validate subscription requires term > 0
	if isSubscription && licenseTerm <= 0 {
		license := &vm.License{
			ProductID:           productID,
			LicenseCount:        licenseCount,
			IsSubscription:      isSubscription,
			LicenseTerm:         licenseTerm,
			StartDate:           req.StartDate,
			ExpirationDate:      req.ExpirationDate,
			MaintExpirationDate: req.MaintExpirationDate,
			MaxProductVersion:   req.MaxProductVersion,
		}
		return h.renderLicenseFormWithError(c, ctx, license, customerID, errInvalidSubscriptionTerm)
	}

	// Validate MaxProductVersion format
	if !product.IsValidVersion(req.MaxProductVersion) {
		license := &vm.License{
			ProductID:           productID,
			LicenseCount:        licenseCount,
			IsSubscription:      isSubscription,
			LicenseTerm:         licenseTerm,
			StartDate:           req.StartDate,
			ExpirationDate:      req.ExpirationDate,
			MaintExpirationDate: req.MaintExpirationDate,
			MaxProductVersion:   req.MaxProductVersion,
		}
		return h.renderLicenseFormWithError(c, ctx, license, customerID, errInvalidVersion)
	}

	if _, err := h.svc.CreateLicense(ctx, customerID, req); err != nil {
		license := &vm.License{
			ProductID:           productID,
			LicenseCount:        licenseCount,
			IsSubscription:      isSubscription,
			LicenseTerm:         licenseTerm,
			StartDate:           req.StartDate,
			ExpirationDate:      req.ExpirationDate,
			MaintExpirationDate: req.MaintExpirationDate,
			MaxProductVersion:   req.MaxProductVersion,
		}
		return h.renderLicenseFormWithError(c, ctx, license, customerID, err)
	}

	lics, err := h.svc.GetLicenses(ctx, customerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"closeModal": true, "showToast": {"message": "License created successfully", "type": "success"}}`)
	viewLics := h.convertLicenses(ctx, lics)
	return components.LicensesTable(customerID, h.getCustomerName(ctx, customerID), viewLics).Render(ctx, c.Response())
}

func (h *Handler) UpdateLicense(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	licenseCount, _ := strconv.Atoi(c.FormValue("license_count"))
	licenseTerm, _ := strconv.Atoi(c.FormValue("license_term"))
	isSubscription := c.FormValue("license_type") == "subscription"

	req := &admin.UpdateLicenseRequest{
		LicenseCount:        licenseCount,
		IsSubscription:      isSubscription,
		LicenseTerm:         licenseTerm,
		StartDate:           c.FormValue("start_date"),
		ExpirationDate:      c.FormValue("expiration_date"),
		MaintExpirationDate: c.FormValue("maint_expiration_date"),
		MaxProductVersion:   strings.TrimSpace(c.FormValue("max_product_version")),
	}

	// Validate subscription requires term > 0
	if isSubscription && licenseTerm <= 0 {
		license := &vm.License{
			ProductID:           productID,
			LicenseCount:        licenseCount,
			IsSubscription:      isSubscription,
			LicenseTerm:         licenseTerm,
			StartDate:           req.StartDate,
			ExpirationDate:      req.ExpirationDate,
			MaintExpirationDate: req.MaintExpirationDate,
			MaxProductVersion:   req.MaxProductVersion,
		}
		return h.renderLicenseFormWithError(c, ctx, license, customerID, errInvalidSubscriptionTerm)
	}

	// Validate MaxProductVersion format
	if !product.IsValidVersion(req.MaxProductVersion) {
		license := &vm.License{
			ProductID:           productID,
			LicenseCount:        licenseCount,
			IsSubscription:      isSubscription,
			LicenseTerm:         licenseTerm,
			StartDate:           req.StartDate,
			ExpirationDate:      req.ExpirationDate,
			MaintExpirationDate: req.MaintExpirationDate,
			MaxProductVersion:   req.MaxProductVersion,
		}
		return h.renderLicenseFormWithError(c, ctx, license, customerID, errInvalidVersion)
	}

	if err := h.svc.UpdateLicense(ctx, customerID, productID, req); err != nil {
		license := &vm.License{
			ProductID:           productID,
			LicenseCount:        licenseCount,
			IsSubscription:      isSubscription,
			LicenseTerm:         licenseTerm,
			StartDate:           req.StartDate,
			ExpirationDate:      req.ExpirationDate,
			MaintExpirationDate: req.MaintExpirationDate,
			MaxProductVersion:   req.MaxProductVersion,
		}
		return h.renderLicenseFormWithError(c, ctx, license, customerID, err)
	}

	lics, err := h.svc.GetLicenses(ctx, customerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"closeModal": true, "showToast": {"message": "License updated successfully", "type": "success"}}`)
	viewLics := h.convertLicenses(ctx, lics)
	return components.LicensesTable(customerID, h.getCustomerName(ctx, customerID), viewLics).Render(ctx, c.Response())
}

// renderLicenseFormWithError re-renders the license form with appropriate field errors
func (h *Handler) renderLicenseFormWithError(c echo.Context, ctx context.Context, license *vm.License, customerID int64, err error) error {
	errors := make(map[string]string)

	switch {
	case err == errInvalidVersion:
		errors["max_product_version"] = "Must be empty or in #.#.# format (e.g., 1.0.0)"
	case err == errInvalidSubscriptionTerm:
		errors["license_term"] = "Subscription licenses require a term greater than 0"
	case sqlite.IsUniqueConstraintError(err):
		errors["product_id"] = "This customer already has a license for this product"
	default:
		// Unknown error - show toast instead
		setTriggerWithData(c, fmt.Sprintf(`{"showToast": {"message": %q, "type": "error"}}`, "Failed to save license"))
		return c.String(http.StatusUnprocessableEntity, "")
	}

	// Get products for form dropdown
	unlicensedProducts, _ := h.svc.GetUnlicensedProducts(ctx, customerID)
	viewProducts := FromDomainProducts(unlicensedProducts)

	// For edit, include current product in list
	if license.ProductID != 0 {
		prod, _ := h.productSvc.Get(ctx, license.ProductID)
		if prod != nil {
			license.ProductName = prod.ProductName
			viewProducts = append([]vm.Product{FromDomainProduct(*prod)}, viewProducts...)
		}
	}

	formData := components.LicenseFormData{
		License:    license,
		CustomerID: customerID,
		Products:   viewProducts,
		Errors:     errors,
	}
	c.Response().Header().Set("HX-Retarget", "#modal-content")
	c.Response().Header().Set("HX-Reswap", "innerHTML")
	return components.LicenseFormWithErrors(formData).Render(ctx, c.Response())
}

func (h *Handler) DeleteLicense(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	if err := h.svc.DeleteLicense(ctx, customerID, productID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	lics, err := h.svc.GetLicenses(ctx, customerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	setTriggerWithData(c, `{"showToast": {"message": "License deleted successfully", "type": "success"}}`)
	viewLics := h.convertLicenses(ctx, lics)
	return components.LicensesTable(customerID, h.getCustomerName(ctx, customerID), viewLics).Render(ctx, c.Response())
}

// --------------------------
// Product Features (customer values)
// --------------------------

func (h *Handler) GetProductFeatures(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	features := h.getEnrichedFeatureValues(ctx, customerID, productID)
	return components.FeaturesTable(customerID, productID, features).Render(ctx, c.Response())
}

func (h *Handler) EditFeatureValueForm(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}
	featureID, err := strconv.ParseInt(c.Param("featureID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid feature ID")
	}

	features := h.getEnrichedFeatureValues(ctx, customerID, productID)
	for _, f := range features {
		if f.FeatureID == featureID {
			return components.FeatureValueForm(customerID, productID, f).Render(ctx, c.Response())
		}
	}

	return echo.NewHTTPError(http.StatusNotFound, "Feature not found")
}

func (h *Handler) UpdateFeatureValue(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}
	featureID, err := strconv.ParseInt(c.Param("featureID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid feature ID")
	}

	req := &admin.UpdateProductFeatureRequest{
		Value: c.FormValue("feature_value"),
	}

	if err := h.svc.UpdateProductFeature(ctx, customerID, productID, featureID, req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	features := h.getEnrichedFeatureValues(ctx, customerID, productID)
	setTriggerWithData(c, `{"closeModal": true, "showToast": {"message": "Feature value updated successfully", "type": "success"}}`)
	return components.FeaturesTable(customerID, productID, features).Render(ctx, c.Response())
}

// --------------------------
// Machines
// --------------------------

func (h *Handler) GetMachineRegistrations(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	activeOnly := c.QueryParam("active") == "true"
	machines, err := h.svc.GetMachineRegistrations(ctx, customerID, productID, activeOnly)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	prod, err := h.svc.GetProduct(ctx, productID)
	productName := ""
	if err == nil && prod != nil {
		productName = prod.ProductName
	}

	viewMachines := h.convertMachines(ctx, machines, productID)
	return components.MachinesModal(customerID, productID, productName, viewMachines).Render(ctx, c.Response())
}

func (h *Handler) DeleteMachineRegistration(c echo.Context) error {
	ctx := c.Request().Context()
	machineID, err := strconv.ParseInt(c.Param("machineID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid machine ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	// Get machine to find customerID before deleting
	machine, err := h.machineSvc.Get(ctx, machineID)
	if err != nil || machine == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Machine not found")
	}
	customerID := machine.CustomerID

	if err := h.svc.DeleteMachineRegistration(ctx, machineID, productID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Return updated modal content
	prod, err := h.productSvc.Get(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Product not found")
	}

	machines, err := h.machineSvc.GetForLicense(ctx, customerID, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewMachines := h.convertMachines(ctx, machines, productID)
	setTriggerWithData(c, `{"showToast": {"message": "Machine registration deleted successfully", "type": "success"}}`)
	return components.MachinesModal(customerID, productID, prod.ProductName, viewMachines).Render(ctx, c.Response())
}

func (h *Handler) ManualRegistrationForm(c echo.Context) error {
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	prod, _ := h.productSvc.Get(c.Request().Context(), productID)
	productName := ""
	if prod != nil {
		productName = prod.ProductName
	}

	return components.ManualRegistrationForm(customerID, productID, productName, "").
		Render(c.Request().Context(), c.Response())
}

func (h *Handler) CreateManualRegistration(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, err := strconv.ParseInt(c.Param("customerID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid customer ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	req := &activation.Request{
		MachineCode: c.FormValue("machine_code"),
		UserName:    c.FormValue("user_name"),
	}

	_, err = h.activationSvc.Activate(ctx, customerID, productID, req)
	if err != nil {
		// Return to form with error displayed inline
		prod, _ := h.productSvc.Get(ctx, productID)
		productName := ""
		if prod != nil {
			productName = prod.ProductName
		}
		return components.ManualRegistrationForm(customerID, productID, productName, err.Error()).Render(ctx, c.Response())
	}

	// Success: return updated machines modal
	machines, _ := h.svc.GetMachineRegistrations(ctx, customerID, productID, false)
	prod, _ := h.productSvc.Get(ctx, productID)
	productName := ""
	if prod != nil {
		productName = prod.ProductName
	}
	viewMachines := h.convertMachines(ctx, machines, productID)
	setTriggerWithData(c, `{"showToast": {"message": "Registration created successfully", "type": "success"}}`)
	return components.MachinesModal(customerID, productID, productName, viewMachines).Render(ctx, c.Response())
}

func (h *Handler) ExportMachineRegistration(c echo.Context) error {
	ctx := c.Request().Context()
	machineID, err := strconv.ParseInt(c.Param("machineID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid machine ID")
	}
	productID, err := strconv.ParseInt(c.Param("productID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	// Get machine to find customerID and machineCode
	machine, err := h.machineSvc.Get(ctx, machineID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Machine not found")
	}

	// Get product for filename
	prod, err := h.productSvc.Get(ctx, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Product not found")
	}

	// Build activation response using existing data
	req := &activation.Request{
		MachineCode: machine.MachineCode,
		UserName:    machine.UserName,
	}
	resp, err := h.activationSvc.Activate(ctx, machine.CustomerID, productID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Set headers for file download: {product}-{company}-{username}.json
	filename := fmt.Sprintf("%s-%s-%s.json",
		sanitizeFilename(prod.ProductName),
		sanitizeFilename(resp.UserCompany),
		sanitizeFilename(resp.UserName),
	)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	return c.JSONPretty(http.StatusOK, resp, "  ")
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(s string) string {
	// Replace spaces and common problematic characters
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(s)
}

// --------------------------
// Expirations
// --------------------------

func (h *Handler) ListExpirations(c echo.Context) error {
	ctx := c.Request().Context()

	before := c.QueryParam("before")
	if before == "" {
		before = time.Now().Format("2006-01-02")
	}

	expired, err := h.svc.GetExpiredLicenses(ctx, before)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewExpired := FromDomainExpiredLicenses(expired)
	if isHTMX(c) {
		return components.ExpirationsTable(viewExpired, before).Render(ctx, c.Response())
	}
	return pages.Expirations(viewExpired, before).Render(ctx, c.Response())
}

func (h *Handler) ExportExpirationsCSV(c echo.Context) error {
	ctx := c.Request().Context()

	before := c.QueryParam("before")
	if before == "" {
		before = time.Now().Format("2006-01-02")
	}

	expired, err := h.svc.GetExpiredLicenses(ctx, before)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Build CSV content
	var builder strings.Builder
	builder.WriteString("Customer Name,Contact Name,Email,Product Name,Expiration Date,Maint. End\n")
	for _, e := range expired {
		// Escape fields that may contain commas or quotes
		builder.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
			csvEscape(e.CustomerName),
			csvEscape(e.ContactName),
			csvEscape(e.Email),
			csvEscape(e.ProductName),
			e.ExpirationDate,
			e.MaintExpirationDate,
		))
	}

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=expired_licenses_%s.csv", before))
	return c.String(http.StatusOK, builder.String())
}

// csvEscape escapes a string for CSV output
func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

// --------------------------
// Helper methods
// --------------------------

func isHTMX(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func setTriggerWithData(c echo.Context, eventJSON string) {
	c.Response().Header().Set("HX-Trigger", eventJSON)
}

// Index renders the dashboard page
func (h *Handler) Index(c echo.Context) error {
	ctx := c.Request().Context()
	customers, err := h.svc.GetCustomers(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// Check for pre-selected customer via query param
	selectedCustomerID := c.QueryParam("customer")
	return pages.Index(FromDomainCustomers(customers), selectedCustomerID).Render(ctx, c.Response())
}

func (h *Handler) convertLicenses(ctx context.Context, lics []license.License) []License {
	result := make([]License, len(lics))
	for i, lic := range lics {
		prod, _ := h.svc.GetProduct(ctx, lic.ProductID)
		productName := ""
		if prod != nil {
			productName = prod.ProductName
		}
		result[i] = FromDomainLicense(lic, productName)
	}
	return result
}

func (h *Handler) getEnrichedFeatureValues(ctx context.Context, customerID, productID int64) []ProductFeature {
	// Get feature definitions
	features, _ := h.featureSvc.GetForProduct(ctx, productID)
	// Get feature values
	values, _ := h.featureValSvc.GetFeatureValues(ctx, customerID, productID)

	// Create a map of feature values by feature ID
	valueMap := make(map[int64]featurevalue.FeatureValue)
	for _, v := range values {
		valueMap[v.FeatureID] = v
	}

	// Combine features with values
	result := make([]ProductFeature, len(features))
	for i, f := range features {
		fv := featurevalue.FeatureValue{
			CustomerID: customerID,
			ProductID:  productID,
			FeatureID:  f.FeatureID,
		}
		if v, ok := valueMap[f.FeatureID]; ok {
			fv.FeatureValue = v.FeatureValue
		}
		result[i] = FromDomainFeatureValue(fv, f)
	}
	return result
}

func (h *Handler) convertMachines(ctx context.Context, machines []machine.Machine, productID int64) []MachineRegistration {
	result := make([]MachineRegistration, len(machines))
	for i, m := range machines {
		// Get registration details for this machine/product
		regs, _ := h.regSvc.GetForMachine(ctx, m.MachineID)
		var regHash, expDate, firstRegDate, lastRegDate, installedVersion string
		for _, r := range regs {
			if r.ProductID == productID {
				regHash = r.RegistrationHash
				expDate = r.ExpirationDate
				firstRegDate = r.FirstRegistrationDate
				lastRegDate = r.LastRegistrationDate
				installedVersion = r.InstalledVersion
				break
			}
		}
		result[i] = FromDomainMachine(m, productID, regHash, expDate, firstRegDate, lastRegDate, installedVersion)
	}
	return result
}
