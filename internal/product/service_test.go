package product_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/sqlite"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestProductLifecycle(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	svc := product.NewService(db)

	// -------------------------
	// Create
	// -------------------------
	p := &product.Product{
		ProductName:   "Test Product",
		ProductGUID:   "ABC-123",
		LatestVersion: "1.0.0",
		DownloadURL:   "https://example.com/download",
	}

	created, err := svc.Create(ctx, p)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if created.ProductID == 0 {
		t.Fatalf("expected ProductID to be assigned")
	}

	// -------------------------
	// Get
	// -------------------------
	got, err := svc.Get(ctx, created.ProductID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.ProductName != "Test Product" {
		t.Errorf("expected name %q, got %q", "Test Product", got.ProductName)
	}

	// -------------------------
	// Update
	// -------------------------
	got.ProductName = "Updated Name"
	got.LatestVersion = "2.0.0"

	if err := svc.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Fetch again to verify update
	updated, err := svc.Get(ctx, got.ProductID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}

	if updated.ProductName != "Updated Name" {
		t.Errorf("expected updated name %q, got %q", "Updated Name", updated.ProductName)
	}

	if updated.LatestVersion != "2.0.0" {
		t.Errorf("expected updated version %q, got %q", "2.0.0", updated.LatestVersion)
	}

	// -------------------------
	// Delete
	// -------------------------
	if err := svc.Delete(ctx, updated.ProductID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Should now return error
	_, err = svc.Get(ctx, updated.ProductID)
	if err == nil {
		t.Fatalf("expected error getting deleted product")
	}
}

func TestProductVersionValidation(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	svc := product.NewService(db)

	tests := []struct {
		name        string
		version     string
		shouldError bool
	}{
		{"valid 1.0.0", "1.0.0", false},
		{"valid 2.3.4", "2.3.4", false},
		{"valid 10.20.30", "10.20.30", false},
		{"empty allowed", "", false},
		{"invalid 1.0", "1.0", true},
		{"invalid 1.0.0.0", "1.0.0.0", true},
		{"invalid v1.0.0", "v1.0.0", true},
		{"invalid 1.0.0-beta", "1.0.0-beta", true},
		{"invalid abc", "abc", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &product.Product{
				ProductName:   "Test " + tc.name,
				ProductGUID:   "GUID-" + tc.name,
				LatestVersion: tc.version,
				DownloadURL:   "https://example.com",
			}

			_, err := svc.Create(ctx, p)

			if tc.shouldError && err == nil {
				t.Errorf("expected error for version %q, got none", tc.version)
			}
			if !tc.shouldError && err != nil {
				t.Errorf("unexpected error for version %q: %v", tc.version, err)
			}
		})
	}
}

func TestProductDuplicateNameCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	svc := product.NewService(db)

	// Create first product
	_, err := svc.Create(ctx, &product.Product{
		ProductName:   "Widget Pro",
		ProductGUID:   "GUID-123",
		LatestVersion: "1.0.0",
		DownloadURL:   "https://example.com",
	})
	if err != nil {
		t.Fatalf("create first product: %v", err)
	}

	// Try to create product with same name but different case
	testCases := []string{"WIDGET PRO", "widget pro", "WiDgEt PrO"}
	for _, name := range testCases {
		_, err = svc.Create(ctx, &product.Product{
			ProductName:   name,
			ProductGUID:   "GUID-" + name,
			LatestVersion: "1.0.0",
			DownloadURL:   "https://example.com",
		})
		if err == nil {
			t.Errorf("expected error for duplicate name %q, got none", name)
		}
		if !sqlite.IsUniqueConstraintError(err) {
			t.Errorf("expected unique constraint error for %q, got: %v", name, err)
		}
	}
}

func TestProductDuplicateGUIDCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	svc := product.NewService(db)

	// Create first product
	_, err := svc.Create(ctx, &product.Product{
		ProductName:   "Widget Pro",
		ProductGUID:   "abc-123-def",
		LatestVersion: "1.0.0",
		DownloadURL:   "https://example.com",
	})
	if err != nil {
		t.Fatalf("create first product: %v", err)
	}

	// Try to create product with same GUID but different case
	testCases := []string{"ABC-123-DEF", "Abc-123-Def"}
	for _, guid := range testCases {
		_, err = svc.Create(ctx, &product.Product{
			ProductName:   "Other Product " + guid,
			ProductGUID:   guid,
			LatestVersion: "1.0.0",
			DownloadURL:   "https://example.com",
		})
		if err == nil {
			t.Errorf("expected error for duplicate GUID %q, got none", guid)
		}
		if !sqlite.IsUniqueConstraintError(err) {
			t.Errorf("expected unique constraint error for %q, got: %v", guid, err)
		}
	}
}
