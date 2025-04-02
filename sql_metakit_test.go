package metakit

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

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
