package metakit

import (
	"encoding/base64"
	"fmt"

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

		// Apply cursor-based pagination if enabled
		if m.IsCursorBased() {
			return applyCursorPagination(db, m)
		}

		// Apply offset-based pagination
		return db.Offset(m.GetOffset()).Limit(m.GetLimit())
	}
}

// Paginate is a helper function that handles pagination for a GORM query
// It returns the paginated results and updates the metadata with total count
func Paginate(db *gorm.DB, m *Metadata, result interface{}) error {
	// Validate metadata
	validation := m.Validate()
	if !validation.IsValid {
		return fmt.Errorf("invalid metadata: %v", validation.Errors)
	}

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
	// Validate metadata
	validation := m.Validate()
	if !validation.IsValid {
		return fmt.Errorf("invalid metadata: %v", validation.Errors)
	}

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

// applyCursorPagination applies cursor-based pagination to the query
func applyCursorPagination(db *gorm.DB, m *Metadata) *gorm.DB {
	if m.Cursor == "" {
		// First page
		return db.Limit(m.GetLimit())
	}

	// Decode cursor
	cursorValue, err := decodeCursor(m.Cursor)
	if err != nil {
		return db
	}

	// Apply cursor condition
	operator := ">"
	if m.CursorOrder == "desc" {
		operator = "<"
	}

	condition := fmt.Sprintf("%s %s ?", m.CursorField, operator)
	return db.Where(condition, cursorValue).Limit(m.GetLimit())
}

// encodeCursor encodes a value into a cursor string
func encodeCursor(value interface{}) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", value)))
}

// decodeCursor decodes a cursor string back to its original value
func decodeCursor(cursor string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
