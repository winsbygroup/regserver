package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"

	"winsbygroup.com/regserver/internal/middleware"
	"winsbygroup.com/regserver/internal/testutil"
)

// Helper to create echo context with request/response
func newContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// Dummy handler that returns 200 OK
func okHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// ============================================================================
// AdminAPIKeyAuth Tests
// ============================================================================

func TestAdminAPIKeyAuth(t *testing.T) {
	const testAPIKey = "test-admin-key-12345"

	t.Run("allows request with valid API key", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		c, rec := newContext(http.MethodGet, "/api/admin/test")
		c.Request().Header.Set("X-API-Key", testAPIKey)

		mw := middleware.AdminAPIKeyAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("rejects request with invalid API key", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		c, _ := newContext(http.MethodGet, "/api/admin/test")
		c.Request().Header.Set("X-API-Key", "wrong-key")

		mw := middleware.AdminAPIKeyAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Fatalf("expected echo.HTTPError, got %T", err)
		}
		if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", httpErr.Code)
		}
	})

	t.Run("rejects request with missing API key", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		c, _ := newContext(http.MethodGet, "/api/admin/test")
		// No X-API-Key header

		mw := middleware.AdminAPIKeyAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Fatalf("expected echo.HTTPError, got %T", err)
		}
		if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", httpErr.Code)
		}
	})

	t.Run("rejects when ADMIN_API_KEY env var not set", func(t *testing.T) {
		os.Unsetenv("ADMIN_API_KEY")

		c, _ := newContext(http.MethodGet, "/api/admin/test")
		c.Request().Header.Set("X-API-Key", "any-key")

		mw := middleware.AdminAPIKeyAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Fatalf("expected echo.HTTPError, got %T", err)
		}
		if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", httpErr.Code)
		}
	})
}

// ============================================================================
// WebAuth Tests
// ============================================================================

func TestWebAuth(t *testing.T) {
	const testAPIKey = "test-admin-key-12345"

	t.Run("allows login page without auth", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/login", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/web/login")

		mw := middleware.WebAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("allows request with valid X-API-Key header", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/customers", nil)
		req.Header.Set("X-API-Key", testAPIKey)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/web/customers")

		mw := middleware.WebAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("allows request with valid session cookie", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		// Create a session and use its ID as the cookie value
		sessionID := middleware.CreateSession()
		defer middleware.DeleteSession(sessionID)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/customers", nil)
		req.AddCookie(&http.Cookie{
			Name:  middleware.SessionCookieName,
			Value: sessionID,
		})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/web/customers")

		mw := middleware.WebAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("redirects to login without auth", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/customers", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/web/customers")

		mw := middleware.WebAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error (redirect), got %v", err)
		}
		if rec.Code != http.StatusFound {
			t.Errorf("expected status 302, got %d", rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/web/login" {
			t.Errorf("expected redirect to /web/login, got %s", loc)
		}
	})

	t.Run("redirects with invalid cookie", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/customers", nil)
		req.AddCookie(&http.Cookie{
			Name:  middleware.SessionCookieName,
			Value: "wrong-value",
		})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/web/customers")

		mw := middleware.WebAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error (redirect), got %v", err)
		}
		if rec.Code != http.StatusFound {
			t.Errorf("expected status 302, got %d", rec.Code)
		}
	})

	t.Run("redirects to login when ADMIN_API_KEY not set and no session", func(t *testing.T) {
		os.Unsetenv("ADMIN_API_KEY")

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/customers", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/web/customers")

		mw := middleware.WebAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error (redirect), got %v", err)
		}
		// Should redirect to login since no valid auth
		if rec.Code != http.StatusFound {
			t.Errorf("expected status 302, got %d", rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/web/login" {
			t.Errorf("expected redirect to /web/login, got %s", loc)
		}
	})
}

// ============================================================================
// ValidateAdminKey Tests
// ============================================================================

func TestValidateAdminKey(t *testing.T) {
	const testAPIKey = "test-admin-key-12345"

	t.Run("returns true for valid key", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		if !middleware.ValidateAdminKey(testAPIKey) {
			t.Error("expected true for valid key")
		}
	})

	t.Run("returns false for invalid key", func(t *testing.T) {
		os.Setenv("ADMIN_API_KEY", testAPIKey)
		defer os.Unsetenv("ADMIN_API_KEY")

		if middleware.ValidateAdminKey("wrong-key") {
			t.Error("expected false for invalid key")
		}
	})

	t.Run("returns false when env var not set", func(t *testing.T) {
		os.Unsetenv("ADMIN_API_KEY")

		if middleware.ValidateAdminKey(testAPIKey) {
			t.Error("expected false when env var not set")
		}
	})
}

// ============================================================================
// LicenseKeyAuth Tests
// ============================================================================

func TestLicenseKeyAuth(t *testing.T) {
	db := testutil.NewTestDB(t)

	// Seed test data
	_, err := db.Exec(`INSERT INTO customer (customer_id, customer_name) VALUES (1, 'Test Customer')`)
	if err != nil {
		t.Fatalf("failed to insert customer: %v", err)
	}
	_, err = db.Exec(`INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url)
		VALUES (1, 'Test Product', 'test-guid', '1.0.0', 'http://example.com')`)
	if err != nil {
		t.Fatalf("failed to insert product: %v", err)
	}
	_, err = db.Exec(`INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term)
		VALUES (1, 1, 'valid-reg-guid-123', 5, 0, 12)`)
	if err != nil {
		t.Fatalf("failed to insert license: %v", err)
	}

	t.Run("allows request with valid license key", func(t *testing.T) {
		c, rec := newContext(http.MethodPost, "/api/v1/activate")
		c.Request().Header.Set("X-License-Key", "valid-reg-guid-123")

		mw := middleware.LicenseKeyAuth(db)
		handler := mw(func(c echo.Context) error {
			// Verify license context was set
			lic, ok := c.Get("license").(middleware.LicenseContext)
			if !ok {
				t.Error("license context not set")
			}
			if lic.CustomerID != 1 || lic.ProductID != 1 {
				t.Errorf("unexpected license context: %+v", lic)
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("rejects request with invalid license key", func(t *testing.T) {
		c, _ := newContext(http.MethodPost, "/api/v1/activate")
		c.Request().Header.Set("X-License-Key", "invalid-key")

		mw := middleware.LicenseKeyAuth(db)
		handler := mw(okHandler)

		err := handler(c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Fatalf("expected echo.HTTPError, got %T", err)
		}
		if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", httpErr.Code)
		}
	})

	t.Run("rejects request with missing license key", func(t *testing.T) {
		c, _ := newContext(http.MethodPost, "/api/v1/activate")
		// No X-License-Key header

		mw := middleware.LicenseKeyAuth(db)
		handler := mw(okHandler)

		err := handler(c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Fatalf("expected echo.HTTPError, got %T", err)
		}
		if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", httpErr.Code)
		}
	})
}

// ============================================================================
// Theme Middleware Tests
// ============================================================================

func TestTheme(t *testing.T) {
	t.Run("defaults to light theme", func(t *testing.T) {
		c, rec := newContext(http.MethodGet, "/web/")

		mw := middleware.Theme()
		handler := mw(func(c echo.Context) error {
			theme := middleware.GetTheme(c.Request().Context())
			if theme != "light" {
				t.Errorf("expected 'light', got %q", theme)
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("reads dark theme from cookie", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/", nil)
		req.AddCookie(&http.Cookie{
			Name:  middleware.ThemeCookieName,
			Value: "dark",
		})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := middleware.Theme()
		handler := mw(func(c echo.Context) error {
			theme := middleware.GetTheme(c.Request().Context())
			if theme != "dark" {
				t.Errorf("expected 'dark', got %q", theme)
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("ignores invalid theme cookie", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/", nil)
		req.AddCookie(&http.Cookie{
			Name:  middleware.ThemeCookieName,
			Value: "invalid-theme",
		})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := middleware.Theme()
		handler := mw(func(c echo.Context) error {
			theme := middleware.GetTheme(c.Request().Context())
			if theme != "light" {
				t.Errorf("expected 'light' for invalid cookie, got %q", theme)
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestGetTheme(t *testing.T) {
	t.Run("returns light for empty context", func(t *testing.T) {
		ctx := context.Background()
		theme := middleware.GetTheme(ctx)
		if theme != "light" {
			t.Errorf("expected 'light', got %q", theme)
		}
	})
}

// ============================================================================
// Version Middleware Tests
// ============================================================================

func TestVersion(t *testing.T) {
	t.Run("adds version to context", func(t *testing.T) {
		c, rec := newContext(http.MethodGet, "/web/")

		mw := middleware.Version()
		handler := mw(func(c echo.Context) error {
			version := middleware.GetVersion(c.Request().Context())
			if version == "" {
				t.Error("expected version to be set")
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestGetVersion(t *testing.T) {
	t.Run("returns version for empty context", func(t *testing.T) {
		ctx := context.Background()
		version := middleware.GetVersion(ctx)
		if version == "" {
			t.Error("expected non-empty version")
		}
	})
}

// ============================================================================
// Session Tests
// ============================================================================

func TestCreateSession(t *testing.T) {
	t.Run("creates session with valid ID", func(t *testing.T) {
		sessionID := middleware.CreateSession()
		defer middleware.DeleteSession(sessionID)

		if sessionID == "" {
			t.Error("expected non-empty session ID")
		}

		// UUID format: 8-4-4-4-12 = 36 chars
		if len(sessionID) != 36 {
			t.Errorf("expected UUID format (36 chars), got %d chars", len(sessionID))
		}
	})

	t.Run("creates unique sessions", func(t *testing.T) {
		id1 := middleware.CreateSession()
		id2 := middleware.CreateSession()
		defer middleware.DeleteSession(id1)
		defer middleware.DeleteSession(id2)

		if id1 == id2 {
			t.Error("expected unique session IDs")
		}
	})
}

func TestGetSession(t *testing.T) {
	t.Run("retrieves valid session", func(t *testing.T) {
		sessionID := middleware.CreateSession()
		defer middleware.DeleteSession(sessionID)

		session, ok := middleware.GetSession(sessionID)
		if !ok {
			t.Fatal("expected session to exist")
		}
		if session.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
		if session.ExpiresAt.IsZero() {
			t.Error("expected ExpiresAt to be set")
		}
		if !session.ExpiresAt.After(session.CreatedAt) {
			t.Error("expected ExpiresAt to be after CreatedAt")
		}
	})

	t.Run("returns false for non-existent session", func(t *testing.T) {
		_, ok := middleware.GetSession("non-existent-session-id")
		if ok {
			t.Error("expected session to not exist")
		}
	})

	t.Run("returns false for empty session ID", func(t *testing.T) {
		_, ok := middleware.GetSession("")
		if ok {
			t.Error("expected empty session ID to not exist")
		}
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("deletes existing session", func(t *testing.T) {
		sessionID := middleware.CreateSession()

		// Verify it exists
		_, ok := middleware.GetSession(sessionID)
		if !ok {
			t.Fatal("expected session to exist before delete")
		}

		// Delete it
		middleware.DeleteSession(sessionID)

		// Verify it's gone
		_, ok = middleware.GetSession(sessionID)
		if ok {
			t.Error("expected session to not exist after delete")
		}
	})

	t.Run("handles non-existent session gracefully", func(t *testing.T) {
		// Should not panic
		middleware.DeleteSession("non-existent-session-id")
	})
}

// ============================================================================
// CSRF Middleware Tests
// ============================================================================

func TestCSRF(t *testing.T) {
	t.Run("copies token from echo context to request context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Simulate Echo's CSRF middleware setting the token
		c.Set("csrf", "test-csrf-token-12345")

		mw := middleware.CSRF()
		handler := mw(func(c echo.Context) error {
			token := middleware.GetCSRF(c.Request().Context())
			if token != "test-csrf-token-12345" {
				t.Errorf("expected 'test-csrf-token-12345', got %q", token)
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("handles missing csrf token gracefully", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// No csrf token set in echo context

		mw := middleware.CSRF()
		handler := mw(func(c echo.Context) error {
			token := middleware.GetCSRF(c.Request().Context())
			if token != "" {
				t.Errorf("expected empty token, got %q", token)
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestGetCSRF(t *testing.T) {
	t.Run("returns empty string for empty context", func(t *testing.T) {
		ctx := context.Background()
		token := middleware.GetCSRF(ctx)
		if token != "" {
			t.Errorf("expected empty string, got %q", token)
		}
	})
}

// ============================================================================
// WebAuth Static Assets Test
// ============================================================================

func TestWebAuthStaticAssets(t *testing.T) {
	t.Run("allows static assets without auth", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/web/static/js/app.js", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/web/static/js/app.js")

		mw := middleware.WebAuth()
		handler := mw(okHandler)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}
