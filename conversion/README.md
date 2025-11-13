# Conversion Package

Utilidades para conversión de tipos en Go.

## Funciones

### `UUIDToString(val interface{}) string`
Convierte UUID a string. Retorna string vacío si el valor es nil o tipo no soportado.

### `ToUUID(val interface{}) (uuid.UUID, error)`
Convierte string o UUID a tipo UUID. Retorna error solo si el string es inválido.

### `ToInt64(val interface{}) (int64, error)`
Convierte varios tipos numéricos y strings a int64. Retorna error solo si el string es inválido.

## Ejemplos de uso

### UUID

```go
// UUID a string
id := uuid.New()
str := conversion.UUIDToString(id)

// String a UUID
id, err := conversion.ToUUID("550e8400-e29b-41d4-a716-446655440000")
if err != nil {
    log.Fatal(err)
}
```

### Números

```go
// int, int32, uint, float64 a int64
val, err := conversion.ToInt64(42)
val, err := conversion.ToInt64(int32(100))
val, err := conversion.ToInt64(uint(250))
val, err := conversion.ToInt64(123.45) // trunca a 123

// String a int64
val, err := conversion.ToInt64("789")
if err != nil {
    log.Fatal(err)
}
```

## Tipos soportados

| Función | Entrada | Salida |
|---------|---------|--------|
| `UUIDToString` | `uuid.UUID`, `string`, `nil` | `string` |
| `ToUUID` | `string`, `uuid.UUID`, `nil` | `uuid.UUID`, `error` |
| `ToInt64` | `int`, `uint`, `float`, `string`, otros | `int64`, `error` |
