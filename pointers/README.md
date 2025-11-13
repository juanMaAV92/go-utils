# Pointers Package

Utilidades para trabajar con punteros en Go.

## Funciones

### `Pointer[T any](v T) *T`
Devuelve un puntero a cualquier valor.

```go
intPtr := pointers.Pointer(42)
strPtr := pointers.Pointer("hello")
```

### `FirstNonNil[T any](values ...*T) *T`
Retorna el primer puntero no nulo de la lista.

```go
result := pointers.FirstNonNil(nil, pointers.Pointer(10), pointers.Pointer(20))
// result apunta a 10
```

### `FirstNotNilOrEmptyString(values ...*string) *string`
Retorna el primer puntero que no sea nil ni apunte a un string vacío.

```go
result := pointers.FirstNotNilOrEmptyString(nil, pointers.Pointer(""), pointers.Pointer("hello"))
// result apunta a "hello"
```

### `StringPtr(s string) *string`
Retorna un puntero al string solo si no está vacío, de lo contrario retorna nil.

```go
ptr := pointers.StringPtr("hello")  // puntero a "hello"
ptr := pointers.StringPtr("")       // nil
```

## Uso común

```go
type User struct {
    Name  string
    Email *string  // campo opcional
}

user := User{
    Name:  "John",
    Email: pointers.StringPtr(emailInput), // nil si emailInput está vacío
}
```
