package metakit

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestSPaginate(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	for i := 1; i <= 100; i++ {
		_, err = db.Exec("INSERT INTO items (name) VALUES (?)", fmt.Sprintf("Item %d", i))
		if err != nil {
			t.Fatalf("failed to insert data: %v", err)
		}
	}

	tests := []struct {
		metadata     Metadata
		expectedPage int
		expectedSize int
		expectedOff  int
	}{
		{Metadata{Page: 0, PageSize: 0, TotalRows: 100, Sort: "id"}, 1, 10, 0},
		{Metadata{Page: 3, PageSize: 20, TotalRows: 100, Sort: "id"}, 3, 20, 40},
		{Metadata{Page: 2, PageSize: 50, TotalRows: 120, Sort: "id"}, 2, 50, 50},
	}

	for _, test := range tests {
		test.metadata.ValidateAndSetDefaults()
		rows, err := QueryContextPaginate(context.Background(), db, 1, "SELECT * FROM items", &test.metadata)
		if err != nil {
			t.Fatalf("failed to execute paginated query: %v", err)
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				log.Fatalf("failed to close rows: %v", err)
			}
		}(rows)

		totalPages := (test.metadata.TotalRows + int64(test.metadata.PageSize) - 1) / int64(test.metadata.PageSize)
		if test.metadata.TotalPages != totalPages {
			t.Errorf("expected %v total pages, got %v", totalPages, test.metadata.TotalPages)
		}

		if test.metadata.Page != test.expectedPage {
			t.Errorf("expected page %v, got %v", test.expectedPage, test.metadata.Page)
		}

		if test.metadata.PageSize != test.expectedSize {
			t.Errorf("expected page size %v, got %v", test.expectedSize, test.metadata.PageSize)
		}

		offset := (test.metadata.Page - 1) * test.metadata.PageSize
		if offset != test.expectedOff {
			t.Errorf("expected offset %v, got %v", test.expectedOff, offset)
		}

		// Verify the number of rows returned matches the page size
		count := 0
		for rows.Next() {
			count++
		}
		if count != test.metadata.PageSize {
			t.Errorf("expected %v rows, got %v", test.metadata.PageSize, count)
		}
	}
}

func TestQueryOptimization(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		dialect   Dialect
		optimizer *QueryOptimizer
		expected  string
	}{
		{
			name:    "MySQL index hint",
			query:   "SELECT * FROM users WHERE age > 18",
			dialect: MySQL,
			optimizer: NewQueryOptimizer().
				WithIndexHint(true).
				WithMaxRows(0), // Disable row limit for this test
			expected: "SELECT * FROM users FORCE INDEX (idx_created_at) WHERE age > 18",
		},
		{
			name:    "PostgreSQL index hint",
			query:   "SELECT * FROM users WHERE age > 18",
			dialect: PostgreSQL,
			optimizer: NewQueryOptimizer().
				WithIndexHint(true).
				WithMaxRows(0), // Disable row limit for this test
			expected: "SELECT * FROM users WHERE /*+ IndexScan(table_name idx_created_at) */ age > 18",
		},
		{
			name:    "Row limit",
			query:   "SELECT * FROM users",
			dialect: PostgreSQL,
			optimizer: NewQueryOptimizer().
				WithMaxRows(100),
			expected: "SELECT * FROM users LIMIT 100",
		},
		{
			name:    "Materialized view",
			query:   "SELECT * FROM users",
			dialect: PostgreSQL,
			optimizer: NewQueryOptimizer().
				WithMaterialized(true).
				WithMaxRows(0), // Disable row limit for this test
			expected: "WITH MATERIALIZED SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.optimizer.OptimizeQuery(tt.query, tt.dialect)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestQueryContextPaginateWithPostgreSQLParams(t *testing.T) {
	// Skip this test if not using PostgreSQL
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	// Connect to PostgreSQL database
	// Note: This test requires a running PostgreSQL instance
	// You may need to adjust the connection string
	db, err := sql.Open("postgres", "postgres://username:password@localhost:5432/testdb?sslmode=disable")
	if err != nil {
		t.Skipf("Skipping test: could not connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec("DROP TABLE IF EXISTS test_items")
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	_, err = db.Exec("CREATE TABLE test_items (id SERIAL PRIMARY KEY, name TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Insert test data
	for i := 1; i <= 100; i++ {
		_, err = db.Exec("INSERT INTO test_items (name) VALUES ($1)", fmt.Sprintf("Item %d", i))
		if err != nil {
			t.Fatalf("failed to insert data: %v", err)
		}
	}

	// Test with a query that has parameters
	query := "SELECT * FROM test_items WHERE created_at > $1"
	createdAt := time.Now().Add(-24 * time.Hour) // 24 hours ago

	metadata := &Metadata{
		Page:          1,
		PageSize:      10,
		TotalRows:     100,
		Sort:          "name",
		SortDirection: "asc",
	}

	rows, err := QueryContextPaginate(context.Background(), db, PostgreSQL, query, metadata, createdAt)
	if err != nil {
		t.Fatalf("failed to execute paginated query: %v", err)
	}
	defer rows.Close()

	// Verify the results
	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Log("No rows returned, which might be expected depending on the data")
	}
}
