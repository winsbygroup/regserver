package activation_test

import (
	"context"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/activation"
	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/registration"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestActivate_LicenseCountCheck(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	// Create services
	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licenseSvc := license.NewService(db)
	machineSvc := machine.NewService(db)
	regSvc := registration.NewService(db)
	featureSvc := feature.NewService(db)
	fvSvc := featurevalue.NewService(db)

	activationSvc := activation.NewService(
		db,
		"test-secret",
		custSvc,
		machineSvc,
		regSvc,
		licenseSvc,
		prodSvc,
		featureSvc,
		fvSvc,
	)

	// Create test customer
	cust, err := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Test Company",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	// Create test product
	prod, err := prodSvc.Create(ctx, &product.Product{
		ProductName:   "Test Product",
		ProductGUID:   "TEST-GUID-123",
		LatestVersion: "1.0.0",
		DownloadURL:   "http://example.com/download",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	// Create license with license count of 2
	_, err = licenseSvc.Create(ctx, &license.License{
		CustomerID:          cust.CustomerID,
		ProductID:           prod.ProductID,
		LicenseKey:          "REG-GUID-123",
		LicenseCount:        2,
		IsSubscription:      false,
		LicenseTerm:         365,
		StartDate:           time.Now().Format("2006-01-02"),
		ExpirationDate:      futureDate,
		MaintExpirationDate: futureDate,
		MaxProductVersion:   "99.0.0",
	})
	if err != nil {
		t.Fatalf("create customer product: %v", err)
	}

	t.Run("first activation succeeds", func(t *testing.T) {
		req := &activation.Request{
			MachineCode: "MACHINE-001",
			UserName:    "user1",
		}
		resp, err := activationSvc.Activate(ctx, cust.CustomerID, prod.ProductID, req)
		if err != nil {
			t.Fatalf("first activation should succeed: %v", err)
		}
		if resp.UserCompany != "Test Company" {
			t.Errorf("expected UserCompany 'Test Company', got %q", resp.UserCompany)
		}
	})

	t.Run("second activation succeeds (at limit)", func(t *testing.T) {
		req := &activation.Request{
			MachineCode: "MACHINE-002",
			UserName:    "user2",
		}
		_, err := activationSvc.Activate(ctx, cust.CustomerID, prod.ProductID, req)
		if err != nil {
			t.Fatalf("second activation should succeed: %v", err)
		}
	})

	t.Run("third activation fails (over limit)", func(t *testing.T) {
		req := &activation.Request{
			MachineCode: "MACHINE-003",
			UserName:    "user3",
		}
		_, err := activationSvc.Activate(ctx, cust.CustomerID, prod.ProductID, req)
		if err == nil {
			t.Fatal("third activation should fail due to license count exceeded")
		}
		if !strings.Contains(err.Error(), "license count exceeded") {
			t.Errorf("expected 'license count exceeded' error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "2 of 2") {
			t.Errorf("expected error to mention '2 of 2', got: %v", err)
		}
	})

	t.Run("re-activation of existing machine succeeds at limit", func(t *testing.T) {
		// Machine 1 should be able to re-activate even though we're at the limit
		req := &activation.Request{
			MachineCode: "MACHINE-001",
			UserName:    "user1-updated",
		}
		resp, err := activationSvc.Activate(ctx, cust.CustomerID, prod.ProductID, req)
		if err != nil {
			t.Fatalf("re-activation should succeed: %v", err)
		}
		if resp.UserName != "user1-updated" {
			t.Errorf("expected updated username, got %q", resp.UserName)
		}
	})
}

func TestActivate_NoLicense(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	// Create services
	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licenseSvc := license.NewService(db)
	machineSvc := machine.NewService(db)
	regSvc := registration.NewService(db)
	featureSvc := feature.NewService(db)
	fvSvc := featurevalue.NewService(db)

	activationSvc := activation.NewService(
		db,
		"test-secret",
		custSvc,
		machineSvc,
		regSvc,
		licenseSvc,
		prodSvc,
		featureSvc,
		fvSvc,
	)

	// Create test customer
	cust, _ := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Test Company",
	})

	// Create test product
	prod, _ := prodSvc.Create(ctx, &product.Product{
		ProductName:   "Test Product",
		ProductGUID:   "TEST-GUID-123",
		LatestVersion: "1.0.0",
		DownloadURL:   "http://example.com/download",
	})

	// Don't create customer product - no license

	req := &activation.Request{
		MachineCode: "MACHINE-001",
		UserName:    "user1",
	}
	_, err := activationSvc.Activate(ctx, cust.CustomerID, prod.ProductID, req)
	if err == nil {
		t.Fatal("activation without license should fail")
	}
	// Error message could be "no license" or "customer_product not found"
	errStr := err.Error()
	if !strings.Contains(errStr, "no license") && !strings.Contains(errStr, "not found") {
		t.Errorf("expected license error, got: %v", err)
	}
}

func TestActivate_ExpiredMachineDoesNotCount(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	// Create services
	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licenseSvc := license.NewService(db)
	machineSvc := machine.NewService(db)
	regSvc := registration.NewService(db)
	featureSvc := feature.NewService(db)
	fvSvc := featurevalue.NewService(db)

	activationSvc := activation.NewService(
		db,
		"test-secret",
		custSvc,
		machineSvc,
		regSvc,
		licenseSvc,
		prodSvc,
		featureSvc,
		fvSvc,
	)

	// Create test customer and product
	cust, _ := custSvc.Create(ctx, &customer.Customer{CustomerName: "Test Company"})
	prod, _ := prodSvc.Create(ctx, &product.Product{
		ProductName:   "Test Product",
		ProductGUID:   "TEST-GUID-123",
		LatestVersion: "1.0.0",
		DownloadURL:   "http://example.com/download",
	})

	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02") // yesterday

	// Create customer product with license count of 1
	_, _ = licenseSvc.Create(ctx, &license.License{
		CustomerID:          cust.CustomerID,
		ProductID:           prod.ProductID,
		LicenseKey:    "REG-GUID-123",
		LicenseCount:        1,
		IsSubscription:      false,
		LicenseTerm:         365,
		StartDate:           time.Now().Format("2006-01-02"),
		ExpirationDate:      futureDate,
		MaintExpirationDate: futureDate,
		MaxProductVersion:   "99.0.0",
	})

	// Manually create an expired registration for machine 1
	// First, create the machine via GetOrCreate in a transaction
	tx, _ := db.Beginx()
	machineID, _ := machineSvc.GetOrCreate(ctx, tx, cust.CustomerID, "EXPIRED-MACHINE", "expired-user")
	// Create expired registration
	_ = regSvc.Upsert(ctx, tx, &registration.Registration{
		MachineID:             machineID,
		ProductID:             prod.ProductID,
		ExpirationDate:        pastDate,
		RegistrationHash:      "EXPIRED-MACHINE",
		FirstRegistrationDate: pastDate,
		LastRegistrationDate:  pastDate,
	})
	tx.Commit()

	// Now a new machine should be able to activate (expired one doesn't count)
	req := &activation.Request{
		MachineCode: "NEW-MACHINE",
		UserName:    "new-user",
	}
	_, err := activationSvc.Activate(ctx, cust.CustomerID, prod.ProductID, req)
	if err != nil {
		t.Fatalf("new machine should be able to activate when only existing machine is expired: %v", err)
	}

	// But a second new machine should fail
	req2 := &activation.Request{
		MachineCode: "ANOTHER-NEW-MACHINE",
		UserName:    "another-user",
	}
	_, err = activationSvc.Activate(ctx, cust.CustomerID, prod.ProductID, req2)
	if err == nil {
		t.Fatal("should fail when license count exceeded")
	}
	if !strings.Contains(err.Error(), "license count exceeded") {
		t.Errorf("expected 'license count exceeded' error, got: %v", err)
	}
}
