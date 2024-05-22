package metakit

import (
	"gorm.io/gorm"
	"math"
	"net/http"
)

type Metadata struct {
	// Page represents current page
	Page int `form:"page" json:"page"`

	// PageSize is capacity of per page items
	PageSize int `form:"page_size" json:"page_size"`

	// Sort is string type which defines the sort type of data
	Sort string `form:"sort" json:"sort"`

	// SortDirection defines sorted column name
	SortDirection string `form:"sort_direction" json:"sort_direction"`

	// TotalRows defines the quantity of total rows
	TotalRows int64 `json:"total_rows"`

	// TotalPages defines the quantity of total pages, it's defined based on page size and total rows
	TotalPages int64 `json:"total_pages"`
}

// SortDirectionParams function check SortDirection parameter, if it's empty, then it sets ascending order by default
func (m *Metadata) SortDirectionParams() {
	if m.SortDirection == "" {
		m.SortDirection = "asc"
	}
}

// SortParams function take string parameter of sort and set of Sort value
func (m *Metadata) SortParams(sort string) {
	m.Sort = sort
}

// SetPage function sets Page value as a 1 by default, if its equals to 0
func (m *Metadata) setPage() {
	if m.Page == 0 {
		m.Page = 1
	}
}

// SetPageSize function handle PageSize, first it's set default value 10. If page size is greater than 100, then it sets 100
func (m *Metadata) setPageSize() {
	switch {
	case m.PageSize > 100:
		m.PageSize = 100
	case m.PageSize <= 0:
		m.PageSize = 10
	}
}

// Paginate is GORM scope function. Paginate calculates the total pages and offset based on current metadata and applies pagination to the Gorm query
// Paginate function cares Page and PageSize automatically, you can use your own function to replace it, it just overwrite fields
func Paginate(m *Metadata) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		m.setPage()
		m.setPageSize()

		// Calculate total pages based on total rows and page size
		totalPages := int(math.Ceil(float64(m.TotalRows) / float64(m.PageSize)))
		m.TotalPages = int64(totalPages)

		// Calculate offset for the current page
		offset := (m.Page - 1) * m.PageSize

		// Apply offset and limit to the Gorm query
		return db.Offset(offset).Limit(m.PageSize)
	}
}

// GetFilterableFields function iterates through query parameters and removes those that are not needed
//
// Parameters:
//
//	-Pointer of http.Request
//	-Second parameter is string type to delete field from query
//
// Returns:
//
//	-map[string][]strings
func GetFilterableFields(r *http.Request, q string) map[string][]string {
	// Parse the URL query parameters into a map
	query := r.URL.Query()

	// Loop through all query parameters
	for field := range query {
		// Remove the current field from the query
		query.Del(field)
	}

	// Remove the q parameter specifically
	query.Del(q)

	// Return the modified query parameters as a map
	return query
}
