package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/activation"
	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/http/client"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/middleware"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/registration"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestGetProductVersion(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	// Create services
	productSvc := product.NewService(db)
	regSvc := registration.NewService(db)
	licenseSvc := license.NewService(db)
	machineSvc := machine.NewService(db)
	featureSvc := feature.NewService(db)
	featureValueSvc := featurevalue.NewService(db)
	customerSvc := customer.NewService(db)
	activationSvc := activation.NewService(
		db,
		"test-secret",
		customerSvc,
		machineSvc,
		regSvc,
		licenseSvc,
		productSvc,
		featureSvc,
		featureValueSvc,
	)

	handler := client.NewHandler(activationSvc, regSvc, productSvc, licenseSvc, machineSvc, featureSvc, featureValueSvc, customerSvc)

	// Create a test product
	testProduct := &product.Product{
		ProductName:   "Test App",
		ProductGUID:   "TEST-GUID-123",
		LatestVersion: "2.5.0",
		DownloadURL:   "https://example.com/download/test-app",
	}
	created, err := productSvc.Create(ctx, testProduct)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	t.Run("returns product version for valid GUID", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/productver/"+created.ProductGUID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("guid")
		c.SetParamValues(created.ProductGUID)

		err := handler.GetProductVersion(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if resp["ProductGUID"] != created.ProductGUID {
			t.Errorf("expected ProductGUID %q, got %q", created.ProductGUID, resp["ProductGUID"])
		}
		if resp["LatestVersion"] != "2.5.0" {
			t.Errorf("expected LatestVersion %q, got %q", "2.5.0", resp["LatestVersion"])
		}
		if resp["DownloadURL"] != "https://example.com/download/test-app" {
			t.Errorf("expected DownloadURL %q, got %q", "https://example.com/download/test-app", resp["DownloadURL"])
		}
	})

	t.Run("returns 404 for unknown GUID", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/productver/UNKNOWN-GUID", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("guid")
		c.SetParamValues("UNKNOWN-GUID")

		err := handler.GetProductVersion(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp["error"] != "product not found" {
			t.Errorf("expected error %q, got %q", "product not found", resp["error"])
		}
	})

	t.Run("returns 400 for missing GUID", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/productver/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("guid")
		c.SetParamValues("")

		err := handler.GetProductVersion(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp["error"] != "missing product guid" {
			t.Errorf("expected error %q, got %q", "missing product guid", resp["error"])
		}
	})
}

func TestActivate(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	// Create services
	customerSvc := customer.NewService(db)
	productSvc := product.NewService(db)
	machineSvc := machine.NewService(db)
	regSvc := registration.NewService(db)
	licenseSvc := license.NewService(db)
	featureSvc := feature.NewService(db)
	featureValueSvc := featurevalue.NewService(db)

	activationSvc := activation.NewService(
		db,
		"test-secret",
		customerSvc,
		machineSvc,
		regSvc,
		licenseSvc,
		productSvc,
		featureSvc,
		featureValueSvc,
	)

	handler := client.NewHandler(activationSvc, regSvc, productSvc, licenseSvc, machineSvc, featureSvc, featureValueSvc, customerSvc)

	// Setup test data
	testCustomer := &customer.Customer{
		CustomerName: "Test Company",
		ContactName:  "John Doe",
		Email:        "john@test.com",
	}
	createdCustomer, err := customerSvc.Create(ctx, testCustomer)
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	testProduct := &product.Product{
		ProductName:   "Test App",
		ProductGUID:   "PROD-GUID-456",
		LatestVersion: "3.0.0",
		DownloadURL:   "https://example.com/download",
	}
	createdProduct, err := productSvc.Create(ctx, testProduct)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	// Create license
	lic := &license.License{
		CustomerID:          createdCustomer.CustomerID,
		ProductID:           createdProduct.ProductID,
		LicenseKey:          "REG-GUID-789",
		LicenseCount:        5,
		IsSubscription:      false,
		LicenseTerm:         12,
		StartDate:           "2024-01-01",
		ExpirationDate:      "2025-12-31",
		MaintExpirationDate: "2025-12-31",
		MaxProductVersion:   "99.0.0",
	}
	if _, err := licenseSvc.Create(ctx, lic); err != nil {
		t.Fatalf("create license: %v", err)
	}

	t.Run("successful activation", func(t *testing.T) {
		e := echo.New()

		reqBody := activation.Request{
			MachineCode: "MACHINE-001",
			UserName:    "testuser",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/activate", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Set the license context (normally set by middleware)
		c.Set("license", middleware.LicenseContext{
			CustomerID: createdCustomer.CustomerID,
			ProductID:  createdProduct.ProductID,
		})

		err := handler.Activate(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var resp activation.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if resp.UserName != "testuser" {
			t.Errorf("expected UserName %q, got %q", "testuser", resp.UserName)
		}
		if resp.UserCompany != "Test Company" {
			t.Errorf("expected UserCompany %q, got %q", "Test Company", resp.UserCompany)
		}
		if resp.MachineCode != "MACHINE-001" {
			t.Errorf("expected MachineCode %q, got %q", "MACHINE-001", resp.MachineCode)
		}
		if resp.ProductGUID != "prod-guid-456" {
			t.Errorf("expected ProductGUID %q, got %q", "prod-guid-456", resp.ProductGUID)
		}
		if resp.ExpirationDate != "2025-12-31" {
			t.Errorf("expected ExpirationDate %q, got %q", "2025-12-31", resp.ExpirationDate)
		}
	})

	t.Run("returns 401 without custProd context", func(t *testing.T) {
		e := echo.New()

		reqBody := activation.Request{
			MachineCode: "MACHINE-002",
			UserName:    "testuser2",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/activate", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Not setting license context

		err := handler.Activate(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp["error"] != "invalid license context" {
			t.Errorf("expected error %q, got %q", "invalid license context", resp["error"])
		}
	})

	t.Run("returns 400 for invalid request body", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/activate", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Activate(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp["error"] != "invalid request body" {
			t.Errorf("expected error %q, got %q", "invalid request body", resp["error"])
		}
	})
}

func TestRegisterRoutes(t *testing.T) {
	db := testutil.NewTestDB(t)

	productSvc := product.NewService(db)
	regSvc := registration.NewService(db)
	licenseSvc := license.NewService(db)
	machineSvc := machine.NewService(db)
	featureSvc := feature.NewService(db)
	featureValueSvc := featurevalue.NewService(db)
	customerSvc := customer.NewService(db)
	activationSvc := activation.NewService(
		db,
		"test-secret",
		customerSvc,
		machineSvc,
		regSvc,
		licenseSvc,
		productSvc,
		featureSvc,
		featureValueSvc,
	)

	handler := client.NewHandler(activationSvc, regSvc, productSvc, licenseSvc, machineSvc, featureSvc, featureValueSvc, customerSvc)

	e := echo.New()
	g := e.Group("/api/v1")
	// Pass a no-op middleware for testing
	noopMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
	client.RegisterRoutes(g, handler, noopMiddleware)

	// Verify routes are registered
	routes := e.Routes()
	expectedRoutes := map[string]string{
		"POST:/api/v1/activate":            "activate",
		"GET:/api/v1/productver/:guid":     "productver",
		"GET:/api/v1/license/:license_key": "license",
		"PUT:/api/v1/license/:license_key": "license update",
	}

	found := make(map[string]bool)
	for _, r := range routes {
		key := r.Method + ":" + r.Path
		if _, ok := expectedRoutes[key]; ok {
			found[key] = true
		}
	}

	for key := range expectedRoutes {
		if !found[key] {
			t.Errorf("expected route %s to be registered", key)
		}
	}
}

func TestGetLicenseInfo(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	// Create services
	customerSvc := customer.NewService(db)
	productSvc := product.NewService(db)
	machineSvc := machine.NewService(db)
	regSvc := registration.NewService(db)
	licenseSvc := license.NewService(db)
	featureSvc := feature.NewService(db)
	featureValueSvc := featurevalue.NewService(db)

	activationSvc := activation.NewService(
		db,
		"test-secret",
		customerSvc,
		machineSvc,
		regSvc,
		licenseSvc,
		productSvc,
		featureSvc,
		featureValueSvc,
	)

	handler := client.NewHandler(activationSvc, regSvc, productSvc, licenseSvc, machineSvc, featureSvc, featureValueSvc, customerSvc)

	// Setup test data
	testCustomer := &customer.Customer{
		CustomerName: "Test Company",
		ContactName:  "John Doe",
		Email:        "john@test.com",
	}
	createdCustomer, err := customerSvc.Create(ctx, testCustomer)
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	testProduct := &product.Product{
		ProductName:   "Test App",
		ProductGUID:   "PROD-GUID-INFO",
		LatestVersion: "3.0.0",
		DownloadURL:   "https://example.com/download",
	}
	createdProduct, err := productSvc.Create(ctx, testProduct)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	// Create license with future expiration dates
	lic := &license.License{
		CustomerID:          createdCustomer.CustomerID,
		ProductID:           createdProduct.ProductID,
		LicenseKey:          "LICENSE-KEY-123",
		LicenseCount:        5,
		IsSubscription:      false,
		LicenseTerm:         12,
		StartDate:           "2024-01-01",
		ExpirationDate:      "2099-12-31",
		MaintExpirationDate: "2099-12-31",
		MaxProductVersion:   "4.0.0",
	}
	if _, err := licenseSvc.Create(ctx, lic); err != nil {
		t.Fatalf("create license: %v", err)
	}

	t.Run("returns license info for valid license key", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/license/LICENSE-KEY-123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("license_key")
		c.SetParamValues("LICENSE-KEY-123")

		err := handler.GetLicenseInfo(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var resp client.LicenseInfoResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if resp.CustomerName != "Test Company" {
			t.Errorf("expected CustomerName %q, got %q", "Test Company", resp.CustomerName)
		}
		// ProductGUID is normalized to lowercase
		if resp.ProductGUID != "prod-guid-info" {
			t.Errorf("expected ProductGUID %q, got %q", "prod-guid-info", resp.ProductGUID)
		}
		if resp.ProductName != "Test App" {
			t.Errorf("expected ProductName %q, got %q", "Test App", resp.ProductName)
		}
		if resp.LicenseCount != 5 {
			t.Errorf("expected LicenseCount %d, got %d", 5, resp.LicenseCount)
		}
		if resp.LicensesAvailable != 5 {
			t.Errorf("expected LicensesAvailable %d, got %d (no machines registered)", 5, resp.LicensesAvailable)
		}
		if resp.ExpirationDate != "2099-12-31" {
			t.Errorf("expected ExpirationDate %q, got %q", "2099-12-31", resp.ExpirationDate)
		}
		if resp.MaxProductVersion != "4.0.0" {
			t.Errorf("expected MaxProductVersion %q, got %q", "4.0.0", resp.MaxProductVersion)
		}
		if resp.LatestVersion != "3.0.0" {
			t.Errorf("expected LatestVersion %q, got %q", "3.0.0", resp.LatestVersion)
		}
	})

	t.Run("returns 404 for unknown license key", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/license/UNKNOWN-KEY", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("license_key")
		c.SetParamValues("UNKNOWN-KEY")

		err := handler.GetLicenseInfo(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("returns correct available count with active registrations", func(t *testing.T) {
		// Activate a machine to reduce available count
		activationSvc.Activate(ctx, createdCustomer.CustomerID, createdProduct.ProductID, &activation.Request{
			MachineCode: "MACHINE-INFO-001",
			UserName:    "testuser",
		})

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/license/LICENSE-KEY-123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("license_key")
		c.SetParamValues("LICENSE-KEY-123")

		err := handler.GetLicenseInfo(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		var resp client.LicenseInfoResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		// One machine activated, so available should be 4
		if resp.LicensesAvailable != 4 {
			t.Errorf("expected LicensesAvailable %d after 1 activation, got %d", 4, resp.LicensesAvailable)
		}
	})
}

func TestUpdateLicenseInfo(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	// Create services
	customerSvc := customer.NewService(db)
	productSvc := product.NewService(db)
	machineSvc := machine.NewService(db)
	regSvc := registration.NewService(db)
	licenseSvc := license.NewService(db)
	featureSvc := feature.NewService(db)
	featureValueSvc := featurevalue.NewService(db)

	activationSvc := activation.NewService(
		db,
		"test-secret",
		customerSvc,
		machineSvc,
		regSvc,
		licenseSvc,
		productSvc,
		featureSvc,
		featureValueSvc,
	)

	handler := client.NewHandler(activationSvc, regSvc, productSvc, licenseSvc, machineSvc, featureSvc, featureValueSvc, customerSvc)

	// Setup test data
	testCustomer := &customer.Customer{
		CustomerName: "Update Test Company",
		ContactName:  "Jane Doe",
		Email:        "jane@test.com",
	}
	createdCustomer, err := customerSvc.Create(ctx, testCustomer)
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	testProduct := &product.Product{
		ProductName:   "Update Test App",
		ProductGUID:   "PROD-GUID-UPDATE",
		LatestVersion: "4.0.0",
		DownloadURL:   "https://example.com/download",
	}
	createdProduct, err := productSvc.Create(ctx, testProduct)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	// Create license
	lic := &license.License{
		CustomerID:          createdCustomer.CustomerID,
		ProductID:           createdProduct.ProductID,
		LicenseKey:          "UPDATE-LICENSE-KEY",
		LicenseCount:        5,
		IsSubscription:      false,
		LicenseTerm:         12,
		StartDate:           "2024-01-01",
		ExpirationDate:      "2099-12-31",
		MaintExpirationDate: "2099-12-31",
		MaxProductVersion:   "5.0.0",
	}
	if _, err := licenseSvc.Create(ctx, lic); err != nil {
		t.Fatalf("create license: %v", err)
	}

	// Activate a machine first
	_, err = activationSvc.Activate(ctx, createdCustomer.CustomerID, createdProduct.ProductID, &activation.Request{
		MachineCode: "UPDATE-MACHINE-001",
		UserName:    "updateuser",
	})
	if err != nil {
		t.Fatalf("activate machine: %v", err)
	}

	t.Run("successfully updates installed version", func(t *testing.T) {
		e := echo.New()
		reqBody := client.UpdateLicenseRequest{
			MachineCode:      "UPDATE-MACHINE-001",
			InstalledVersion: "3.5.0",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/license/UPDATE-LICENSE-KEY", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("license_key")
		c.SetParamValues("UPDATE-LICENSE-KEY")

		err := handler.UpdateLicenseInfo(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var resp client.LicenseInfoResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if resp.CustomerName != "Update Test Company" {
			t.Errorf("expected CustomerName %q, got %q", "Update Test Company", resp.CustomerName)
		}
		if resp.ProductName != "Update Test App" {
			t.Errorf("expected ProductName %q, got %q", "Update Test App", resp.ProductName)
		}
		if resp.LatestVersion != "4.0.0" {
			t.Errorf("expected LatestVersion %q, got %q", "4.0.0", resp.LatestVersion)
		}
	})

	t.Run("returns 404 for unknown license key", func(t *testing.T) {
		e := echo.New()
		reqBody := client.UpdateLicenseRequest{
			MachineCode:      "UPDATE-MACHINE-001",
			InstalledVersion: "3.5.0",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/license/UNKNOWN-KEY", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("license_key")
		c.SetParamValues("UNKNOWN-KEY")

		err := handler.UpdateLicenseInfo(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("returns 404 for unknown machine code", func(t *testing.T) {
		e := echo.New()
		reqBody := client.UpdateLicenseRequest{
			MachineCode:      "UNKNOWN-MACHINE",
			InstalledVersion: "3.5.0",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/license/UPDATE-LICENSE-KEY", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("license_key")
		c.SetParamValues("UPDATE-LICENSE-KEY")

		err := handler.UpdateLicenseInfo(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp["error"] != "machine not found for this license" {
			t.Errorf("expected error %q, got %q", "machine not found for this license", resp["error"])
		}
	})

	t.Run("returns 400 for missing machineCode", func(t *testing.T) {
		e := echo.New()
		reqBody := client.UpdateLicenseRequest{
			InstalledVersion: "3.5.0",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/license/UPDATE-LICENSE-KEY", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("license_key")
		c.SetParamValues("UPDATE-LICENSE-KEY")

		err := handler.UpdateLicenseInfo(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp["error"] != "machineCode is required" {
			t.Errorf("expected error %q, got %q", "machineCode is required", resp["error"])
		}
	})

	t.Run("returns 400 for invalid request body", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/license/UPDATE-LICENSE-KEY", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("license_key")
		c.SetParamValues("UPDATE-LICENSE-KEY")

		err := handler.UpdateLicenseInfo(c)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp["error"] != "invalid request body" {
			t.Errorf("expected error %q, got %q", "invalid request body", resp["error"])
		}
	})
}
