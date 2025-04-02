package metakit

import (
	"context"
	"database/sql"
	"fmt"
)

type Dialect int

const (
	MySQL Dialect = iota
	PostgreSQL
	SQLite
)

// QueryContextPaginate calculates the total pages and offset based on the current metadata and applies pagination to the SQL query
func QueryContextPaginate(ctx context.Context, db *sql.DB, dialect Dialect, query string, m *Metadata, args ...any) (*sql.Rows, error) {
	// calculate the total pages
	if m.PageSize > 0 {
		totalPages := (m.TotalRows + int64(m.PageSize) - 1) / int64(m.PageSize)
		m.TotalPages = totalPages
	} else {
		m.TotalPages = 1
	}

	// calculate offset for the current page
	offset := (m.Page - 1) * m.PageSize

	var sortDirection string

	// build the paginated query
	var paginatedQuery string
	switch dialect {
	case PostgreSQL:
		// Use $1, $2 for parameterized queries
		paginatedQuery = fmt.Sprintf("%s ORDER BY %s %s LIMIT $%d OFFSET $%d", query, m.Sort, sortDirection, len(args)+1, len(args)+2)
		args = append(args, m.PageSize, offset)
	case MySQL, SQLite:
		// Use ? for parameterized queries
		paginatedQuery = fmt.Sprintf("%s ORDER BY %s %s LIMIT ? OFFSET ?", query, m.Sort, sortDirection)
		args = append(args, m.PageSize, offset)
	}

	rows, err := db.QueryContext(ctx, paginatedQuery, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
