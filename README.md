# GORM Metakit

A powerful pagination toolkit for Go applications using GORM and standard SQL databases. This package provides flexible pagination solutions with support for both offset-based and cursor-based pagination.

## Features

- 🔄 **Dual Pagination Support**
  - Offset-based pagination (traditional)
  - Cursor-based pagination (for better performance with large datasets)
- 📊 **Rich Metadata**
  - Total rows and pages
  - Current page information
  - Row range indicators
  - Navigation helpers (has next/previous)
- 🔍 **Sorting Support**
  - Flexible field sorting
  - Direction control (asc/desc)
- ✅ **Validation**
  - Input validation
  - Default value handling
  - Error reporting
- 🔗 **Method Chaining**
  - Fluent interface for easy configuration
  - Clear and readable code

## Installation

```bash
go get github.com/nccapo/paginate-metakit
```

## Quick Start

### Basic Usage

```go
import "github.com/nccapo/paginate-metakit"

// Create pagination metadata
metadata := metakit.NewMetadata().
    WithPage(1).
    WithPageSize(10).
    WithSort("created_at").
    WithSortDirection("desc")

// Use with GORM
var users []User
result := db.Model(&User{}).
    Order(metadata.GetSortClause()).
    Offset(metadata.GetOffset()).
    Limit(metadata.GetLimit()).
    Find(&users)

// Get total count
var total int64
db.Model(&User{}).Count(&total)
metadata.TotalRows = total
metadata.ValidateAndSetDefaults()
```

### Cursor-Based Pagination

```go
// Initialize cursor-based pagination
metadata := metakit.NewMetadata().
    WithCursorField("created_at").
    WithCursorOrder("desc").
    WithPageSize(10)

var users []User
query := db.Model(&User{})

// Apply cursor if provided
if metadata.Cursor != "" {
    cursorValue, _ := decodeCursor(metadata.Cursor)
    query = query.Where("created_at < ?", cursorValue)
}

// Execute query
result := query.
    Order(metadata.GetSortClause()).
    Limit(metadata.GetLimit() + 1).
    Find(&users)

// Handle pagination
hasMore := len(users) > metadata.GetLimit()
if hasMore {
    users = users[:metadata.GetLimit()]
}

// Set next cursor
if hasMore {
    lastUser := users[len(users)-1]
    nextCursor := encodeCursor(map[string]interface{}{
        "created_at": lastUser.CreatedAt,
        "id": lastUser.ID,
    })
    metadata.Cursor = nextCursor
}
```

## API Reference

### Metadata Configuration

```go
// Create new metadata with defaults
metadata := metakit.NewMetadata()

// Configure pagination
metadata.WithPage(1)           // Set page number
metadata.WithPageSize(10)      // Set items per page
metadata.WithSort("created_at") // Set sort field
metadata.WithSortDirection("desc") // Set sort direction

// Configure cursor-based pagination
metadata.WithCursorField("created_at") // Set cursor field
metadata.WithCursorOrder("desc")       // Set cursor order
metadata.WithCursor("base64-encoded-cursor") // Set cursor value
```

### Validation

```go
// Validate metadata
result := metadata.Validate()
if !result.IsValid {
    // Handle validation errors
    for _, err := range result.Errors {
        fmt.Printf("Error in %s: %s\n", err.Field, err.Message)
    }
}

// Validate and set defaults
metadata.ValidateAndSetDefaults()
```

### Helper Methods

```go
offset := metadata.GetOffset()    // Get current offset
limit := metadata.GetLimit()      // Get current limit
sortClause := metadata.GetSortClause() // Get formatted sort clause
isCursorBased := metadata.IsCursorBased() // Check pagination type
```

## Best Practices

1. **Always Validate**

   ```go
   result := metadata.Validate()
   if !result.IsValid {
       // Handle validation errors
   }
   ```

2. **Use Cursor-Based Pagination for Large Datasets**

   ```go
   metadata := metakit.NewMetadata().
       WithCursorField("created_at").
       WithCursorOrder("desc")
   ```

3. **Set Reasonable Page Sizes**

   ```go
   metadata.WithPageSize(20) // Default is 10, max is 100
   ```

4. **Handle Edge Cases**
   ```go
   if metadata.TotalRows == 0 {
       // Handle empty result set
   }
   ```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
