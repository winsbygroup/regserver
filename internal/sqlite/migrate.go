package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/GuiaBolso/darwin"
	_ "github.com/mattn/go-sqlite3"
)

// ApplicationID is the SQLite application_id for regserver databases.
// "REGS" in ASCII: R=0x52, E=0x45, G=0x47, S=0x53
const ApplicationID = 0x52454753

// ErrInvalidDatabase is returned when the database is not a valid regserver database.
var ErrInvalidDatabase = errors.New("not a valid 'regserver' database")

// defineMigrations returns a slice of database migrations
// Each migration is defined in a separate row (versioned by major db release)
// comments must only appear after sql on a line and cannot span lines (comments are stripped before checksum calc)
// *NEVER* change/remove a step once released! (because a checksum of the script is saved with the migration)
func defineMigrations() []darwin.Migration {
	m := []darwin.Migration{

		// Each database change release is given a major version number (1.xx, 2.xx) with minor numbers (x.01, x.02)
		// representing the actual migration steps within that release. Version numbers must be ascending.

		// Set application_id first to identify this as a regserver database
		// 0x52454753 = "REGS" in ASCII (R=0x52, E=0x45, G=0x47, S=0x53)
		{Version: 1.00, Description: "Set application_id", Script: `
		PRAGMA application_id = 0x52454753;`},

		{Version: 1.01, Description: "Create Table 'customer'", Script: `
		CREATE TABLE IF NOT EXISTS customer (
			customer_id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_name VARCHAR(255) NOT NULL UNIQUE COLLATE NOCASE,
			contact_name VARCHAR(255),
			phone VARCHAR(255),
			email VARCHAR(255),
			notes TEXT
		);`},

		{Version: 1.02, Description: "Create Table 'machine'", Script: `
		CREATE TABLE IF NOT EXISTS machine (
			machine_id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			machine_code VARCHAR(255) NOT NULL,
			user_name VARCHAR(255),
			FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON DELETE CASCADE
		);`},

		{Version: 1.03, Description: "Create Index 'idx_machine_customer_id'", Script: `
		CREATE INDEX IF NOT EXISTS idx_machine_customer_id ON machine (customer_id ASC);`},

		{Version: 1.04, Description: "Create Table 'product'", Script: `
		CREATE TABLE IF NOT EXISTS product (
			product_id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_name VARCHAR(255) NOT NULL UNIQUE COLLATE NOCASE,
			product_guid VARCHAR(36) NOT NULL UNIQUE COLLATE NOCASE,
			latest_version VARCHAR(10) NOT NULL,
			download_url VARCHAR(255) NOT NULL
		);`},

		{Version: 1.05, Description: "Create Table 'registration'", Script: `
		CREATE TABLE IF NOT EXISTS registration (
				machine_id INTEGER NOT NULL,
				product_id INTEGER NOT NULL,
				expiration_date VARCHAR(10) NOT NULL,
				registration_hash CHAR(28) NOT NULL,
				first_registration_date VARCHAR(10),
				last_registration_date VARCHAR(10),
				installed_version VARCHAR(20) NOT NULL DEFAULT '',
				CONSTRAINT pk_registration PRIMARY KEY (machine_id, product_id),
				FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE,
				FOREIGN KEY (machine_id) REFERENCES machine (machine_id) ON DELETE CASCADE
		);`},

		{Version: 1.06, Description: "Create Index 'idx_registration_machine_id'", Script: `
		CREATE INDEX IF NOT EXISTS idx_registration_machine_id ON registration (machine_id ASC);`},

		{Version: 1.07, Description: "Create Index 'idx_registration_product_id'", Script: `
		CREATE INDEX IF NOT EXISTS idx_registration_product_id ON registration (product_id ASC);`},

		{Version: 1.08, Description: "Create Table 'license'", Script: `
		CREATE TABLE IF NOT EXISTS license (
			customer_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			license_key VARCHAR(36) NOT NULL COLLATE NOCASE,
			license_count INTEGER NOT NULL,
			is_subscription INTEGER NOT NULL,
			license_term INTEGER NOT NULL,
			start_date VARCHAR(10),
			expiration_date VARCHAR(10),
			maint_expiration_date VARCHAR(10) NOT NULL DEFAULT '9999-12-31',
			max_product_version VARCHAR(255),
			CONSTRAINT pk_license PRIMARY KEY (customer_id, product_id),
			FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
		);`},

		{Version: 1.09, Description: "Create Index 'idx_license_customer_id'", Script: `
		CREATE INDEX IF NOT EXISTS idx_license_customer_id ON license (customer_id ASC);`},

		{Version: 1.10, Description: "Create Index 'idx_license_product_id'", Script: `
		CREATE INDEX IF NOT EXISTS idx_license_product_id ON license (product_id ASC);`},

		{Version: 1.11, Description: "Create Unique Index 'idx_license_key'", Script: `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_license_key ON license (license_key);`},

		{Version: 1.12, Description: "Create Table 'feature'", Script: `
		CREATE TABLE IF NOT EXISTS feature (
			feature_id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL,
			feature_name VARCHAR(255) NOT NULL,
			feature_type INTEGER NOT NULL CHECK (feature_type in (0,1,2)) DEFAULT 0,
			allowed_values VARCHAR(255),
			default_value VARCHAR(255),
			FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
		);`},

		{Version: 1.13, Description: "Create Index 'idx_feature_product_id'", Script: `
		CREATE INDEX IF NOT EXISTS idx_feature_product_id ON feature (product_id ASC);`},

		{Version: 1.14, Description: "Create Unique Index 'idx_feature_product_name'", Script: `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_feature_product_name ON feature (product_id, feature_name COLLATE NOCASE);`},

		{Version: 1.15, Description: "Create Table 'license_feature'", Script: `
		CREATE TABLE IF NOT EXISTS license_feature (
			customer_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			feature_id INTEGER NOT NULL,
			feature_value VARCHAR(255) NOT NULL,
			CONSTRAINT pk_license_feature PRIMARY KEY (customer_id, product_id, feature_id),
			FOREIGN KEY (feature_id) REFERENCES feature (feature_id),
			FOREIGN KEY (customer_id, product_id) REFERENCES license (customer_id, product_id) ON DELETE CASCADE
		);`},

		{Version: 1.16, Description: "Create Index 'idx_licfeat_feature_id'", Script: `
		CREATE INDEX IF NOT EXISTS idx_licfeat_feature_id ON license_feature (feature_id ASC);`},

		{Version: 1.17, Description: "Create Index 'idx_licfeat_custid_prodid'", Script: `
		CREATE INDEX IF NOT EXISTS idx_licfeat_custid_prodid ON license_feature (customer_id ASC, product_id ASC);`},
	}
	return m
}

// changes returns a user-friendly display of database version changes
func changes(v1, v2 float64) string {
	if v1 != v2 {
		return fmt.Sprintf("DB Version: %.2f (migrated from %.2f to %.2f)", v2, v1, v2)
	}
	return fmt.Sprintf("DB Version: %.2f", v1)
}

// currentVersion reads from migration table to get the latest version and number of steps applied
func currentVersion(db *sql.DB) (count int, ver float64, err error) {
	// might not have any migrations yet...
	s := `select count(*) as n from sqlite_master where tbl_name = 'darwin_migrations';`
	err = db.QueryRow(s).Scan(&count)
	if err != nil || count == 0 {
		return 0, 0, err
	}

	s = `select count(*) as n, max(version) as ver from darwin_migrations;`
	err = db.QueryRow(s).Scan(&count, &ver)
	return count, ver, err
}

// minifiedMigrations returns our migrations with minified scripts so comments or formatting changes
// will not generate a new checksum
func minifiedMigrations() []darwin.Migration {
	migrations := defineMigrations()
	for i := range migrations {
		migrations[i].Script = minify(migrations[i].Script)
	}
	return migrations
}

// minify simplifies the script to keep certain changes (spaces, tabs, case and comments) from
// creating a new checksum
func minify(script string) string {
	b := strings.Builder{}
	s := strings.ToLower(strings.ReplaceAll(script, "/*", "--"))
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if i := strings.Index(line, "--"); i != -1 {
			line = line[0:i]
		}
		b.WriteString(strings.TrimSpace(line) + "\n")
	}
	result := strings.TrimSpace(strings.ReplaceAll(b.String(), "\t", " "))
	before := 0
	for len(result) != before {
		before = len(result)
		result = strings.ReplaceAll(result, "  ", " ")
	}
	return strings.TrimSpace(result)
}

// progress returns the steps attempted during this migration
func progress(ch <-chan darwin.MigrationInfo) string {
	var b strings.Builder

	for info := range ch {
		_, _ = fmt.Fprintf(&b, "v%.2f: \"%s\" (%s) Error: %v\n",
			info.Migration.Version, info.Migration.Description, info.Status.String(), info.Error)
	}
	return b.String()
}

// Schema returns the current sqlite definitions as a string for display (without comments)
func Schema() string {
	var b strings.Builder

	schema := defineMigrations()
	for _, m := range schema {
		_, _ = fmt.Fprintf(&b, "-- %s (%.2f)\n%s\n\n", m.Description, m.Version, m.Script)
	}
	return b.String()
}

// VerifyApplicationID checks that the database has the correct application_id.
// Returns ErrInvalidDatabase if the database belongs to a different application.
// Returns nil for empty databases (application_id = 0, no tables) or regserver databases.
func VerifyApplicationID(db *sql.DB) error {
	var appID int
	if err := db.QueryRow("PRAGMA application_id;").Scan(&appID); err != nil {
		return fmt.Errorf("read application_id: %w", err)
	}

	// Accept our application ID
	if appID == ApplicationID {
		return nil
	}

	// Reject non-zero application IDs that aren't ours
	if appID != 0 {
		return fmt.Errorf("%w (application_id 0x%X)", ErrInvalidDatabase, appID)
	}

	// appID is 0 - only accept if database is empty (no user tables)
	var tableCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`).Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("check tables: %w", err)
	}
	if tableCount > 0 {
		return fmt.Errorf("%w (has tables but no application_id)", ErrInvalidDatabase)
	}

	return nil
}

// RunMigrations applies all migrations to an already-open *sql.DB.
// This is perfect for tests using :memory: SQLite.
func RunMigrations(db *sql.DB) error {
	// Verify this is a regserver database (or new) before migrating
	if err := VerifyApplicationID(db); err != nil {
		return err
	}

	count, v1, err := currentVersion(db)
	if err != nil {
		return err
	}

	migrations := minifiedMigrations()
	if count == len(migrations) && v1 == migrations[count-1].Version {
		log.Printf("Database version %.2f is current, no migrations needed", v1)
		return nil // already up to date
	}

	// setup for the migrations
	driver := darwin.NewGenericDriver(db, darwin.SqliteDialect{})
	infoChan := make(chan darwin.MigrationInfo, len(migrations))
	d := darwin.New(driver, migrations, infoChan)

	// perform the migrations
	var v2 float64
	if err := d.Migrate(); err != nil {
		close(infoChan)
		_, v2, _ = currentVersion(db)
		prog := progress(infoChan)
		log.Printf("migration (was v%.2f now v%.2f): %v (%s)", v1, v2, err, prog)
		return fmt.Errorf("migration error: %w\n%s", err, prog)
	}
	close(infoChan)

	_, v2, err = currentVersion(db)
	if err != nil {
		return err
	}

	log.Print(changes(v1, v2))
	return nil
}
