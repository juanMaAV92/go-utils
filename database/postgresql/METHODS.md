# Methods

## Condition forms

Every method that accepts `conditions any, args ...any` supports four forms:

```go
// 1. String + positional args (most common)
db.Find(ctx, &user, nil, "email = ?", "alice@example.com")
db.Find(ctx, &user, nil, "age > ? AND active = ?", 18, true)
db.Find(ctx, &user, nil, "id IN ?", []int{1, 2, 3})

// 2. Struct — zero values are IGNORED (empty string, 0, false are skipped)
db.Find(ctx, &user, nil, User{Name: "Alice", Role: "admin"})

// 3. Map — zero values ARE included
db.Find(ctx, &user, nil, map[string]any{"name": "Alice", "active": false})

// 4. nil — no WHERE clause (returns all records / deletes by primary key)
db.FindMany(ctx, &users, nil, nil)
```

> **Struct vs map:** use a struct when you only filter on non-zero fields. Use a map when you need to filter on `false`, `0`, or `""`.

---

## Create

Inserts a single record. `model.ID` is populated after insert.

```go
user := &User{Name: "Alice", Email: "alice@example.com"}
rows, err := db.Create(ctx, user)
```

---

## CreateMany

Inserts a slice of records in batches. `batchSize=0` defaults to 100.

```go
users := &[]User{{Name: "Alice"}, {Name: "Bob"}}
rows, err := db.CreateMany(ctx, users, 50)
```

---

## Find

Retrieves the **first** record matching conditions (ORDER BY primary key, LIMIT 1).
Returns `found=false, err=nil` when no record matches.

```go
var user User

// string condition
found, err := db.Find(ctx, &user, nil, "email = ?", "alice@example.com")

// struct condition
found, err := db.Find(ctx, &user, nil, User{Email: "alice@example.com"})

// with preloads
found, err := db.Find(ctx, &user, []string{"Profile", "Orders"}, "id = ?", id)
```

---

## FindMany

Retrieves **all** records matching conditions into a pointer to a slice.
Returns `found=false, err=nil` when the result set is empty.

```go
var users []User

// string condition
found, err := db.FindMany(ctx, &users, nil, "active = ?", true)

// map condition — includes zero values
found, err := db.FindMany(ctx, &users, nil, map[string]any{"role": "admin", "active": false})

// IN clause
found, err := db.FindMany(ctx, &users, nil, "id IN ?", []int{1, 2, 3})

// no condition — fetch all
found, err := db.FindMany(ctx, &users, nil, nil)

// with options
found, err := db.FindMany(ctx, &users, &postgresql.QueryOptions{
    OrderBy:  "created_at DESC",
    Preloads: []string{"Profile"},
    Joins:    []string{"JOIN orders ON orders.user_id = users.id"},
    Pagination: &postgresql.PaginationOptions{Page: 2, Limit: 20},
}, "active = ?", true)
```

---

## UpdateWhere

Updates fields on rows matching conditions.

**Important:** pass `map[string]any` when zero values (`false`, `0`, `""`) must be written — structs skip them.

```go
// map — zero values ARE written
rows, err := db.UpdateWhere(ctx, &User{}, map[string]any{
    "active":     false,
    "deleted_at": time.Now(),
}, "id = ?", userID)

// struct — zero values are skipped
rows, err := db.UpdateWhere(ctx, &User{}, User{Role: "admin"}, "id = ?", userID)

// GORM expression — for computed updates
rows, err := db.UpdateWhere(ctx, &Inventory{}, map[string]any{
    "stock": gorm.Expr("stock - ?", quantity),
}, "product_id = ?", productID)

// map condition
rows, err := db.UpdateWhere(ctx, &User{}, map[string]any{"active": true},
    map[string]any{"role": "guest"})
```

---

## Delete

Removes records matching conditions.
If the model has `gorm.DeletedAt`, GORM performs a **soft delete** automatically.

```go
// by primary key (nil conditions)
rows, err := db.Delete(ctx, &User{ID: 42}, nil)

// by string condition
rows, err := db.Delete(ctx, &User{}, "email = ?", "alice@example.com")

// by map condition
rows, err := db.Delete(ctx, &User{}, map[string]any{"active": false, "role": "guest"})

// handle sentinel errors
switch {
case errors.Is(err, postgresql.ErrInvalidReference):
    // a child record still references this row
}
```

---

## Count

Returns the number of rows matching conditions.

```go
// string condition
total, err := db.Count(ctx, &User{}, nil, "active = ?", true)

// no condition — count all
total, err := db.Count(ctx, &User{}, nil, nil)

// with join
total, err := db.Count(ctx, &Order{}, &postgresql.QueryOptions{
    Joins: []string{"JOIN users ON users.id = orders.user_id"},
}, "users.active = ?", true)
```

---

## Exec

Runs raw SQL. Use when GORM's query builder is not expressive enough.

Pass `model=nil` for non-SELECT (INSERT/UPDATE/DELETE).
Pass a pointer to a slice for SELECT — rows are scanned into it.

```go
// non-SELECT
result, err := db.Exec(ctx, nil,
    "UPDATE users SET last_seen = NOW() WHERE id = ?", userID)

// SELECT into anonymous struct
var rows []struct {
    Name  string
    Count int
}
result, err := db.Exec(ctx, &rows,
    "SELECT name, COUNT(*) as count FROM users GROUP BY name HAVING COUNT(*) > ?", 5)
```

---

## WithTransaction

Runs a function inside a transaction. Return `nil` to commit, return an error to rollback.
`tx` is a full `Database` scoped to the transaction — use it for all operations inside.

```go
err := db.WithTransaction(ctx, func(tx postgresql.Database) error {
    if _, err := tx.Create(ctx, &order); err != nil {
        return err // triggers rollback
    }
    if _, err := tx.UpdateWhere(ctx, &Inventory{},
        map[string]any{"stock": gorm.Expr("stock - ?", order.Quantity)},
        "product_id = ?", order.ProductID,
    ); err != nil {
        return err // triggers rollback
    }
    return nil // commit
})
```
