# Pagination-Metakit

Pagination-Metakit is a Go package designed to simplify pagination and sorting functionalities for applications using GORM (Go Object-Relational Mapping) and Pure SQL. This package provides a Metadata structure to manage pagination and sorting parameters, and a GORM scope function to apply these settings to database queries, but not for only GORM, Pagination-Metakit also support pure sql pagination.

## Overview

- Pagination: Easily handle pagination with customizable page size
- Sorting: Built-in support for field sorting with direction control
- Default Settings: Provides sensible defaults for page, page size, and sort direction
- Dual Functionality: Supports both GORM and pure SQL pagination
- Rich Metadata: Includes pagination info like total rows, pages, and navigation helpers

## Installation

To install the package, use go get:

```shell
go get github.com/nccapo/gorm-metakit
```

## Usage

### Metadata Structure

The Metadata structure holds the pagination and sorting parameters:

```go
type Metadata struct {
    Page          int    `form:"page" json:"page"`
    PageSize      int    `form:"page_size" json:"page_size"`
    Sort          string `form:"sort" json:"sort"`
    SortDirection string `form:"sort_direction" json:"sort_direction"`
    TotalRows     int64  `json:"total_rows"`
    TotalPages    int64  `json:"total_pages"`
    HasNext       bool   `json:"has_next"`
    HasPrevious   bool   `json:"has_previous"`
    FromRow       int64  `json:"from_row"`
    ToRow         int64  `json:"to_row"`
}
```

### Creating Metadata

```go
// Create with defaults
metadata := metakit.NewMetadata()

// Or customize with method chaining
metadata := metakit.NewMetadata().
    WithPage(1).
    WithPageSize(20).
    WithSort("created_at").
    WithSortDirection("desc")
```

## Example Usage with GORM

```go
package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "github.com/nccapo/gorm-metakit"
    "net/http"
)

func main() {
    r.GET("/items", func(c *gin.Context) {
        // Create metadata with defaults
        metadata := metakit.NewMetadata()

        // Bind query parameters
        if err := c.ShouldBind(&metadata); err != nil {
            c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err})
            return
        }

        // Fetch paginated and sorted results
        var results []YourModel
        err := metakit.Paginate(db, metadata, &results)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "metadata": metadata,
            "results": results,
        })
    })
}
```

## Example Usage with Pure SQL

```go
package main

import (
    "database/sql"
    "log"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/nccapo/gorm-metakit"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "user=youruser dbname=yourdb sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    r := gin.Default()

    r.GET("/items", func(c *gin.Context) {
        metadata := metakit.NewMetadata()

        // Bind query parameters
        if err := c.ShouldBind(&metadata); err != nil {
            c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err})
            return
        }

        // Create custom count query if needed
        countQuery := db.QueryRow("SELECT COUNT(*) FROM your_table WHERE status = ?", "active")
        var total int64
        if err := countQuery.Scan(&total); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err})
            return
        }
        metadata.TotalRows = total

        // Fetch paginated and sorted results
        query := "SELECT * FROM your_table WHERE status = ?"
        rows, err := metakit.QueryContextPaginate(context.Background(), db, "active", query, metadata)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err})
            return
        }
        defer rows.Close()

        var results []YourModel
        for rows.Next() {
            var item YourModel
            // Scan your results here
            results = append(results, item)
        }

        c.JSON(http.StatusOK, gin.H{
            "metadata": metadata,
            "results": results,
        })
    })

    r.Run()
}
```

## Contributing

This project is licensed under the MIT License. See the LICENSE file for details.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
