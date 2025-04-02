package metakit

// ValidationError represents a single validation error with field-specific information.
// It provides both human-readable messages and machine-readable error codes.
type ValidationError struct {
	Field   string // The field that failed validation
	Message string // Human-readable error message
	Code    string // Machine-readable error code
}

// ValidationResult represents the result of metadata validation.
// It contains both the overall validation status and a list of specific errors.
type ValidationResult struct {
	IsValid bool              // Whether the validation passed
	Errors  []ValidationError // List of validation errors if any
}

// Metadata represents pagination and sorting metadata for database queries.
// It supports both offset-based and cursor-based pagination.
//
// Example (offset-based):
//
//	metadata := NewMetadata().
//	  WithPage(1).
//	  WithPageSize(10).
//	  WithSort("created_at").
//	  WithSortDirection("desc")
//
// Example (cursor-based):
//
//	metadata := NewMetadata().
//	  WithCursor("eyJpZCI6MTIzLCJjcmVhdGVkX2F0IjoiMjAyNC0wMy0yMFQxMjowMDowMFoiLCJwYWdlIjoxfQ==").
//	  WithCursorField("created_at").
//	  WithCursorOrder("desc")
type Metadata struct {
	// Page represents current page (1-based)
	Page int `form:"page" json:"page"`

	// PageSize is capacity of per page items
	PageSize int `form:"page_size" json:"page_size"`

	// Sort is string type which defines the sort field
	Sort string `form:"sort" json:"sort"`

	// SortDirection defines sort direction (asc/desc)
	SortDirection string `form:"sort_direction" json:"sort_direction"`

	// TotalRows defines the quantity of total rows
	TotalRows int64 `json:"total_rows"`

	// TotalPages defines the quantity of total pages
	TotalPages int64 `json:"total_pages"`

	// HasNext indicates if there is a next page
	HasNext bool `json:"has_next"`

	// HasPrevious indicates if there is a previous page
	HasPrevious bool `json:"has_previous"`

	// FromRow indicates the starting row number of the current page
	FromRow int64 `json:"from_row"`

	// ToRow indicates the ending row number of the current page
	ToRow int64 `json:"to_row"`

	// Cursor-based pagination fields
	Cursor      string `form:"cursor" json:"cursor"`
	CursorField string `form:"cursor_field" json:"cursor_field"`
	CursorOrder string `form:"cursor_order" json:"cursor_order"`
}

// NewMetadata creates a new Metadata instance with default values.
// Defaults:
//   - Page: 1
//   - PageSize: 10
//   - SortDirection: "asc"
//
// Example:
//
//	metadata := NewMetadata()
//	// metadata.Page == 1
//	// metadata.PageSize == 10
//	// metadata.SortDirection == "asc"
func NewMetadata() *Metadata {
	return &Metadata{
		Page:          1,
		PageSize:      10,
		SortDirection: "asc",
	}
}

// WithPage sets the page number and returns the metadata for method chaining.
// Page numbers are 1-based.
//
// Example:
//
//	metadata := NewMetadata().WithPage(2)
//	// metadata.Page == 2
func (m *Metadata) WithPage(page int) *Metadata {
	m.Page = page
	return m
}

// WithPageSize sets the page size and returns the metadata for method chaining.
// Page size must be between 1 and 100.
//
// Example:
//
//	metadata := NewMetadata().WithPageSize(20)
//	// metadata.PageSize == 20
func (m *Metadata) WithPageSize(pageSize int) *Metadata {
	m.PageSize = pageSize
	return m
}

// WithSort sets the sort field and returns the metadata for method chaining.
// The sort field should match a column name in your database.
//
// Example:
//
//	metadata := NewMetadata().WithSort("created_at")
//	// metadata.Sort == "created_at"
func (m *Metadata) WithSort(sort string) *Metadata {
	m.Sort = sort
	return m
}

// WithSortDirection sets the sort direction and returns the metadata for method chaining.
// Valid values are "asc" or "desc".
//
// Example:
//
//	metadata := NewMetadata().WithSortDirection("desc")
//	// metadata.SortDirection == "desc"
func (m *Metadata) WithSortDirection(direction string) *Metadata {
	m.SortDirection = direction
	return m
}

// ValidateAndSetDefaults validates and sets default values for the metadata.
// This method should be called before using the metadata for pagination.
//
// Defaults:
//   - Page: 1 (if < 1)
//   - PageSize: 10 (if < 1) or 100 (if > 100)
//   - SortDirection: "asc" (if empty or invalid)
//
// Example:
//
//	metadata := NewMetadata()
//	metadata.Page = 0
//	metadata.PageSize = 150
//	metadata.ValidateAndSetDefaults()
//	// metadata.Page == 1
//	// metadata.PageSize == 100
func (m *Metadata) ValidateAndSetDefaults() {
	// Set default page
	if m.Page < 1 {
		m.Page = 1
	}

	// Set default page size
	if m.PageSize < 1 {
		m.PageSize = 10
	} else if m.PageSize > 100 {
		m.PageSize = 100
	}

	// Set default sort direction
	if m.SortDirection == "" {
		m.SortDirection = "asc"
	} else if m.SortDirection != "asc" && m.SortDirection != "desc" {
		m.SortDirection = "asc"
	}

	// Calculate pagination metadata
	if m.TotalRows > 0 {
		m.TotalPages = (m.TotalRows + int64(m.PageSize) - 1) / int64(m.PageSize)
		m.HasNext = m.Page < int(m.TotalPages)
		m.HasPrevious = m.Page > 1
		m.FromRow = int64((m.Page-1)*m.PageSize + 1)
		m.ToRow = int64(m.Page * m.PageSize)
		if m.ToRow > m.TotalRows {
			m.ToRow = m.TotalRows
		}
	}
}

// GetOffset returns the offset for the current page.
// This is calculated as (page - 1) * pageSize.
//
// Example:
//
//	metadata := NewMetadata().WithPage(2).WithPageSize(10)
//	offset := metadata.GetOffset()
//	// offset == 10
func (m *Metadata) GetOffset() int {
	return (m.Page - 1) * m.PageSize
}

// GetLimit returns the limit for the current page.
// This is the same as the page size.
//
// Example:
//
//	metadata := NewMetadata().WithPageSize(20)
//	limit := metadata.GetLimit()
//	// limit == 20
func (m *Metadata) GetLimit() int {
	return m.PageSize
}

// GetSortClause returns the sort clause for the current sort settings.
// Returns an empty string if no sort field is specified.
//
// Example:
//
//	metadata := NewMetadata().WithSort("created_at").WithSortDirection("desc")
//	sortClause := metadata.GetSortClause()
//	// sortClause == "created_at desc"
func (m *Metadata) GetSortClause() string {
	if m.Sort == "" {
		return ""
	}
	return m.Sort + " " + m.SortDirection
}

// Validate performs validation on the metadata and returns a ValidationResult.
// This method checks:
//   - Page is greater than 0
//   - PageSize is between 1 and 100
//   - SortDirection is either "asc" or "desc"
//   - CursorField is provided when using cursor-based pagination
//   - CursorOrder is either "asc" or "desc" when provided
//
// Example:
//
//	metadata := NewMetadata()
//	metadata.Page = 0
//	result := metadata.Validate()
//	// result.IsValid == false
//	// result.Errors[0].Field == "page"
//	// result.Errors[0].Message == "Page must be greater than 0"
func (m *Metadata) Validate() ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  make([]ValidationError, 0),
	}

	// Validate page
	if m.Page < 1 {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "page",
			Message: "Page must be greater than 0",
			Code:    "INVALID_PAGE",
		})
	}

	// Validate page size
	if m.PageSize < 1 {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "page_size",
			Message: "Page size must be greater than 0",
			Code:    "INVALID_PAGE_SIZE",
		})
	} else if m.PageSize > 100 {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "page_size",
			Message: "Page size cannot exceed 100",
			Code:    "PAGE_SIZE_TOO_LARGE",
		})
	}

	// Validate sort direction
	if m.SortDirection != "" && m.SortDirection != "asc" && m.SortDirection != "desc" {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "sort_direction",
			Message: "Sort direction must be either 'asc' or 'desc'",
			Code:    "INVALID_SORT_DIRECTION",
		})
	}

	// Validate cursor pagination fields if using cursor-based pagination
	if m.Cursor != "" {
		if m.CursorField == "" {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "cursor_field",
				Message: "Cursor field is required when using cursor-based pagination",
				Code:    "MISSING_CURSOR_FIELD",
			})
		}
		if m.CursorOrder != "" && m.CursorOrder != "asc" && m.CursorOrder != "desc" {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "cursor_order",
				Message: "Cursor order must be either 'asc' or 'desc'",
				Code:    "INVALID_CURSOR_ORDER",
			})
		}
	}

	return result
}

// WithCursor sets the cursor for cursor-based pagination and returns the metadata for method chaining.
// The cursor should be a base64-encoded string containing the last item's data.
//
// Example:
//
//	metadata := NewMetadata().WithCursor("eyJpZCI6MTIzLCJjcmVhdGVkX2F0IjoiMjAyNC0wMy0yMFQxMjowMDowMFoiLCJwYWdlIjoxfQ==")
//	// metadata.Cursor == "eyJpZCI6MTIzLCJjcmVhdGVkX2F0IjoiMjAyNC0wMy0yMFQxMjowMDowMFoiLCJwYWdlIjoxfQ=="
func (m *Metadata) WithCursor(cursor string) *Metadata {
	m.Cursor = cursor
	return m
}

// WithCursorField sets the field to use for cursor-based pagination and returns the metadata for method chaining.
// This field should match a column name in your database.
//
// Example:
//
//	metadata := NewMetadata().WithCursorField("created_at")
//	// metadata.CursorField == "created_at"
func (m *Metadata) WithCursorField(field string) *Metadata {
	m.CursorField = field
	return m
}

// WithCursorOrder sets the order for cursor-based pagination and returns the metadata for method chaining.
// Valid values are "asc" or "desc".
//
// Example:
//
//	metadata := NewMetadata().WithCursorOrder("desc")
//	// metadata.CursorOrder == "desc"
func (m *Metadata) WithCursorOrder(order string) *Metadata {
	m.CursorOrder = order
	return m
}

// IsCursorBased returns true if cursor-based pagination is being used.
// This is determined by checking if either Cursor or CursorField is set.
//
// Example:
//
//	metadata := NewMetadata()
//	// metadata.IsCursorBased() == false
//
//	metadata.WithCursorField("created_at")
//	// metadata.IsCursorBased() == true
func (m *Metadata) IsCursorBased() bool {
	return m.Cursor != "" || m.CursorField != ""
}

// Complete pagination examples:
//
// Example 1: Offset-based pagination with GORM
//   metadata := NewMetadata().
//     WithPage(1).
//     WithPageSize(10).
//     WithSort("created_at").
//     WithSortDirection("desc")
//
//   var users []User
//   result := db.Model(&User{}).
//     Order(metadata.GetSortClause()).
//     Offset(metadata.GetOffset()).
//     Limit(metadata.GetLimit()).
//     Find(&users)
//
//   // Get total count
//   var total int64
//   db.Model(&User{}).Count(&total)
//   metadata.TotalRows = total
//   metadata.ValidateAndSetDefaults()
//
// Example 2: Cursor-based pagination with GORM
//   metadata := NewMetadata().
//     WithCursorField("created_at").
//     WithCursorOrder("desc").
//     WithPageSize(10)
//
//   var users []User
//   query := db.Model(&User{})
//
//   if metadata.Cursor != "" {
//     cursorValue, _ := decodeCursor(metadata.Cursor)
//     query = query.Where("created_at < ?", cursorValue)
//   }
//
//   result := query.
//     Order(metadata.GetSortClause()).
//     Limit(metadata.GetLimit() + 1). // Get one extra to check if there are more
//     Find(&users)
//
//   // Check if there are more results
//   hasMore := len(users) > metadata.GetLimit()
//   if hasMore {
//     users = users[:metadata.GetLimit()]
//   }
//
//   // Set next cursor if there are more results
//   if hasMore {
//     lastUser := users[len(users)-1]
//     nextCursor := encodeCursor(map[string]interface{}{
//       "created_at": lastUser.CreatedAt,
//       "id": lastUser.ID,
//     })
//     metadata.Cursor = nextCursor
//   }
