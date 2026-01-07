package feature_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/sqlite"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestFeatureLifecycle(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	prodSvc := product.NewService(db)
	featSvc := feature.NewService(db)

	// Create a product first
	p, err := prodSvc.Create(ctx, &product.Product{
		ProductName:   "Widget",
		ProductGUID:   "GUID-123",
		LatestVersion: "1.0.0",
		DownloadURL:   "https://example.com",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	// Create feature
	f := &feature.Feature{
		ProductID:     p.ProductID,
		FeatureName:   "MaxUsers",
		FeatureType:   feature.ToInt("values"),
		AllowedValues: "1,5,10",
		DefaultValue:  "1",
	}

	created, err := featSvc.Create(ctx, f)
	if err != nil {
		t.Fatalf("create feature: %v", err)
	}

	// Get features for product
	list, err := featSvc.GetForProduct(ctx, p.ProductID)
	if err != nil {
		t.Fatalf("get features: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(list))
	}

	// Delete
	if err := featSvc.Delete(ctx, created.FeatureID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	list, _ = featSvc.GetForProduct(ctx, p.ProductID)
	if len(list) != 0 {
		t.Fatalf("expected 0 features after delete, got %d", len(list))
	}
}

func TestFeatureDuplicateName(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	prodSvc := product.NewService(db)
	featSvc := feature.NewService(db)

	// Create a product
	p, err := prodSvc.Create(ctx, &product.Product{
		ProductName:   "Widget",
		ProductGUID:   "GUID-123",
		LatestVersion: "1.0.0",
		DownloadURL:   "https://example.com",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	// Create first feature
	_, err = featSvc.Create(ctx, &feature.Feature{
		ProductID:    p.ProductID,
		FeatureName:  "MaxUsers",
		FeatureType:  0,
		DefaultValue: "10",
	})
	if err != nil {
		t.Fatalf("create first feature: %v", err)
	}

	// Try to create duplicate feature name for same product
	_, err = featSvc.Create(ctx, &feature.Feature{
		ProductID:    p.ProductID,
		FeatureName:  "MaxUsers",
		FeatureType:  0,
		DefaultValue: "20",
	})
	if err == nil {
		t.Fatal("expected error for duplicate feature name, got none")
	}
	if !sqlite.IsUniqueConstraintError(err) {
		t.Errorf("expected unique constraint error, got: %v", err)
	}
}
