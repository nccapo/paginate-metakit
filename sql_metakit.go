package metakit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Dialect int

const (
	MySQL Dialect = iota
	PostgreSQL
	SQLite
)

// QueryContextPaginate calculates the total pages and offset based on the current metadata and applies pagination to the SQL query
func QueryContextPaginate(ctx context.Context, db *sql.DB, dialect Dialect, query string, m *Metadata, args ...any) (*sql.Rows, error) {
	// Validate metadata
	validation := m.Validate()
	if !validation.IsValid {
		return nil, fmt.Errorf("invalid metadata: %v", validation.Errors)
	}

	// Apply cursor-based pagination if enabled
	if m.IsCursorBased() {
		return applyCursorSQLPagination(ctx, db, dialect, query, m, args...)
	}

	// Calculate the total pages
	if m.PageSize > 0 {
		totalPages := (m.TotalRows + int64(m.PageSize) - 1) / int64(m.PageSize)
		m.TotalPages = totalPages
	} else {
		m.TotalPages = 1
	}

	// Calculate offset for the current page
	offset := (m.Page - 1) * m.PageSize

	// Build the paginated query
	var paginatedQuery string
	switch dialect {
	case PostgreSQL:
		// Use $1, $2 for parameterized queries
		paginatedQuery = fmt.Sprintf("%s ORDER BY %s %s LIMIT $%d OFFSET $%d", query, m.Sort, m.SortDirection, len(args)+1, len(args)+2)
		args = append(args, m.PageSize, offset)
	case MySQL, SQLite:
		// Use ? for parameterized queries
		paginatedQuery = fmt.Sprintf("%s ORDER BY %s %s LIMIT ? OFFSET ?", query, m.Sort, m.SortDirection)
		args = append(args, m.PageSize, offset)
	}

	rows, err := db.QueryContext(ctx, paginatedQuery, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// applyCursorSQLPagination applies cursor-based pagination to the SQL query
func applyCursorSQLPagination(ctx context.Context, db *sql.DB, dialect Dialect, query string, m *Metadata, args ...any) (*sql.Rows, error) {
	var paginatedQuery string
	var cursorCondition string

	// Build cursor condition
	if m.Cursor != "" {
		cursorValue, err := decodeCursor(m.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		operator := ">"
		if m.CursorOrder == "desc" {
			operator = "<"
		}

		switch dialect {
		case PostgreSQL:
			cursorCondition = fmt.Sprintf("WHERE %s %s $%d", m.CursorField, operator, len(args)+1)
			args = append(args, cursorValue)
		case MySQL, SQLite:
			cursorCondition = fmt.Sprintf("WHERE %s %s ?", m.CursorField, operator)
			args = append(args, cursorValue)
		}
	}

	// Build the complete query
	switch dialect {
	case PostgreSQL:
		paginatedQuery = fmt.Sprintf("%s %s ORDER BY %s %s LIMIT $%d", query, cursorCondition, m.CursorField, m.CursorOrder, len(args)+1)
		args = append(args, m.PageSize)
	case MySQL, SQLite:
		paginatedQuery = fmt.Sprintf("%s %s ORDER BY %s %s LIMIT ?", query, cursorCondition, m.CursorField, m.CursorOrder)
		args = append(args, m.PageSize)
	}

	rows, err := db.QueryContext(ctx, paginatedQuery, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// New types for cursor pagination
type CursorPage struct {
	Data       []map[string]interface{} `json:"data"`
	NextCursor string                   `json:"next_cursor,omitempty"`
	PrevCursor string                   `json:"prev_cursor,omitempty"`
	HasMore    bool                     `json:"has_more"`
}

// New cache types
type CacheConfig struct {
	Enabled bool
	TTL     time.Duration
	MaxSize int
}

// New optimization options
type QueryOptions struct {
	CacheConfig   CacheConfig
	UseIndexHint  bool
	Timeout       time.Duration
	OptimizeCount bool
}
