package sql

import (
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
