# database/postgresql

PostgreSQL client built on [GORM](https://gorm.io) with OTel tracing, structured error handling, and optional migration support.

## Setup

```go
import "github.com/juanmaAV/go-utils/database/postgresql"

// From environment variables
cfg, err := postgresql.ConfigFromEnv("DB")  // reads DB_HOST, DB_USER, DB_PASSWORD, DB_NAME …
if err != nil {
    log.Fatal(err)
}

db, err := postgresql.New(cfg, logger)
if err != nil {
    log.Fatal(err)
}
```

### Config fields

| Field | Env var (prefix=DB) | Default |
|---|---|---|
| `Host` | `DB_HOST` | required |
| `Port` | `DB_PORT` | `5432` |
| `User` | `DB_USER` | required |
| `Password` | `DB_PASSWORD` | required |
| `Name` | `DB_NAME` | required |
| `SSLMode` | `DB_SSLMODE` | `require` |
| `MaxPoolSize` | `DB_MAX_POOL_SIZE` | `2` |
| `MaxLifeTime` | `DB_MAX_LIFE_TIME` | `5m` |
| `Verbose` | — | `false` |

Multiple databases use different prefixes:

```go
ordersDB, _ := postgresql.ConfigFromEnv("ORDERS_DB")  // ORDERS_DB_HOST …
auditDB,  _ := postgresql.ConfigFromEnv("AUDIT_DB")   // AUDIT_DB_HOST …
```

### Migrations

Migrations are **not** run automatically. Call explicitly at startup:

```go
err := postgresql.RunMigrations(cfg, "file:///app/migrations")
```

### Sentinel errors

```go
_, err := db.Create(ctx, &user)
switch {
case errors.Is(err, postgresql.ErrDuplicateRecord):    // unique constraint
case errors.Is(err, postgresql.ErrConstraintViolation): // check constraint
case errors.Is(err, postgresql.ErrInvalidReference):   // foreign key
}
```

## Methods

See [METHODS.md](METHODS.md) for usage examples of each method.

## Interface

```go
type Database interface {
    Create(ctx, model) (affectedRows int64, err error)
    CreateMany(ctx, models, batchSize int) (affectedRows int64, err error)
    Find(ctx, model, preloads []string, conditions, args...) (found bool, err error)
    FindMany(ctx, model, options *QueryOptions, conditions, args...) (found bool, err error)
    UpdateWhere(ctx, model, updates, conditions, args...) (affectedRows int64, err error)
    Delete(ctx, model, conditions, args...) (affectedRows int64, err error)
    Count(ctx, model, options *QueryOptions, conditions, args...) (count int64, err error)
    Exec(ctx, model, sql string, args...) (QueryResult, error)
    WithTransaction(ctx, fn TransactionFunc) error
}
```
