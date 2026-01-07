package sqlite

import (
	"errors"

	"github.com/mattn/go-sqlite3"
)

// IsUniqueConstraintError checks if the error is a SQLite UNIQUE or PRIMARY KEY constraint violation.
func IsUniqueConstraintError(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique ||
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
	}
	return false
}
