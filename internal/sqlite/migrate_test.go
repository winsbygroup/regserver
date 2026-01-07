package sqlite_test

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"winsbygroup.com/regserver/internal/sqlite"
)

func TestMigrationsApplyCleanly(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := sqlite.RunMigrations(db); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	// Verify a table exists
	row := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='customer';`)
	var name string
	if err := row.Scan(&name); err != nil {
		t.Fatalf("expected customer table to exist: %v", err)
	}
}

func TestMigrationsSetsApplicationID(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := sqlite.RunMigrations(db); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	var appID int
	if err := db.QueryRow("PRAGMA application_id;").Scan(&appID); err != nil {
		t.Fatalf("read application_id: %v", err)
	}

	if appID != sqlite.ApplicationID {
		t.Errorf("expected application_id 0x%X, got 0x%X", sqlite.ApplicationID, appID)
	}
}

func TestVerifyApplicationID(t *testing.T) {
	t.Run("accepts new database with appID 0", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		defer db.Close()

		// New database has application_id = 0
		if err := sqlite.VerifyApplicationID(db); err != nil {
			t.Errorf("expected no error for new database, got %v", err)
		}
	})

	t.Run("accepts regserver database", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		defer db.Close()

		// Run migrations to set application_id
		if err := sqlite.RunMigrations(db); err != nil {
			t.Fatalf("migrations failed: %v", err)
		}

		if err := sqlite.VerifyApplicationID(db); err != nil {
			t.Errorf("expected no error for regserver database, got %v", err)
		}
	})

	t.Run("rejects database with wrong appID", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		defer db.Close()

		// Run migrations first to create a valid database, then change the appID
		if err := sqlite.RunMigrations(db); err != nil {
			t.Fatalf("migrations failed: %v", err)
		}
		// Simulate database with wrong appID
		if _, err := db.Exec("PRAGMA application_id = 305419896;"); err != nil { // 0x12345678
			t.Fatalf("set application_id: %v", err)
		}

		err = sqlite.VerifyApplicationID(db)
		if err == nil {
			t.Fatal("expected error for wrong application_id, got nil")
		}
		if !errors.Is(err, sqlite.ErrInvalidDatabase) {
			t.Errorf("expected ErrInvalidDatabase, got %v", err)
		}
	})

	t.Run("rejects database with tables but no appID", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		defer db.Close()

		// Simulate another app's database that never set application_id
		if _, err := db.Exec("CREATE TABLE other_app (id INTEGER);"); err != nil {
			t.Fatalf("create table: %v", err)
		}

		err = sqlite.VerifyApplicationID(db)
		if err == nil {
			t.Fatal("expected error for database with tables but no appID, got nil")
		}
		if !errors.Is(err, sqlite.ErrInvalidDatabase) {
			t.Errorf("expected ErrInvalidDatabase, got %v", err)
		}
	})
}
