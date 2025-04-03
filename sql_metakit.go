package metakit

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

// QueryOptimizer provides optimization strategies for queries
type QueryOptimizer struct {
	UseIndexHint    bool
	UseQueryCache   bool
	BatchSize       int
	Timeout         time.Duration
	MaxRows         int
	UseMaterialized bool
}

// NewQueryOptimizer creates a new query optimizer with default settings
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		UseIndexHint:    true,
		UseQueryCache:   true,
		BatchSize:       1000,
		Timeout:         30 * time.Second,
		MaxRows:         10000,
		UseMaterialized: false,
	}
}

// WithIndexHint enables or disables index hints
func (q *QueryOptimizer) WithIndexHint(use bool) *QueryOptimizer {
	q.UseIndexHint = use
	return q
}

// WithQueryCache enables or disables query caching
func (q *QueryOptimizer) WithQueryCache(use bool) *QueryOptimizer {
	q.UseQueryCache = use
	return q
}

// WithBatchSize sets the batch size for batch operations
func (q *QueryOptimizer) WithBatchSize(size int) *QueryOptimizer {
	q.BatchSize = size
	return q
}

// WithTimeout sets the query timeout
func (q *QueryOptimizer) WithTimeout(timeout time.Duration) *QueryOptimizer {
	q.Timeout = timeout
	return q
}

// WithMaxRows sets the maximum number of rows to return
func (q *QueryOptimizer) WithMaxRows(max int) *QueryOptimizer {
	q.MaxRows = max
	return q
}

// WithMaterialized enables or disables materialized views
func (q *QueryOptimizer) WithMaterialized(use bool) *QueryOptimizer {
	q.UseMaterialized = use
	return q
}

// OptimizeQuery applies optimization strategies to the query
func (q *QueryOptimizer) OptimizeQuery(query string, dialect Dialect) string {
	optimized := query

	// Add index hints if enabled
	if q.UseIndexHint {
		switch dialect {
		case MySQL:
			optimized = addMySQLIndexHints(optimized)
		case PostgreSQL:
			optimized = addPostgreSQLIndexHints(optimized)
		}
	}

	// Add materialized view if enabled
	if q.UseMaterialized {
		optimized = addMaterializedView(optimized, dialect)
	}

	// Add row limit if MaxRows is set
	if q.MaxRows > 0 {
		optimized = addRowLimit(optimized, q.MaxRows, dialect)
	}

	return optimized
}

// addMySQLIndexHints adds MySQL-specific index hints
func addMySQLIndexHints(query string) string {
	// Add FORCE INDEX hint for better performance
	if strings.Contains(strings.ToLower(query), "where") {
		return strings.Replace(query, "WHERE", "FORCE INDEX (idx_created_at) WHERE", 1)
	}
	return query
}

// addPostgreSQLIndexHints adds PostgreSQL-specific index hints
func addPostgreSQLIndexHints(query string) string {
	// Add index hints using PostgreSQL syntax
	if strings.Contains(strings.ToLower(query), "where") {
		return strings.Replace(query, "WHERE", "WHERE /*+ IndexScan(table_name idx_created_at) */", 1)
	}
	return query
}

// addMaterializedView adds materialized view support
func addMaterializedView(query string, dialect Dialect) string {
	switch dialect {
	case PostgreSQL:
		return "WITH MATERIALIZED " + query
	case MySQL:
		return "WITH RECURSIVE " + query
	default:
		return query
	}
}

// addRowLimit adds a row limit to the query
func addRowLimit(query string, limit int, dialect Dialect) string {
	switch dialect {
	case PostgreSQL:
		return query + fmt.Sprintf(" LIMIT %d", limit)
	case MySQL, SQLite:
		return query + fmt.Sprintf(" LIMIT %d", limit)
	default:
		return query
	}
}
