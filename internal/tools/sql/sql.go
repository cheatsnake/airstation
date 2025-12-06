package sql

import (
	"database/sql"
	"fmt"
	"strings"
)

func BuildInClause(column string, num int) string {
	if num == 0 {
		return fmt.Sprintf("%s IN ()", column) // Edge case: Empty IN clause (should be avoided in real queries)
	}

	placeholders := strings.Repeat("?,", num)
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
	return fmt.Sprintf("%s IN (%s)", column, placeholders)
}

func ColumnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	query := `
        SELECT COUNT(*) > 0
        FROM pragma_table_info(?)
        WHERE name = ?
    `

	var exists bool
	err := db.QueryRow(query, tableName, columnName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check column existence: %w", err)
	}

	return exists, nil
}

func TableExists(db *sql.DB, tableName string) (bool, error) {
	query := `
        SELECT COUNT(*) > 0
        FROM sqlite_master
        WHERE type = 'table' AND name = ?
    `

	var exists bool
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return exists, nil
}
