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
- **Transactions**: Safe transaction management with automatic commit/rollback
- **Advanced Conditions**: Support for placeholders and arguments in all query operations
- **Named Returns**: Self-documenting function returns for better IDE experience

## Main Files

- `database.go`: Core database operations (Create, Update, FindOne, FindMany, Count, WithTransaction, etc.)
- `config.go`: Connection configuration and initialization
- `model.go`: Data models, configuration structures, and type definitions
- `errorHandler.go`: Centralized error handling and user-friendly error messages
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

// With string conditions and arguments
found, err := db.FindOne(ctx, &user, "status = ? AND created_at > ?", nil, "ACTIVE", time.Now().AddDate(0, -1, 0))
```

#### Find Single Record (with Preloads)
```go
var user User
preloads := []string{"Profile", "Orders"}
found, err := db.FindOne(ctx, &user, map[string]interface{}{
    "email": "john@example.com",
}, &preloads)
if err != nil {
    // Handle error
}
if !found {
    // Record not found
}

// With string conditions, preloads and arguments
found, err := db.FindOne(ctx, &user, "status = ? AND verified = ?", &preloads, "ACTIVE", true)
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

// With string conditions and arguments
err := db.FindMany(ctx, &users, "status = ? AND created_at > ?", options, "ACTIVE", time.Now().AddDate(0, -1, 0))
```

#### Count Records
```go
// Count all records
totalUsers, err := db.Count(ctx, &User{}, nil)

// Count with conditions
activeUsers, err := db.Count(ctx, &User{}, "status <> ? AND created_at > ?", "INACTIVE", time.Now().AddDate(0, -1, 0))
```

### Transactions

The package provides safe transaction management with automatic commit/rollback functionality.

#### Basic Transaction Usage
```go
err := db.WithTransaction(ctx, func(tx *Database) error {
    // Create user
    user := &User{Name: "Juan", Email: "juan@example.com"}
    if err := tx.Create(ctx, user); err != nil {
        return err // Automatic rollback
    }
    
    // Create profile
    profile := &Profile{UserID: user.ID, Bio: "Developer"}
    if err := tx.Create(ctx, profile); err != nil {
        return err // Automatic rollback
    }
    
    return nil // Automatic commit
})
if err != nil {
    // Handle transaction error
}
```

#### Complex Transaction Example
```go
// Bank transfer with transaction safety
err := db.WithTransaction(ctx, func(tx *Database) error {
    // Debit from source account
    updates := map[string]interface{}{"balance": gorm.Expr("balance - ?", amount)}
    rowsAffected, err := tx.Update(ctx, &fromAccount, updates, 
        "id = ? AND balance >= ?", fromID, amount)
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return errors.New("insufficient funds")
    }
    
    // Credit to destination account
    updates = map[string]interface{}{"balance": gorm.Expr("balance + ?", amount)}
    _, err = tx.Update(ctx, &toAccount, updates, "id = ?", toID)
    if err != nil {
        return err
    }
    
    // Create transaction record
    txRecord := &Transaction{FromID: fromID, ToID: toID, Amount: amount}
    return tx.Create(ctx, txRecord)
})
```

#### Transaction with Raw SQL
```go
err := db.WithTransaction(ctx, func(tx *Database) error {
    // Update inventory
    result, err := tx.ExecuteRawQuery(ctx, nil, 
        "UPDATE inventory SET quantity = quantity - ? WHERE product_id = ? AND quantity >= ?", 
        quantity, productID, quantity)
    if err != nil {
        return err
    }
    if result.RowsAffected == 0 {
        return errors.New("insufficient inventory")
    }
    
    // Create order
    order := &Order{ProductID: productID, Quantity: quantity}
    return tx.Create(ctx, order)
})
```

**Transaction Features:**
- **Automatic Rollback**: On any error or panic
- **Automatic Commit**: When function returns nil
- **Panic Recovery**: Converts panics to errors with rollback
- **Nested Operations**: All database operations work within transactions
- **Context Support**: Full context propagation and cancellation

### Query Conditions

The `conditions` parameter supports multiple formats for flexible querying. This functionality is now available in all query operations: `FindOne()`, `FindMany()`, `Count()`, and `Update()`.

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
// FindOne examples
db.FindOne(ctx, &user, "status <> ?", nil, "INACTIVE")
db.FindOne(ctx, &user, "email LIKE ?", nil, "%@gmail.com")

// FindMany examples
db.FindMany(ctx, &users, "status = ? AND created_at > ?", options, "ACTIVE", time.Now().AddDate(0, -1, 0))
db.FindMany(ctx, &orders, "amount BETWEEN ? AND ?", options, 100.0, 1000.0)

// Count examples
db.Count(ctx, &Order{}, "status <> ?", "CANCELLED")
db.Count(ctx, &Order{}, "amount BETWEEN ? AND ?", 100.0, 1000.0)

// Update examples (race condition prevention)
updates := map[string]interface{}{"status": "PROCESSING"}
db.Update(ctx, &order, updates, "status = ? AND version = ?", "PENDING", 5)

// IN operator with slice
statuses := []string{"PENDING", "PROCESSING"}
db.FindMany(ctx, &orders, "status IN ?", options, statuses)
db.Count(ctx, &Order{}, "status IN ?", statuses)

// Pattern matching
db.FindOne(ctx, &user, "email LIKE ?", nil, "%@gmail.com")
db.Count(ctx, &User{}, "email LIKE ?", "%@gmail.com")

// Complex conditions
db.FindMany(ctx, &orders, "status <> ? AND amount > ? AND created_at > ?", options,
    "CANCELLED", 100.0, time.Now().AddDate(0, -1, 0))
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

The `ExecuteRawQuery` method now returns detailed information about query results.

```go
// SELECT queries with result information
var results []CustomResult
query := `
    SELECT u.name, COUNT(o.id) as order_count 
    FROM users u 
    LEFT JOIN orders o ON u.id = o.user_id 
    WHERE u.created_at > ? 
    GROUP BY u.id, u.name
`
result, err := db.ExecuteRawQuery(ctx, &results, query, time.Now().AddDate(0, -1, 0))
if err != nil {
    // Handle error
}

if result.Found {
    fmt.Printf("Found %d users with orders\n", result.RowsAffected)
    // Process results...
} else {
    fmt.Println("No users found")
}
```

```go
// INSERT/UPDATE/DELETE queries with affected rows
result, err := db.ExecuteRawQuery(ctx, nil, 
    "UPDATE users SET last_login = NOW() WHERE status = ?", "ACTIVE")
if err != nil {
    // Handle error
}

fmt.Printf("Updated %d users\n", result.RowsAffected)
```

**QueryResult Structure:**
```go
type QueryResult struct {
    RowsAffected int64 `json:"rows_affected"` // Number of rows affected/found
    Found        bool  `json:"found"`         // Whether any records were found/affected
}
```

## Named Returns & Autodocumentation

All functions use named returns for better IDE experience and self-documenting code:

```go
// Instead of generic types, you get descriptive names
rowsAffected, err := db.Update(...)     // Clear: number of affected rows
found, err := db.FindOne(...)           // Clear: whether record was found  
totalRecords, err := db.Count(...)      // Clear: total count of records
result, err := db.ExecuteRawQuery(...)  // Clear: detailed query result
```

**Benefits:**
- **Better IDE Support**: IntelliSense shows descriptive parameter names
- **Self-Documenting**: Code is more readable without additional comments
- **Type Safety**: Same functionality with improved developer experience

## Error Handling

The package provides user-friendly error messages for common PostgreSQL errors with centralized error management:

- **Unique Violation (23505)**: "A record with the same values already exists"
- **Check Violation (23514)**: "The provided data violates database constraints"
- **Foreign Key Violation (23503)**: "Invalid reference in the provided data"
- **Generic Errors**: "An unexpected database error occurred"

**Error Management Features:**
- **Centralized Error Functions**: All errors are created through dedicated functions
- **Consistent Messaging**: No magic strings, all error messages are constants
- **Maintainable**: Easy to update error messages from a single location
- **Localization Ready**: Structured for future internationalization support

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

## Recent Updates

### Version 2.0 Features

**🔄 Transactions Support**
- Added `WithTransaction()` method with automatic commit/rollback
- Panic recovery and error handling
- Full context support and operation compatibility

**📊 Enhanced Raw SQL Queries**
- `ExecuteRawQuery()` now returns `QueryResult` with detailed information
- Support for both SELECT and non-SELECT queries
- Clear indication of rows affected and records found

**🏷️ Named Returns**
- All functions now use descriptive named returns
- Better IDE experience with IntelliSense
- Self-documenting code without additional comments

**⚡ Advanced Query Conditions**
- Extended placeholder support to `FindOne()` and `FindMany()`
- Consistent argument handling across all query methods
- Enhanced flexibility for complex conditions

**🛠️ Centralized Error Management**
- Eliminated magic strings throughout the codebase
- Centralized error creation functions
- Improved maintainability and consistency

## License

MIT
