package metakit

// Metadata represents pagination and sorting metadata for database queries
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
}

// NewMetadata creates a new Metadata instance with default values
func NewMetadata() *Metadata {
	return &Metadata{
		Page:          1,
		PageSize:      10,
		SortDirection: "asc",
	}
}

// WithPage sets the page number and returns the metadata for method chaining
func (m *Metadata) WithPage(page int) *Metadata {
	m.Page = page
	return m
}

// WithPageSize sets the page size and returns the metadata for method chaining
func (m *Metadata) WithPageSize(pageSize int) *Metadata {
	m.PageSize = pageSize
	return m
}

// WithSort sets the sort field and returns the metadata for method chaining
func (m *Metadata) WithSort(sort string) *Metadata {
	m.Sort = sort
	return m
}

// WithSortDirection sets the sort direction and returns the metadata for method chaining
func (m *Metadata) WithSortDirection(direction string) *Metadata {
	m.SortDirection = direction
	return m
}

// ValidateAndSetDefaults validates and sets default values for the metadata
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

// GetOffset returns the offset for the current page
func (m *Metadata) GetOffset() int {
	return (m.Page - 1) * m.PageSize
}

// GetLimit returns the limit for the current page
func (m *Metadata) GetLimit() int {
	return m.PageSize
}

// GetSortClause returns the sort clause for the current sort settings
func (m *Metadata) GetSortClause() string {
	if m.Sort == "" {
		return ""
	}
	return m.Sort + " " + m.SortDirection
}
