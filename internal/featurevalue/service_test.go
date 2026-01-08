package featurevalue_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestFeatureValueLifecycle(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licenseSvc := license.NewService(db)
	featSvc := feature.NewService(db)
	fvSvc := featurevalue.NewService(db)

	// Create customer
	c, err := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Acme",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	// Create product
	p, err := prodSvc.Create(ctx, &product.Product{
		ProductName: "Widget",
		ProductGUID: "test-widget-guid",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	// Create license (required for foreign key)
	_, err = licenseSvc.Create(ctx, &license.License{
		CustomerID:   c.CustomerID,
		ProductID:    p.ProductID,
		LicenseCount: 1,
	})
	if err != nil {
		t.Fatalf("create license: %v", err)
	}

	// Create feature definitions
	feat1, err := featSvc.Create(ctx, &feature.Feature{
		ProductID:    p.ProductID,
		FeatureName:  "Feature1",
		FeatureType:  1, // String
		DefaultValue: "default1",
	})
	if err != nil {
		t.Fatalf("create feature 1: %v", err)
	}

	feat2, err := featSvc.Create(ctx, &feature.Feature{
		ProductID:    p.ProductID,
		FeatureName:  "Feature2",
		FeatureType:  1, // String
		DefaultValue: "default2",
	})
	if err != nil {
		t.Fatalf("create feature 2: %v", err)
	}

	// Insert initial feature value overrides
	initial := []featurevalue.FeatureValue{
		{CustomerID: c.CustomerID, ProductID: p.ProductID, FeatureID: feat1.FeatureID, FeatureValue: "A"},
		{CustomerID: c.CustomerID, ProductID: p.ProductID, FeatureID: feat2.FeatureID, FeatureValue: "B"},
	}

	tx := db.MustBeginTx(ctx, nil)
	for _, fv := range initial {
		_, err := tx.Exec(`
            INSERT INTO license_feature
                (customer_id, product_id, feature_id, feature_value)
            VALUES (?, ?, ?, ?)
        `, fv.CustomerID, fv.ProductID, fv.FeatureID, fv.FeatureValue)
		if err != nil {
			tx.Rollback()
			t.Fatalf("insert feature value: %v", err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify GetFeatureValues returns both rows
	values, err := fvSvc.GetFeatureValues(ctx, c.CustomerID, p.ProductID)
	if err != nil {
		t.Fatalf("GetFeatureValues: %v", err)
	}

	if len(values) != 2 {
		t.Fatalf("expected 2 feature values, got %d", len(values))
	}

	// Update one feature value
	updated := &featurevalue.FeatureValue{
		CustomerID:   c.CustomerID,
		ProductID:    p.ProductID,
		FeatureID:    feat1.FeatureID,
		FeatureValue: "Updated",
	}

	if err := fvSvc.Update(ctx, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// Verify update persisted
	values, err = fvSvc.GetFeatureValues(ctx, c.CustomerID, p.ProductID)
	if err != nil {
		t.Fatalf("GetFeatureValues: %v", err)
	}

	var found bool
	for _, fv := range values {
		if fv.FeatureID == feat1.FeatureID {
			found = true
			if fv.FeatureValue != "Updated" {
				t.Fatalf("expected Updated, got %s", fv.FeatureValue)
			}
		}
	}

	if !found {
		t.Fatalf("feature ID %d not found after update", feat1.FeatureID)
	}
}

// TestFeatureValueUpsert verifies the override-only pattern:
// - Update inserts a new row when no override exists (UPSERT insert path)
// - Update modifies an existing row when override exists (UPSERT update path)
func TestFeatureValueUpsert(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licenseSvc := license.NewService(db)
	featSvc := feature.NewService(db)
	fvSvc := featurevalue.NewService(db)

	// Create customer
	c, err := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Test Customer",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	// Create product
	p, err := prodSvc.Create(ctx, &product.Product{
		ProductName: "Test Product",
		ProductGUID: "test-guid-123",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	// Create license (required for foreign key)
	_, err = licenseSvc.Create(ctx, &license.License{
		CustomerID:   c.CustomerID,
		ProductID:    p.ProductID,
		LicenseCount: 1,
	})
	if err != nil {
		t.Fatalf("create license: %v", err)
	}

	// Create a feature definition with default value
	feat, err := featSvc.Create(ctx, &feature.Feature{
		ProductID:    p.ProductID,
		FeatureName:  "MaxUsers",
		FeatureType:  0, // Integer
		DefaultValue: "10",
	})
	if err != nil {
		t.Fatalf("create feature: %v", err)
	}

	// Verify no override exists yet
	values, err := fvSvc.GetFeatureValues(ctx, c.CustomerID, p.ProductID)
	if err != nil {
		t.Fatalf("GetFeatureValues: %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("expected 0 feature values initially, got %d", len(values))
	}

	// Test INSERT path: Update when no row exists should create one
	override := &featurevalue.FeatureValue{
		CustomerID:   c.CustomerID,
		ProductID:    p.ProductID,
		FeatureID:    feat.FeatureID,
		FeatureValue: "50",
	}
	if err := fvSvc.Update(ctx, override); err != nil {
		t.Fatalf("Update (insert path): %v", err)
	}

	// Verify the override was created
	values, err = fvSvc.GetFeatureValues(ctx, c.CustomerID, p.ProductID)
	if err != nil {
		t.Fatalf("GetFeatureValues after insert: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 feature value after insert, got %d", len(values))
	}
	if values[0].FeatureValue != "50" {
		t.Fatalf("expected value '50', got '%s'", values[0].FeatureValue)
	}

	// Test UPDATE path: Update when row exists should modify it
	override.FeatureValue = "100"
	if err := fvSvc.Update(ctx, override); err != nil {
		t.Fatalf("Update (update path): %v", err)
	}

	// Verify the override was updated
	values, err = fvSvc.GetFeatureValues(ctx, c.CustomerID, p.ProductID)
	if err != nil {
		t.Fatalf("GetFeatureValues after update: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 feature value after update, got %d", len(values))
	}
	if values[0].FeatureValue != "100" {
		t.Fatalf("expected value '100', got '%s'", values[0].FeatureValue)
	}
}
