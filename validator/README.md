# Validator

Centralized validator for Go structs, using [go-playground/validator](https://github.com/go-playground/validator). It validates data and formats errors in a friendly way for APIs.

## Installation

```bash
go get github.com/juanMaAV92/go-utils/validator
```

Add the dependency to your `go.mod`:

```
go get github.com/juanMaAV92/go-utils/validation
```

## Basic Usage

```go
import (
    "github.com/juanMaAV92/go-utils/validator"
)

// Create the singleton instance (only once)
_ = validator.New()
validator := validator.GetInstance()

// Validate a struct
err := validator.Validate(&myStruct)
if err != nil {
    // Handle validation errors
}
```

## Bind and Validate

If you use frameworks like Echo, you can combine binding and validation:

```go
err := validation.BindAndValidate(ctx, &req)
if err != nil {
    // Handle error
}
```

## Error Customization

Errors are automatically formatted according to the validation type (`required`, `min`, `max`, `email`, `uuid`, etc.) and returned in an API-friendly format.

## Supported Validations
- required
- min / max / len / gt / lt / gte / lte
- email
- uuid
- oneof
- numeric / alpha / alphanum
- url
- datetime

## Example Struct

```go
type User struct {
    Name     string `validate:"required,min=3,max=50"`
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}
```

## License

MIT
