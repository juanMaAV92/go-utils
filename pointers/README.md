# pointers

Generic helpers for working with pointer types.

## Functions

```go
func Pointer[T any](v T) *T
```
Returns a pointer to `v`. Works for any type.

```go
func Value[T any](p *T) T
```
Dereferences `p`. Returns the zero value of `T` if `p` is nil.

```go
func FirstNonNil[T any](values ...*T) *T
```
Returns the first non-nil pointer. Returns nil if all are nil.

```go
func FirstNonNilOrEmptyString(values ...*string) *string
```
Returns the first pointer that is non-nil and points to a non-empty string.

```go
func StringPtr(s string) *string
```
Returns a pointer to `s`, or nil if `s == ""`.

```go
func StringValue(s *string) string
```
Dereferences `s`. Returns `""` if nil.

## Usage

```go
import "github.com/juanmaAV/go-utils/pointers"

p := pointers.Pointer(42)             // *int
s := pointers.Pointer("hello")        // *string

n := pointers.Value(p)                // 42
z := pointers.Value[int](nil)         // 0

first := pointers.FirstNonNil(a, b, c)
str   := pointers.FirstNonNilOrEmptyString(x, y, z)

ptr := pointers.StringPtr("value")    // *string, nil if ""
val := pointers.StringValue(ptr)      // string, "" if nil
```

## Notes

- `Value` is the generic equivalent of `StringValue` for any type.
- `StringPtr`/`StringValue` are convenience wrappers for the common string case.
- No external dependencies.
