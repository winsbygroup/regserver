package backup_test

import (
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/backup"
	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestBackupService(t *testing.T) {
	ctx := context.Background()

	// Create a temp directory for the database
	tmpDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test database in temp directory
	dbPath := filepath.Join(tmpDir, "test.db")
	db := testutil.NewTestDBAt(t, dbPath)

	// Add some test data
	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	licSvc := license.NewService(db)

	c, err := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Test Customer",
		ContactName:  "John Doe",
		Email:        "john@test.com",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	p, err := prodSvc.Create(ctx, &product.Product{
		ProductName:   "Test Product",
		ProductGUID:   "TEST-GUID-123",
		LatestVersion: "1.0.0",
		DownloadURL:   "https://example.com/download",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	_, err = licSvc.Create(ctx, &license.License{
		CustomerID:          c.CustomerID,
		ProductID:           p.ProductID,
		LicenseKey:          "LIC-TEST-KEY",
		LicenseCount:        5,
		StartDate:           "2024-01-01",
		ExpirationDate:      "2099-12-31",
		MaintExpirationDate: "2099-12-31",
	})
	if err != nil {
		t.Fatalf("create license: %v", err)
	}

	// Create backup
	backupSvc := backup.NewService(db, dbPath)
	result, err := backupSvc.CreateBackup(ctx)
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	// Verify result
	if result.Filename == "" {
		t.Error("expected filename to be set")
	}
	if result.Size == 0 {
		t.Error("expected size > 0")
	}
	if !strings.HasSuffix(result.Filename, "_regdump.sql.gz") {
		t.Errorf("expected filename to end with _regdump.sql.gz, got %s", result.Filename)
	}

	// Verify backup file exists and decompress to check content
	file, err := os.Open(result.Path)
	if err != nil {
		t.Fatalf("open backup file: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("create gzip reader: %v", err)
	}
	defer gzReader.Close()

	content, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("read gzip content: %v", err)
	}

	dump := string(content)

	// Check for expected content
	if !strings.Contains(dump, "CREATE TABLE") {
		t.Error("expected dump to contain CREATE TABLE statements")
	}
	if !strings.Contains(dump, "Test Customer") {
		t.Error("expected dump to contain customer data")
	}
	if !strings.Contains(dump, "Test Product") {
		t.Error("expected dump to contain product data")
	}
	if !strings.Contains(dump, "lic-test-key") {
		t.Error("expected dump to contain license key")
	}
	if !strings.Contains(dump, "PRAGMA journal_mode=WAL") {
		t.Error("expected dump to contain WAL pragma")
	}
	if !strings.Contains(dump, "BEGIN TRANSACTION") {
		t.Error("expected dump to contain BEGIN TRANSACTION")
	}
	if !strings.Contains(dump, "COMMIT") {
		t.Error("expected dump to contain COMMIT")
	}
}
