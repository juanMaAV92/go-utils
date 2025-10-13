# Pagination Package

This package provides utilities for handling pagination in web applications, specifically designed to work with Echo framework query parameters and build pagination metadata.

## Features

- **Query Parameter Extraction**: Extract and validate pagination parameters from HTTP requests
- **Automatic Validation**: Handle invalid parameters gracefully with fallback to defaults
- **Pagination Metadata**: Build complete pagination information including total pages calculation
- **Echo Framework Integration**: Seamless integration with Echo web framework

## Main Types and Functions

- `Pagination`: Struct to hold pagination metadata (total pages, total items, current page, limit)
- `ExtractParams(c echo.Context, defaultPage, defaultLimit int) (page, limit int)`: Extract pagination parameters from Echo context
- `BuildPagination(total int64, page, limit int) Pagination`: Build pagination metadata from query results

## Usage Example

```go
import (
    "github.com/juanMaAV92/go-utils/pagination"
    "github.com/labstack/echo/v4"
)

func GetUsersHandler(c echo.Context) error {
    // Extract pagination parameters with defaults
    page, limit := pagination.ExtractParams(c, 1, 10)
    
    // Query your database with pagination
    users, total, err := userService.GetUsers(page, limit)
    if err != nil {
        return err
    }
    
    // Build pagination metadata
    paginationInfo := pagination.BuildPagination(total, page, limit)
    
    return c.JSON(200, map[string]interface{}{
        "users":      users,
        "pagination": paginationInfo,
    })
}
```

## Parameter Validation

The `ExtractParams` function handles various edge cases:

- **Missing parameters**: Uses provided defaults
- **Invalid strings**: Falls back to defaults (e.g., "abc" for page)
- **Zero or negative values**: Falls back to defaults
- **Non-numeric values**: Falls back to defaults

```go
// URL: /users?page=2&limit=20
page, limit := pagination.ExtractParams(c, 1, 10) // Returns: 2, 20

// URL: /users?page=invalid&limit=0  
page, limit := pagination.ExtractParams(c, 1, 10) // Returns: 1, 10 (defaults)

// URL: /users (no params)
page, limit := pagination.ExtractParams(c, 1, 10) // Returns: 1, 10 (defaults)
```

## Pagination Metadata Structure

```go
type Pagination struct {
    TotalPages int `json:"total_pages"` // Total number of pages
    TotalItems int `json:"total_items"` // Total number of items
    Page       int `json:"page"`        // Current page number
    Limit      int `json:"limit"`       // Items per page
}
```

## Example Response

```json
{
    "users": [...],
    "pagination": {
        "total_pages": 5,
        "total_items": 47,
        "page": 2,
        "limit": 10
    }
}
```

## Testing

The package includes comprehensive unit tests covering:
- Parameter extraction with various input scenarios
- Pagination calculation with edge cases (empty results, partial pages, exact boundaries)
- Integration with Echo framework

Run tests with:
```bash
go test ./pagination
```

## Dependencies

- `github.com/labstack/echo/v4`: Web framework for HTTP context handling
- Standard library: `strconv` for string to integer conversion

## License

MIT
