package testutil

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"winsbygroup.com/regserver/internal/sqlite"
)

func NewTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	return NewTestDBAt(t, filepath.Join(t.TempDir(), "test.db"))
}

func NewTestDBAt(t *testing.T, dbPath string) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	// Register cleanup immediately
	t.Cleanup(func() {
		db.Close()
	})

	// DELETE mode for tests
	if _, err := db.Exec(`PRAGMA journal_mode=DELETE;`); err != nil {
		t.Fatalf("set journal mode: %v", err)
	}

	// Enable foreign key enforcement
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	// Verify foreign keys are supported and enabled
	var fkEnabled int
	if err := db.QueryRow(`PRAGMA foreign_keys;`).Scan(&fkEnabled); err != nil {
		t.Fatalf("foreign key support check failed: %v", err)
	}
	if fkEnabled != 1 {
		t.Fatal("SQLite foreign keys not supported (requires SQLite 3.6.19+)")
	}

	// Run migrations
	if err := sqlite.RunMigrations(db.DB); err != nil {
		// Ensure DB closes even on failure
		db.Close()
		t.Fatalf("migrate: %v", err)
	}

	return db
}
