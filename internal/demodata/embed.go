// Package demodata provides sample data for demo deployments.
package demodata

import (
	"database/sql"
	"embed"
)

//go:embed sample.sql
var sampleSQL embed.FS

// Load inserts demo data into the database.
// This should only be called on a freshly created database after migrations.
func Load(db *sql.DB) error {
	data, err := sampleSQL.ReadFile("sample.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(data))
	return err
}
