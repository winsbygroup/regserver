package backup

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db     *sqlx.DB
	dbPath string
}

func NewService(db *sqlx.DB, dbPath string) *Service {
	return &Service{
		db:     db,
		dbPath: dbPath,
	}
}

// BackupResult contains information about a completed backup
type BackupResult struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
}

// CreateBackup creates a SQL dump of the database
func (s *Service) CreateBackup(ctx context.Context) (*BackupResult, error) {
	// Determine backup directory (relative to DB file)
	dbDir := filepath.Dir(s.dbPath)
	backupDir := filepath.Join(dbDir, "backups")

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	// Generate timestamped filename
	timestamp := time.Now().Format("2006-01-02_15.04.05")
	filename := timestamp + "_regdump.sql.gz"
	backupPath := filepath.Join(backupDir, filename)

	// Create temp file for VACUUM INTO
	tempPath := filepath.Join(backupDir, "temp_backup.db")
	defer os.Remove(tempPath)

	// VACUUM INTO creates a clean, consolidated copy
	if _, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, tempPath); err != nil {
		return nil, fmt.Errorf("vacuum into temp: %w", err)
	}

	// Open the temp database for reading
	tempDB, err := sqlx.Open("sqlite3", tempPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open temp db: %w", err)
	}
	defer tempDB.Close()

	// Generate SQL dump
	dump, err := generateDump(ctx, tempDB)
	if err != nil {
		return nil, fmt.Errorf("generate dump: %w", err)
	}

	// Write gzip-compressed dump to file
	file, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("create backup file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	if _, err := gzWriter.Write([]byte(dump)); err != nil {
		return nil, fmt.Errorf("write gzip data: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("close gzip writer: %w", err)
	}

	// Get file size
	info, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("stat backup file: %w", err)
	}

	return &BackupResult{
		Filename: filename,
		Path:     backupPath,
		Size:     info.Size(),
	}, nil
}

// generateDump creates a SQL dump from the database
func generateDump(ctx context.Context, db *sqlx.DB) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("-- RegServer Database Backup\n")
	sb.WriteString(fmt.Sprintf("-- Generated: %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("PRAGMA foreign_keys=OFF;\n")
	sb.WriteString("BEGIN TRANSACTION;\n\n")

	// Get all schema objects (tables, indexes, triggers, views)
	schemas, err := getSchemas(ctx, db)
	if err != nil {
		return "", err
	}

	// Write schema definitions
	for _, schema := range schemas {
		sb.WriteString(schema.SQL)
		sb.WriteString(";\n")
	}
	sb.WriteString("\n")

	// Get all user tables
	tables, err := getUserTables(ctx, db)
	if err != nil {
		return "", err
	}

	// Generate INSERT statements for each table
	for _, table := range tables {
		inserts, err := generateInserts(ctx, db, table)
		if err != nil {
			return "", fmt.Errorf("generate inserts for %s: %w", table, err)
		}
		if inserts != "" {
			sb.WriteString(inserts)
			sb.WriteString("\n")
		}
	}

	// Footer
	sb.WriteString("COMMIT;\n")
	sb.WriteString("PRAGMA journal_mode=WAL;\n")

	return sb.String(), nil
}

type schemaObject struct {
	Type string `db:"type"`
	Name string `db:"name"`
	SQL  string `db:"sql"`
}

func getSchemas(ctx context.Context, db *sqlx.DB) ([]schemaObject, error) {
	var schemas []schemaObject
	query := `
		SELECT type, name, sql
		FROM sqlite_master
		WHERE sql IS NOT NULL
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY
			CASE type
				WHEN 'table' THEN 1
				WHEN 'index' THEN 2
				WHEN 'trigger' THEN 3
				WHEN 'view' THEN 4
			END,
			name
	`
	if err := db.SelectContext(ctx, &schemas, query); err != nil {
		return nil, fmt.Errorf("query schemas: %w", err)
	}
	return schemas, nil
}

func getUserTables(ctx context.Context, db *sqlx.DB) ([]string, error) {
	var tables []string
	query := `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`
	if err := db.SelectContext(ctx, &tables, query); err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	return tables, nil
}

func generateInserts(ctx context.Context, db *sqlx.DB, table string) (string, error) {
	// Get column names
	rows, err := db.QueryxContext(ctx, fmt.Sprintf("SELECT * FROM %q LIMIT 0", table))
	if err != nil {
		return "", fmt.Errorf("query columns: %w", err)
	}
	columns, err := rows.Columns()
	rows.Close()
	if err != nil {
		return "", fmt.Errorf("get columns: %w", err)
	}

	// Query all rows
	rows, err = db.QueryxContext(ctx, fmt.Sprintf("SELECT * FROM %q", table))
	if err != nil {
		return "", fmt.Errorf("query rows: %w", err)
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		row, err := rows.SliceScan()
		if err != nil {
			return "", fmt.Errorf("scan row: %w", err)
		}

		values := make([]string, len(row))
		for i, v := range row {
			values[i] = formatValue(v)
		}

		sb.WriteString(fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s);\n",
			table,
			strings.Join(quoteColumns(columns), ", "),
			strings.Join(values, ", ")))
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate rows: %w", err)
	}

	return sb.String(), nil
}

func quoteColumns(columns []string) []string {
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = fmt.Sprintf("%q", col)
	}
	return quoted
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case []byte:
		// Could be string or blob
		s := string(val)
		return fmt.Sprintf("'%s'", escapeString(s))
	case string:
		return fmt.Sprintf("'%s'", escapeString(val))
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		// Fallback: treat as string
		return fmt.Sprintf("'%s'", escapeString(fmt.Sprintf("%v", val)))
	}
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
