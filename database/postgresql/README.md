# Postgresql Package

This package provides a high-level database abstraction layer built on top of GORM for PostgreSQL databases. It includes connection management, migrations, error handling, and a clean API for common database operations.

## Features

- **Connection Management**: Singleton pattern with connection pooling and retries
- **Automatic Migrations**: Built-in migration support using golang-migrate
- **Error Handling**: User-friendly error messages with PostgreSQL error code mapping
- **Query Builder**: Support for complex queries including JOINs and raw SQL
- **Pagination**: Built-in pagination support
- **Validation**: Comprehensive parameter validation
- **Context Support**: All operations support context for cancellation and timeouts
- **Logging**: Integrated logging with configurable levels

## Main Files

- `database.go`: Core database operations (Create, Update, FindOne, FindMany, Count, etc.)
- `config.go`: Connection configuration and initialization
- `model.go`: Data models and configuration structures
- `database_test.go`: Unit tests for validation logic

## Configuration

The package uses environment variables for database configuration:

```md
DB_HOST_POSTGRES = IP address of the PostgreSQL server
DB_PORT_POSTGRES = Port of the PostgreSQL server
DB_USER_POSTGRES = Username for the PostgreSQL database
DB_PASSWORD_POSTGRES = Password for the PostgreSQL database
DB_NAME_POSTGRES = Name of the PostgreSQL database
DB_MAX_POOL_SIZE_POSTGRES = Maximum number of connections in the pool (default: 2)
DB_MAX_LIFE_TIME_POSTGRES = Maximum lifetime of a connection in the pool (default: "5m")
```

## Usage

### Initialize Database Connection

```go
import "github.com/juanMaAV92/go-utils/database"

// Get configuration from environment variables
config := database.GetDBConfig()

// Create database instance (singleton)
db, err := database.New(*config, logger)
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}
```

### Basic CRUD Operations

#### Create Records
```go
user := &User{Name: "John Doe", Email: "john@example.com"}
err := db.Create(ctx, user)
if err != nil {
    // Handle error
}
```

#### Update Records
```go
// Basic update
updates := map[string]interface{}{
    "name": "Jane Doe",
    "email": "jane@example.com",
}
rowsAffected, err := db.Update(ctx, &user, updates, nil)

// Update with conditions (prevent race conditions)
rowsAffected, err := db.Update(ctx, &user, updates, "version = ? AND status = ?", 5, "ACTIVE")
if err != nil {
    // Handle error
}
if rowsAffected == 0 {
    // No records updated (not found or conditions not met)
}
```

#### Find Single Record
```go
var user User
found, err := db.FindOne(ctx, &user, map[string]interface{}{
    "email": "john@example.com",
}, nil)
if err != nil {
    // Handle error
}
if !found {
    // Record not found
}
```

#### Find Single Record (with Preloads)
```go
var user User
preloads := []string{"Profile", "Orders"}
found, err := db.FindOne(ctx, &user, map[string]interface{}{
    "email": "john@example.com",
}, preloads)
if err != nil {
    // Handle error
}
if !found {
    // Record not found
}
```

#### Find Multiple Records
```go
var users []User
options := &database.QueryOptions{
    Pagination: &database.PaginationOptions{
        Page:  1,
        Limit: 10,
    },
    OrderBy: "created_at DESC",
    Preloads: []string{"Profile", "Orders"},
}

err := db.FindMany(ctx, &users, map[string]interface{}{
    "active": true,
}, options)
```

#### Count Records
```go
// Count all records
totalUsers, err := db.Count(ctx, &User{}, nil)

// Count with conditions
activeUsers, err := db.Count(ctx, &User{}, "status <> ? AND created_at > ?", "INACTIVE", time.Now().AddDate(0, -1, 0))
```

### Query Conditions

The `conditions` parameter supports multiple formats for flexible querying. Currently available in `Count()` and `Update()` functions, with planned migration to all Find operations.

> **Note**: Advanced condition support with placeholders is implemented for `Count()` and `Update()`. Future versions will extend this functionality to `FindOne()`, `FindMany()`, and other query operations.

#### Map Conditions (Equality only)
```go
conditions := map[string]interface{}{
    "status": "ACTIVE",
    "verified": true,
}
```

#### String Conditions with Placeholders (Recommended)
```go
// Comparison operators
"status <> ?"          // Not equal
"amount > ?"           // Greater than
"created_at >= ?"      // Greater or equal

// Multiple conditions
"status = ? AND amount > ? AND created_at > ?"

// IN operator
"status IN ?"          // Pass slice as argument

// LIKE for partial matches
"name LIKE ?"          // Pass "%pattern%" as argument

// IS NULL / IS NOT NULL
"deleted_at IS NULL"   // No arguments needed
```

#### Common Examples
```go
// Count examples
db.Count(ctx, &Order{}, "status <> ?", "CANCELLED")
db.Count(ctx, &Order{}, "amount BETWEEN ? AND ?", 100.0, 1000.0)

// Update examples (race condition prevention)
updates := map[string]interface{}{"status": "PROCESSING"}
db.Update(ctx, &order, updates, "status = ? AND version = ?", "PENDING", 5)

// IN operator with slice
statuses := []string{"PENDING", "PROCESSING"}
db.Count(ctx, &Order{}, "status IN ?", statuses)

// Pattern matching
db.Count(ctx, &User{}, "email LIKE ?", "%@gmail.com")

// Complex conditions
db.Count(ctx, &Order{}, "status <> ? AND amount > ? AND created_at > ?", 
    "CANCELLED", 100.0, time.Now().AddDate(0, -1, 0))
```

### Advanced Queries

#### JOIN Queries
```go
var results []UserWithProfile
config := database.JoinConfig{
    BaseTable: "users",
    Joins: []database.JoinClause{
        {
            Type:  "LEFT",
            Table: "profiles", 
            On:    "users.id = profiles.user_id",
        },
    },
    Select: "users.name, users.email, profiles.bio",
    Conditions: map[string]interface{}{
        "users.active": true,
    },
    OrderBy: "users.created_at DESC",
    Limit:   50,
}

err := db.FindWithJoins(ctx, &results, config)
```

#### Raw SQL Queries
```go
var results []CustomResult
query := `
    SELECT u.name, COUNT(o.id) as order_count 
    FROM users u 
    LEFT JOIN orders o ON u.id = o.user_id 
    WHERE u.created_at > ? 
    GROUP BY u.id, u.name
`
err := db.ExecuteRawQuery(ctx, &results, query, time.Now().AddDate(0, -1, 0))
```

## Error Handling

The package provides user-friendly error messages for common PostgreSQL errors:

- **Unique Violation (23505)**: "A record with the same values already exists"
- **Check Violation (23514)**: "The provided data violates database constraints"
- **Foreign Key Violation (23503)**: "Invalid reference in the provided data"
- **Generic Errors**: "An unexpected database error occurred"

## Migration Support

Migrations are automatically applied on startup. Place your migration files in the `migration/` directory relative to your main application path.

Migration files should follow the format:
- `000001_description.up.sql`
- `000001_description.down.sql`


## Requirements

- Go 1.18+
- PostgreSQL
- GORM v2
- golang-migrate/migrate

## Dependencies

- `gorm.io/gorm`: ORM framework
- `gorm.io/driver/postgres`: PostgreSQL driver
- `github.com/golang-migrate/migrate/v4`: Database migrations
- `github.com/jackc/pgconn`: PostgreSQL connection library

## Testing

The package includes comprehensive unit tests for validation logic. Run tests with:

```bash
go test ./database
```

## License

MIT
