package metakit

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
)

// GPaginate is a GORM scope function that applies pagination and sorting to a query
func GPaginate(m *Metadata) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Validate and set defaults
		m.ValidateAndSetDefaults()

		// Apply field selection if specified
		if len(m.SelectedFields) > 0 && m.SelectedFields[0] != "*" {
			db = db.Select(m.SelectedFields)
		}

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
	// Capture start time for debug mode
	var startTime time.Time
	if m.Debug {
		startTime = time.Now()
	}

	// Validate metadata
	validation := m.Validate()
	if !validation.IsValid {
		return fmt.Errorf("invalid metadata: %v", validation.Errors)
	}

	// Create a clone of the DB for counting (to not affect field selection)
	countDB := db.Session(&gorm.Session{})

	// Get total count before applying pagination
	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return err
	}
	m.TotalRows = total

	// Debug: save the raw SQL
	var rawSQL string
	if m.Debug {
		rawSQL = db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Scopes(GPaginate(m)).Find(result)
		})
	}

	// Apply pagination and get results
	if err := db.Scopes(GPaginate(m)).Find(result).Error; err != nil {
		return err
	}

	// Update metadata with calculated values
	m.ValidateAndSetDefaults()

	// Encode cursor for next page if using cursor-based pagination
	if m.IsCursorBased() && m.HasNext {
		resultValue := reflect.ValueOf(result).Elem()
		if resultValue.Len() > 0 {
			lastItem := resultValue.Index(resultValue.Len() - 1).Interface()
			lastItemValue := reflect.ValueOf(lastItem)
			cursorData := map[string]interface{}{
				"id":   lastItemValue.FieldByName("ID").Interface(),
				"name": lastItemValue.FieldByName("Name").Interface(),
				"page": m.Page,
			}
			m.Cursor = encodeCursor(cursorData)
		}
	}

	// Add debug information
	if m.Debug {
		fmt.Printf("Query: %s\n", rawSQL)
		fmt.Printf("Query time: %v\n", time.Since(startTime))
		fmt.Printf("Total rows: %d\n", m.TotalRows)
		fmt.Printf("Total pages: %d\n", m.TotalPages)
	}

	return nil
}

// PaginateWithCount is similar to Paginate but allows you to specify a custom count query
// Useful when you need to count with specific conditions
func PaginateWithCount(db *gorm.DB, countQuery *gorm.DB, m *Metadata, result interface{}) error {
	// Capture start time for debug mode
	var startTime time.Time
	if m.Debug {
		startTime = time.Now()
	}

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

	// Debug: save the raw SQL
	var rawSQL string
	if m.Debug {
		rawSQL = db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Scopes(GPaginate(m)).Find(result)
		})
	}

	// Apply pagination and get results
	if err := db.Scopes(GPaginate(m)).Find(result).Error; err != nil {
		return err
	}

	// Update metadata with calculated values
	m.ValidateAndSetDefaults()

	// Encode cursor for next page if using cursor-based pagination
	if m.IsCursorBased() && m.HasNext {
		resultValue := reflect.ValueOf(result).Elem()
		if resultValue.Len() > 0 {
			lastItem := resultValue.Index(resultValue.Len() - 1).Interface()
			lastItemValue := reflect.ValueOf(lastItem)
			cursorData := map[string]interface{}{
				"id":   lastItemValue.FieldByName("ID").Interface(),
				"name": lastItemValue.FieldByName("Name").Interface(),
				"page": m.Page,
			}
			m.Cursor = encodeCursor(cursorData)
		}
	}

	// Add debug information
	if m.Debug {
		fmt.Printf("Query: %s\n", rawSQL)
		fmt.Printf("Query time: %v\n", time.Since(startTime))
		fmt.Printf("Total rows: %d\n", m.TotalRows)
		fmt.Printf("Total pages: %d\n", m.TotalPages)
	}

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
