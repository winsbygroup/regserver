package license_test

import (
	"context"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestLicenseLifecycle(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licSvc := license.NewService(db)

	// Create customer
	c, _ := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Acme",
		ContactName:  "John Doe",
		Email:        "john@acme.com",
	})

	// Create two products
	p1, _ := prodSvc.Create(ctx, &product.Product{
		ProductName:   "Widget",
		ProductGUID:   "GUID-1",
		LatestVersion: "1.0.0",
		DownloadURL:   "url1",
	})

	p2, _ := prodSvc.Create(ctx, &product.Product{
		ProductName:   "Gadget",
		ProductGUID:   "GUID-2",
		LatestVersion: "1.0.0",
		DownloadURL:   "url2",
	})

	// Initially both are unlicensed
	unlicensed, _ := licSvc.GetUnlicensed(ctx, c.CustomerID)
	if len(unlicensed) != 2 {
		t.Fatalf("expected 2 unlicensed, got %d", len(unlicensed))
	}

	// License product 1
	lic := &license.License{
		CustomerID:          c.CustomerID,
		ProductID:           p1.ProductID,
		LicenseCount:        5,
		IsSubscription:      false,
		LicenseTerm:         0,
		LicenseKey:          "LIC-123",
		StartDate:           "2024-01-01",
		ExpirationDate:      "9999-12-31",
		MaintExpirationDate: "9999-12-31",
	}

	if _, err := licSvc.Create(ctx, lic); err != nil {
		t.Fatalf("create license: %v", err)
	}

	// Now only product 2 is unlicensed
	unlicensed, _ = licSvc.GetUnlicensed(ctx, c.CustomerID)
	if len(unlicensed) != 1 || unlicensed[0].ProductID != p2.ProductID {
		t.Fatalf("expected only product 2 unlicensed")
	}

	// Delete license
	if err := licSvc.Delete(ctx, c.CustomerID, p1.ProductID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Both unlicensed again
	unlicensed, _ = licSvc.GetUnlicensed(ctx, c.CustomerID)
	if len(unlicensed) != 2 {
		t.Fatalf("expected 2 unlicensed after delete")
	}
}

func TestGetExpiredLicenses(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licSvc := license.NewService(db)

	// Create customers
	c1, _ := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Acme Corp",
		ContactName:  "John Doe",
		Email:        "john@acme.com",
	})
	c2, _ := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Beta Inc",
		ContactName:  "Jane Smith",
		Email:        "jane@beta.com",
	})

	// Create products
	p1, _ := prodSvc.Create(ctx, &product.Product{
		ProductName: "Widget Pro",
		ProductGUID: "GUID-1",
	})
	p2, _ := prodSvc.Create(ctx, &product.Product{
		ProductName: "Gadget Plus",
		ProductGUID: "GUID-2",
	})

	// Create licenses with different expiration dates
	// Expired license (past date)
	licSvc.Create(ctx, &license.License{
		CustomerID:          c1.CustomerID,
		ProductID:           p1.ProductID,
		LicenseKey:          "LIC-1",
		LicenseCount:        1,
		StartDate:           "2019-01-01",
		ExpirationDate:      "2020-01-15",
		MaintExpirationDate: "2020-01-15",
	})

	// Another expired license
	licSvc.Create(ctx, &license.License{
		CustomerID:          c2.CustomerID,
		ProductID:           p2.ProductID,
		LicenseKey:          "LIC-2",
		LicenseCount:        1,
		StartDate:           "2020-01-01",
		ExpirationDate:      "2021-06-30",
		MaintExpirationDate: "2021-06-30",
	})

	// Active license (future date)
	licSvc.Create(ctx, &license.License{
		CustomerID:          c1.CustomerID,
		ProductID:           p2.ProductID,
		LicenseKey:          "LIC-3",
		LicenseCount:        1,
		StartDate:           "2024-01-01",
		ExpirationDate:      "2099-12-31",
		MaintExpirationDate: "2099-12-31",
	})

	// Test with a date that captures both expired licenses
	expired, err := licSvc.GetExpiredLicenses(ctx, "2025-01-01")
	if err != nil {
		t.Fatalf("GetExpiredLicenses: %v", err)
	}

	if len(expired) != 2 {
		t.Fatalf("expected 2 expired licenses, got %d", len(expired))
	}

	// Results should be sorted by expiration_date DESC (most recent first)
	if expired[0].ExpirationDate != "2021-06-30" {
		t.Errorf("expected first result to be 2021-06-30, got %s", expired[0].ExpirationDate)
	}
	if expired[1].ExpirationDate != "2020-01-15" {
		t.Errorf("expected second result to be 2020-01-15, got %s", expired[1].ExpirationDate)
	}

	// Verify customer/product details are included
	if expired[0].CustomerName != "Beta Inc" {
		t.Errorf("expected CustomerName 'Beta Inc', got %s", expired[0].CustomerName)
	}
	if expired[0].ProductName != "Gadget Plus" {
		t.Errorf("expected ProductName 'Gadget Plus', got %s", expired[0].ProductName)
	}
	if expired[0].ContactName != "Jane Smith" {
		t.Errorf("expected ContactName 'Jane Smith', got %s", expired[0].ContactName)
	}
	if expired[0].Email != "jane@beta.com" {
		t.Errorf("expected Email 'jane@beta.com', got %s", expired[0].Email)
	}

	// Test with a date that captures only one expired license
	expired, err = licSvc.GetExpiredLicenses(ctx, "2021-01-01")
	if err != nil {
		t.Fatalf("GetExpiredLicenses: %v", err)
	}
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired license before 2021-01-01, got %d", len(expired))
	}
	if expired[0].CustomerName != "Acme Corp" {
		t.Errorf("expected CustomerName 'Acme Corp', got %s", expired[0].CustomerName)
	}
}

func TestLicenseValidation(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licSvc := license.NewService(db)

	// Create customer and product for tests
	c, _ := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Test Co",
		ContactName:  "Test User",
		Email:        "test@test.com",
	})

	p, _ := prodSvc.Create(ctx, &product.Product{
		ProductName: "Test Product",
		ProductGUID: "TEST-GUID",
	})

	// Required field validation tests
	t.Run("license count must be greater than 0", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-ZERO-COUNT",
			LicenseCount:        0,
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		_, err := licSvc.Create(ctx, lic)
		if !errors.Is(err, license.ErrLicenseCountRequired) {
			t.Errorf("expected ErrLicenseCountRequired, got %v", err)
		}
	})

	t.Run("start date is required", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-NO-START",
			LicenseCount:        1,
			StartDate:           "",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		_, err := licSvc.Create(ctx, lic)
		if !errors.Is(err, license.ErrStartDateRequired) {
			t.Errorf("expected ErrStartDateRequired, got %v", err)
		}
	})

	t.Run("expiration date is required", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-NO-EXP",
			LicenseCount:        1,
			StartDate:           "2024-01-01",
			ExpirationDate:      "",
			MaintExpirationDate: "2099-12-31",
		}

		_, err := licSvc.Create(ctx, lic)
		if !errors.Is(err, license.ErrExpirationDateRequired) {
			t.Errorf("expected ErrExpirationDateRequired, got %v", err)
		}
	})

	t.Run("maintenance expiration date is required", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-NO-MAINT",
			LicenseCount:        1,
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "",
		}

		_, err := licSvc.Create(ctx, lic)
		if !errors.Is(err, license.ErrMaintExpirationRequired) {
			t.Errorf("expected ErrMaintExpirationRequired, got %v", err)
		}
	})

	// Subscription term validation tests
	t.Run("subscription requires term greater than 0", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-SUB-ZERO",
			LicenseCount:        1,
			IsSubscription:      true,
			LicenseTerm:         0,
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		_, err := licSvc.Create(ctx, lic)
		if !errors.Is(err, license.ErrSubscriptionRequiresTerm) {
			t.Errorf("expected ErrSubscriptionRequiresTerm, got %v", err)
		}
	})

	t.Run("subscription with negative term fails", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-SUB-NEG",
			LicenseCount:        1,
			IsSubscription:      true,
			LicenseTerm:         -1,
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		_, err := licSvc.Create(ctx, lic)
		if !errors.Is(err, license.ErrSubscriptionRequiresTerm) {
			t.Errorf("expected ErrSubscriptionRequiresTerm, got %v", err)
		}
	})

	t.Run("perpetual license with term 0 succeeds", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-PERP-ZERO",
			LicenseCount:        1,
			IsSubscription:      false,
			LicenseTerm:         0,
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		created, err := licSvc.Create(ctx, lic)
		if err != nil {
			t.Errorf("perpetual with term 0 should succeed, got %v", err)
		}

		// Clean up
		licSvc.Delete(ctx, created.CustomerID, created.ProductID)
	})

	// Max product version validation tests
	t.Run("invalid max product version fails", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-BAD-VER",
			LicenseCount:        1,
			IsSubscription:      false,
			LicenseTerm:         0,
			MaxProductVersion:   "invalid",
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		_, err := licSvc.Create(ctx, lic)
		if !errors.Is(err, license.ErrInvalidMaxVersion) {
			t.Errorf("expected ErrInvalidMaxVersion, got %v", err)
		}
	})

	t.Run("valid max product version succeeds", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-GOOD-VER",
			LicenseCount:        1,
			IsSubscription:      false,
			LicenseTerm:         0,
			MaxProductVersion:   "1.2.3",
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		created, err := licSvc.Create(ctx, lic)
		if err != nil {
			t.Errorf("valid version should succeed, got %v", err)
		}

		// Clean up
		licSvc.Delete(ctx, created.CustomerID, created.ProductID)
	})

	t.Run("empty max product version succeeds", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-EMPTY-VER",
			LicenseCount:        1,
			IsSubscription:      false,
			LicenseTerm:         0,
			MaxProductVersion:   "",
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		created, err := licSvc.Create(ctx, lic)
		if err != nil {
			t.Errorf("empty version should succeed, got %v", err)
		}

		// Clean up
		licSvc.Delete(ctx, created.CustomerID, created.ProductID)
	})

	t.Run("subscription with valid term succeeds", func(t *testing.T) {
		lic := &license.License{
			CustomerID:          c.CustomerID,
			ProductID:           p.ProductID,
			LicenseKey:          "LIC-SUB-OK",
			LicenseCount:        1,
			IsSubscription:      true,
			LicenseTerm:         12,
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		created, err := licSvc.Create(ctx, lic)
		if err != nil {
			t.Errorf("subscription with valid term should succeed, got %v", err)
		}

		// Clean up
		licSvc.Delete(ctx, created.CustomerID, created.ProductID)
	})
}

func TestLicenseValidationOnUpdate(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licSvc := license.NewService(db)

	// Create customer and product
	c, _ := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Update Test Co",
		ContactName:  "Test User",
		Email:        "test@update.com",
	})

	p, _ := prodSvc.Create(ctx, &product.Product{
		ProductName: "Update Test Product",
		ProductGUID: "UPDATE-GUID",
	})

	// Create a valid license first
	lic := &license.License{
		CustomerID:          c.CustomerID,
		ProductID:           p.ProductID,
		LicenseKey:          "LIC-UPDATE",
		LicenseCount:        1,
		IsSubscription:      false,
		LicenseTerm:         0,
		StartDate:           "2024-01-01",
		ExpirationDate:      "2099-12-31",
		MaintExpirationDate: "2099-12-31",
	}

	created, err := licSvc.Create(ctx, lic)
	if err != nil {
		t.Fatalf("failed to create license: %v", err)
	}

	t.Run("update to invalid subscription fails", func(t *testing.T) {
		updated := &license.License{
			CustomerID:          created.CustomerID,
			ProductID:           created.ProductID,
			LicenseKey:          created.LicenseKey,
			LicenseCount:        1,
			IsSubscription:      true,
			LicenseTerm:         0, // Invalid for subscription
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		err := licSvc.Update(ctx, updated)
		if !errors.Is(err, license.ErrSubscriptionRequiresTerm) {
			t.Errorf("expected ErrSubscriptionRequiresTerm, got %v", err)
		}
	})

	t.Run("update to invalid version fails", func(t *testing.T) {
		updated := &license.License{
			CustomerID:          created.CustomerID,
			ProductID:           created.ProductID,
			LicenseKey:          created.LicenseKey,
			LicenseCount:        1,
			IsSubscription:      false,
			LicenseTerm:         0,
			MaxProductVersion:   "bad-version",
			StartDate:           "2024-01-01",
			ExpirationDate:      "2099-12-31",
			MaintExpirationDate: "2099-12-31",
		}

		err := licSvc.Update(ctx, updated)
		if !errors.Is(err, license.ErrInvalidMaxVersion) {
			t.Errorf("expected ErrInvalidMaxVersion, got %v", err)
		}
	})
}
