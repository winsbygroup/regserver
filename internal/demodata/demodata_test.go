package demodata_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"winsbygroup.com/regserver/internal/demodata"
	"winsbygroup.com/regserver/internal/sqlite"
)

// TestDemoDataNotLoadedOnExistingDB verifies that demo data is only loaded
// when the database is newly created, not when it already exists.
// This mirrors the logic in server.Build() that checks isNewDB before loading.
func TestDemoDataNotLoadedOnExistingDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Step 1: Create database and add existing data
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := sqlite.RunMigrations(db.DB); err != nil {
		db.Close()
		t.Fatalf("migrate: %v", err)
	}

	// Insert a customer that should NOT be overwritten
	_, err = db.Exec(`INSERT INTO customer (customer_name, contact_name) VALUES ('Existing Corp', 'Original Contact')`)
	if err != nil {
		db.Close()
		t.Fatalf("insert existing customer: %v", err)
	}

	db.Close()

	// Step 2: Simulate server.Build() logic - check if DB exists BEFORE opening
	isNewDB := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		isNewDB = true
	}

	if isNewDB {
		t.Fatal("expected isNewDB to be false for existing database")
	}

	// Step 3: Reopen database (simulating server startup)
	db, err = sqlx.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer db.Close()

	// Step 4: Simulate DemoMode=true with existing DB - should NOT load demo data
	demoMode := true
	if demoMode && isNewDB {
		// This block should NOT execute for existing DB
		if err := demodata.Load(db.DB); err != nil {
			t.Fatalf("load demo data: %v", err)
		}
	}

	// Step 5: Verify original data is intact (demo data was NOT loaded)
	var customerName string
	err = db.QueryRow(`SELECT customer_name FROM customer WHERE customer_name = 'Existing Corp'`).Scan(&customerName)
	if err != nil {
		t.Fatalf("existing customer should still exist: %v", err)
	}

	// Verify demo data was NOT loaded
	var demoCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM customer WHERE customer_name = 'Acme Corporation'`).Scan(&demoCount)
	if err != nil {
		t.Fatalf("query demo customer: %v", err)
	}
	if demoCount != 0 {
		t.Error("demo data should NOT have been loaded on existing database")
	}
}

// TestDemoDataLoadedOnNewDB verifies that demo data IS loaded on a fresh database.
func TestDemoDataLoadedOnNewDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "newtest.db")

	// Step 1: Check if DB exists BEFORE creating it
	isNewDB := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		isNewDB = true
	}

	if !isNewDB {
		t.Fatal("expected isNewDB to be true for non-existent database")
	}

	// Step 2: Create and open database
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := sqlite.RunMigrations(db.DB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Step 3: Simulate DemoMode=true with new DB - SHOULD load demo data
	demoMode := true
	if demoMode && isNewDB {
		if err := demodata.Load(db.DB); err != nil {
			t.Fatalf("load demo data: %v", err)
		}
	}

	// Step 4: Verify demo data was loaded
	var demoCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM customer WHERE customer_name = 'Acme Corporation'`).Scan(&demoCount)
	if err != nil {
		t.Fatalf("query demo customer: %v", err)
	}
	if demoCount != 1 {
		t.Error("demo data should have been loaded on new database")
	}
}
