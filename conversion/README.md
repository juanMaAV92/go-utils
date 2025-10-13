# Conversion Package

This package provides utility functions for converting between different representations of UUIDs in Go applications, handling type safety and error cases gracefully.

## Features

- **Type-Safe UUID Conversion**: Convert between string and UUID types safely
- **Flexible Input Handling**: Accept multiple input types (string, UUID, nil)
- **Error Handling**: Graceful handling of invalid inputs and parsing errors
- **Nil Safety**: Proper handling of nil inputs without panics

## Main Functions

- `UUIDToString(val interface{}) string`: Convert any supported type to string representation
- `ToUUID(val interface{}) (uuid.UUID, error)`: Convert any supported type to UUID with error handling

## Usage Examples

### Converting to String

```go
import (
    "github.com/juanMaAV92/go-utils/conversion"
    "github.com/google/uuid"
)

func ExampleUUIDToString() {
    // From UUID type
    id := uuid.New()
    str := conversion.UUIDToString(id)
    fmt.Println(str) // Output: "550e8400-e29b-41d4-a716-446655440000"
    
    // From string (passthrough)
    existing := "550e8400-e29b-41d4-a716-446655440000"
    str = conversion.UUIDToString(existing)
    fmt.Println(str) // Output: "550e8400-e29b-41d4-a716-446655440000"
    
    // From nil
    str = conversion.UUIDToString(nil)
    fmt.Println(str) // Output: ""
    
    // From unsupported type
    str = conversion.UUIDToString(123)
    fmt.Println(str) // Output: ""
}
```

### Converting to UUID

```go
func ExampleToUUID() {
    // From valid string
    id, err := conversion.ToUUID("550e8400-e29b-41d4-a716-446655440000")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(id.String())
    
    // From UUID type (passthrough)
    originalID := uuid.New()
    id, err = conversion.ToUUID(originalID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(id.String())
    
    // From nil
    id, err = conversion.ToUUID(nil)
    fmt.Println(id) // Output: 00000000-0000-0000-0000-000000000000 (uuid.Nil)
    fmt.Println(err) // Output: <nil>
    
    // From invalid string
    id, err = conversion.ToUUID("not-a-uuid")
    if err != nil {
        fmt.Printf("Error: %v\n", err) // Output: Error: invalid UUID length: 10
    }
    
    // From unsupported type
    id, err = conversion.ToUUID(123)
    fmt.Println(id) // Output: 00000000-0000-0000-0000-000000000000 (uuid.Nil)
    fmt.Println(err) // Output: <nil>
}
```

## Practical Use Cases

### Database Operations

```go
// When reading from database where UUID might be stored as string
func GetUserByID(userID interface{}) (*User, error) {
    id, err := conversion.ToUUID(userID)
    if err != nil {
        return nil, fmt.Errorf("invalid user ID format: %w", err)
    }
    
    return userRepo.FindByID(id)
}

// When preparing data for JSON response
func (u *User) ToJSON() map[string]interface{} {
    return map[string]interface{}{
        "id":    conversion.UUIDToString(u.ID),
        "name":  u.Name,
        "email": u.Email,
    }
}
```

### API Parameter Handling

```go
func GetUserHandler(c echo.Context) error {
    userIDParam := c.Param("id")
    
    // Convert string parameter to UUID
    userID, err := conversion.ToUUID(userIDParam)
    if err != nil {
        return c.JSON(400, map[string]string{
            "error": "Invalid user ID format",
        })
    }
    
    user, err := userService.GetByID(userID)
    if err != nil {
        return err
    }
    
    return c.JSON(200, user)
}
```

### Configuration and Environment Variables

```go
// When reading UUID from environment variables
func LoadConfig() *Config {
    return &Config{
        ServiceID: conversion.UUIDToString(os.Getenv("SERVICE_ID")),
        // ... other config
    }
}

// When validating configuration UUIDs
func ValidateConfig(cfg *Config) error {
    if _, err := conversion.ToUUID(cfg.ServiceID); err != nil {
        return fmt.Errorf("invalid service ID in config: %w", err)
    }
    return nil
}
```

## Type Support Matrix

| Input Type | UUIDToString | ToUUID |
|------------|--------------|--------|
| `nil` | `""` | `uuid.Nil, nil` |
| `string` | passthrough | parsed or error |
| `uuid.UUID` | `.String()` | passthrough |
| other types | `""` | `uuid.Nil, nil` |

## Error Handling

### UUIDToString
- Never returns errors
- Returns empty string for nil or unsupported types
- Always safe to call

### ToUUID
- Returns error only for invalid string formats
- Returns `uuid.Nil` for nil or unsupported types (no error)
- Use the error to distinguish between invalid strings and unsupported types

```go
id, err := conversion.ToUUID("invalid-uuid")
if err != nil {
    // This was a string but invalid format
    log.Printf("Invalid UUID format: %v", err)
} else if id == uuid.Nil {
    // This might be nil input or unsupported type (no error)
    log.Printf("No valid UUID provided")
}
```

## Testing

The package includes comprehensive unit tests covering:
- All supported input types
- Error cases with invalid UUID strings
- Nil input handling
- Type conversion accuracy

Run tests with:
```bash
go test ./conversion
```

## Dependencies

- `github.com/google/uuid`: UUID type definitions and parsing functions

## License

MIT
