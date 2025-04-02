package metakit

import (
	"gorm.io/gorm"
)

// GPaginate is a GORM scope function that applies pagination and sorting to a query
func GPaginate(m *Metadata) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Validate and set defaults
		m.ValidateAndSetDefaults()

		// Apply sorting if specified
		if m.Sort != "" {
			db = db.Order(m.GetSortClause())
		}

		// Apply pagination
		return db.Offset(m.GetOffset()).Limit(m.GetLimit())
	}
}

// Paginate is a helper function that handles pagination for a GORM query
// It returns the paginated results and updates the metadata with total count
func Paginate(db *gorm.DB, m *Metadata, result interface{}) error {
	// Get total count before applying pagination
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return err
	}
	m.TotalRows = total

	// Apply pagination and get results
	if err := db.Scopes(GPaginate(m)).Find(result).Error; err != nil {
		return err
	}

	// Update metadata with calculated values
	m.ValidateAndSetDefaults()
	return nil
}

// PaginateWithCount is similar to Paginate but allows you to specify a custom count query
// Useful when you need to count with specific conditions
func PaginateWithCount(db *gorm.DB, countQuery *gorm.DB, m *Metadata, result interface{}) error {
	// Get total count using the custom count query
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return err
	}
	m.TotalRows = total

	// Apply pagination and get results
	if err := db.Scopes(GPaginate(m)).Find(result).Error; err != nil {
		return err
	}

	// Update metadata with calculated values
	m.ValidateAndSetDefaults()
	return nil
}
