package metakit

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func TestSortDirectionParams(t *testing.T) {
	tests := []struct {
		input    Metadata
		expected string
	}{
		{Metadata{SortDirection: ""}, "asc"},
		{Metadata{SortDirection: "desc"}, "desc"},
	}

	for _, test := range tests {
		test.input.ValidateAndSetDefaults()
		if test.input.SortDirection != test.expected {
			t.Errorf("expected %v, got %v", test.expected, test.input.SortDirection)
		}
	}
}

func TestSortParams(t *testing.T) {
	m := Metadata{}
	sort := "name"
	m.WithSort(sort)
	if m.Sort != sort {
		t.Errorf("expected %v, got %v", sort, m.Sort)
	}
}

func TestGPaginate(t *testing.T) {
	tests := []struct {
		metadata     Metadata
		expectedPage int
		expectedSize int
		expectedOff  int
	}{
		{Metadata{Page: 0, PageSize: 0, TotalRows: 100}, 1, 10, 0},
		{Metadata{Page: 3, PageSize: 20, TotalRows: 100}, 3, 20, 40},
		{Metadata{Page: 2, PageSize: 50, TotalRows: 120}, 2, 50, 50},
	}

	for _, test := range tests {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("failed to connect to database: %v", err)
		}

		_ = GPaginate(&test.metadata)(db)

		test.metadata.ValidateAndSetDefaults()

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
	}
}

func TestSQLPaginate(t *testing.T) {
	tests := []struct {
		metadata     Metadata
		expectedPage int
		expectedSize int
		expectedOff  int
	}{
		{Metadata{Page: 0, PageSize: 0, TotalRows: 100}, 1, 10, 0},
		{Metadata{Page: 3, PageSize: 20, TotalRows: 100}, 3, 20, 40},
		{Metadata{Page: 2, PageSize: 50, TotalRows: 120}, 2, 50, 50},
	}

	for _, test := range tests {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("failed to connect to database: %v", err)
		}

		_ = GPaginate(&test.metadata)(db)

		test.metadata.ValidateAndSetDefaults()

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
	}
}

type User struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"size:255"`
	Email string `gorm:"size:255"`
	Age   int
}

// setupTestDB creates a test database with sample data
func setupTestDB(t interface{ Fatal(args ...interface{}) }) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatal(err)
	}

	// Insert test data
	users := []User{
		{Name: "John Doe", Email: "john@example.com", Age: 30},
		{Name: "Jane Smith", Email: "jane@example.com", Age: 25},
		{Name: "Bob Johnson", Email: "bob@example.com", Age: 35},
		{Name: "Alice Brown", Email: "alice@example.com", Age: 28},
		{Name: "Charlie Wilson", Email: "charlie@example.com", Age: 32},
	}

	for _, user := range users {
		err = db.Create(&user).Error
		if err != nil {
			t.Fatal(err)
		}
	}

	return db
}

func TestPagination(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name          string
		metadata      *Metadata
		expectedCount int
		expectedFirst string
		expectedLast  string
		hasNext       bool
		hasPrevious   bool
		expectedFrom  int64
		expectedTo    int64
	}{
		{
			name: "First page with default size",
			metadata: NewMetadata().
				WithPage(1).
				WithPageSize(2).
				WithSort("name").
				WithSortDirection("asc"),
			expectedCount: 2,
			expectedFirst: "Alice Brown",
			expectedLast:  "Bob Johnson",
			hasNext:       true,
			hasPrevious:   false,
			expectedFrom:  1,
			expectedTo:    2,
		},
		{
			name: "Second page with custom size",
			metadata: NewMetadata().
				WithPage(2).
				WithPageSize(3).
				WithSort("name").
				WithSortDirection("asc"),
			expectedCount: 2,
			expectedFirst: "Jane Smith",
			expectedLast:  "John Doe",
			hasNext:       false,
			hasPrevious:   true,
			expectedFrom:  4,
			expectedTo:    5,
		},
		{
			name: "Sort by age descending",
			metadata: NewMetadata().
				WithPage(1).
				WithPageSize(2).
				WithSort("age").
				WithSortDirection("desc"),
			expectedCount: 2,
			expectedFirst: "Bob Johnson",
			expectedLast:  "Charlie Wilson",
			hasNext:       true,
			hasPrevious:   false,
			expectedFrom:  1,
			expectedTo:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var users []User
			err := Paginate(db.Model(&User{}), tt.metadata, &users)
			assert.NoError(t, err)

			// Check results
			assert.Equal(t, tt.expectedCount, len(users))
			assert.Equal(t, tt.expectedFirst, users[0].Name)
			assert.Equal(t, tt.expectedLast, users[len(users)-1].Name)

			// Check metadata
			assert.Equal(t, tt.hasNext, tt.metadata.HasNext)
			assert.Equal(t, tt.hasPrevious, tt.metadata.HasPrevious)
			assert.Equal(t, tt.expectedFrom, tt.metadata.FromRow)
			assert.Equal(t, tt.expectedTo, tt.metadata.ToRow)
		})
	}
}

func TestCustomCountQuery(t *testing.T) {
	db := setupTestDB(t)

	// Create a custom count query that only counts users over 30
	countQuery := db.Model(&User{}).Where("age > ?", 30)

	metadata := NewMetadata().
		WithPage(1).
		WithPageSize(2)

	var users []User
	err := PaginateWithCount(db.Model(&User{}), countQuery, metadata, &users)
	assert.NoError(t, err)

	// Should only get users over 30
	assert.Equal(t, 2, len(users))
	assert.Equal(t, int64(2), metadata.TotalRows)
}

func TestFieldSelection(t *testing.T) {
	db := setupTestDB(t)

	// Test with selected fields
	metadata := NewMetadata().
		WithPage(1).
		WithPageSize(2).
		WithSort("name").
		WithSortDirection("asc").
		WithFields("name", "email") // Only select name and email

	var users []User
	err := Paginate(db.Model(&User{}), metadata, &users)
	assert.NoError(t, err)

	// Check that we got 2 users
	assert.Equal(t, 2, len(users))

	// Check that the ID field is zero (not selected)
	assert.Equal(t, uint(0), users[0].ID)

	// Check that name and email are populated
	assert.NotEmpty(t, users[0].Name)
	assert.NotEmpty(t, users[0].Email)
}

func TestValidationRules(t *testing.T) {
	// Test validation rule for page size (max)
	metadata := NewMetadata().
		WithPage(1).
		WithPageSize(30).
		WithValidationRule("page_size", "max:20")

	validation := metadata.Validate()
	assert.False(t, validation.IsValid)
	assert.Equal(t, 1, len(validation.Errors))
	assert.Equal(t, "page_size", validation.Errors[0].Field)
	assert.Equal(t, "PAGE_SIZE_EXCEEDS_MAX", validation.Errors[0].Code)

	// Test validation rule for sort field (in)
	metadata = NewMetadata().
		WithPage(1).
		WithPageSize(10).
		WithSort("invalid_field").
		WithValidationRule("sort", "in:name,email,age")

	validation = metadata.Validate()
	assert.False(t, validation.IsValid)
	assert.Equal(t, 1, len(validation.Errors))
	assert.Equal(t, "sort", validation.Errors[0].Field)
	assert.Equal(t, "INVALID_SORT_FIELD", validation.Errors[0].Code)

	// Test validation rule for fields (in)
	metadata = NewMetadata().
		WithPage(1).
		WithPageSize(10).
		WithFields("id", "invalid_field").
		WithValidationRule("fields", "in:id,name,email,age")

	validation = metadata.Validate()
	assert.False(t, validation.IsValid)
	assert.Equal(t, 1, len(validation.Errors))
	assert.Equal(t, "fields", validation.Errors[0].Field)
	assert.Equal(t, "INVALID_SELECTED_FIELD", validation.Errors[0].Code)
}

func TestDebugMode(t *testing.T) {
	db := setupTestDB(t)

	// Test with debug mode enabled (we can't easily test the output, but we can test it doesn't break)
	metadata := NewMetadata().
		WithPage(1).
		WithPageSize(2).
		WithSort("name").
		WithSortDirection("asc").
		WithDebug(true)

	var users []User
	err := Paginate(db.Model(&User{}), metadata, &users)
	assert.NoError(t, err)

	// Check that we still get results with debug enabled
	assert.Equal(t, 2, len(users))
}
