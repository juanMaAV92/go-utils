# timeutil

Time utilities: nil-safe parsing, comparisons, and day boundaries. Complements the stdlib `time` package for the cases it doesn't cover cleanly.

Not included:
- `TimePtr`/`TimeValue` — use `pointers.Pointer`/`pointers.Value`
- Format functions — use `t.Format(timeutil.DateLayout)` directly

## Constants

```go
timeutil.DateLayout     // "2006-01-02"
timeutil.DateTimeLayout // time.RFC3339
```

Exported to avoid repeating Go's magic date format string across the codebase.

## Functions

### Parsing

```go
func ParseDate(value string) (*time.Time, error)
func ParseDateTime(value string) (*time.Time, error)
```
Parse date/datetime strings. Return `nil, nil` for empty input. Return error on invalid format.

### Comparisons

```go
func IsZeroOrNil(t *time.Time) bool
func After(t *time.Time, u time.Time) bool
func Before(t *time.Time, u time.Time) bool
```
Nil-safe wrappers. `After`/`Before` return `false` if `t` is nil.

### Day boundaries

```go
func StartOfDay(t time.Time) time.Time  // 00:00:00.000000000
func EndOfDay(t time.Time) time.Time    // 23:59:59.999999999
```
Preserve the original location. Useful for date-range queries.

## Usage

```go
import "github.com/juanmaAV/go-utils/timeutil"

// Use constants to avoid the magic format string
t.Format(timeutil.DateLayout)      // "2024-03-15"
t.Format(timeutil.DateTimeLayout)  // "2024-03-15T10:30:00Z"

// Parsing
t, err := timeutil.ParseDate("2024-03-15")
t, err := timeutil.ParseDateTime("2024-03-15T10:30:00Z")
t, err := timeutil.ParseDate("")  // nil, nil — empty is valid

// Nil-safe comparisons
timeutil.IsZeroOrNil(t)
timeutil.After(t, time.Now())
timeutil.Before(t, deadline)

// Date range queries
start := timeutil.StartOfDay(time.Now())
end   := timeutil.EndOfDay(time.Now())
// WHERE created_at BETWEEN start AND end
```

## Notes

- `ParseDate`/`ParseDateTime` return `*time.Time` (not `time.Time`) so the nil case (empty string) is distinguishable from a parsed zero time.
- `StartOfDay`/`EndOfDay` use the location embedded in the input — pass UTC or a specific timezone explicitly if needed.
- No external dependencies.
