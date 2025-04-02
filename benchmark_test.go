package metakit

import (
	"testing"
	"time"
)

// TestUser is a test model for benchmarking
type TestUser struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Name      string
	Email     string
}

// BenchmarkOffsetPagination benchmarks offset-based pagination
func BenchmarkOffsetPagination(b *testing.B) {
	db := setupTestDB(b)

	// Add benchmark-specific data
	for i := 0; i < 100000; i++ {
		user := User{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   30,
		}
		if err := db.Create(&user).Error; err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []User
		metadata := NewMetadata().
			WithPage(1).
			WithPageSize(10).
			WithSort("id").
			WithSortDirection("desc")

		if err := Paginate(db.Model(&User{}), metadata, &users); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCursorPagination benchmarks cursor-based pagination
func BenchmarkCursorPagination(b *testing.B) {
	db := setupTestDB(b)

	// Add benchmark-specific data
	for i := 0; i < 100000; i++ {
		user := User{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   30,
		}
		if err := db.Create(&user).Error; err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []User
		metadata := NewMetadata().
			WithCursorField("id").
			WithCursorOrder("desc").
			WithPageSize(10)

		if err := Paginate(db.Model(&User{}), metadata, &users); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkOffsetPaginationWithCount benchmarks offset-based pagination with count
func BenchmarkOffsetPaginationWithCount(b *testing.B) {
	db := setupTestDB(b)

	// Add benchmark-specific data
	for i := 0; i < 100000; i++ {
		user := User{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   30,
		}
		if err := db.Create(&user).Error; err != nil {
			b.Fatal(err)
		}
	}

	countQuery := db.Model(&User{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []User
		metadata := NewMetadata().
			WithPage(1).
			WithPageSize(10).
			WithSort("id").
			WithSortDirection("desc")

		if err := PaginateWithCount(db.Model(&User{}), countQuery, metadata, &users); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCursorPaginationWithCount benchmarks cursor-based pagination with count
func BenchmarkCursorPaginationWithCount(b *testing.B) {
	db := setupTestDB(b)

	// Add benchmark-specific data
	for i := 0; i < 100000; i++ {
		user := User{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   30,
		}
		if err := db.Create(&user).Error; err != nil {
			b.Fatal(err)
		}
	}

	countQuery := db.Model(&User{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []User
		metadata := NewMetadata().
			WithCursorField("id").
			WithCursorOrder("desc").
			WithPageSize(10)

		if err := PaginateWithCount(db.Model(&User{}), countQuery, metadata, &users); err != nil {
			b.Fatal(err)
		}
	}
}
