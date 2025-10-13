# Crypto Package

This package provides secure password hashing and validation utilities using bcrypt with salt generation for enhanced security in Go applications.

## Features

- **Secure Salt Generation**: Generate cryptographically secure random salts
- **Password Hashing**: Hash passwords using bcrypt with configurable cost
- **Password Validation**: Verify passwords against stored hashes
- **Security Best Practices**: Uses recommended bcrypt cost factor and salt length

## Constants

- `SaltLength = 16`: Length of generated salt in bytes (32 hex characters)
- `BCryptCost = 12`: Bcrypt cost factor for password hashing (recommended security level)

## Main Functions

- `GeneratePasswordSalt() (string, error)`: Generate a cryptographically secure random salt
- `HashPassword(password, salt string) (string, error)`: Hash a password with salt using bcrypt
- `ValidatePassword(password, salt, hash string) bool`: Validate a password against its hash

## Usage Example

```go
import "github.com/juanMaAV92/go-utils/crypto"

// User registration flow
func RegisterUser(password string) error {
    // Generate a unique salt for this user
    salt, err := crypto.GeneratePasswordSalt()
    if err != nil {
        return fmt.Errorf("failed to generate salt: %w", err)
    }
    
    // Hash the password with the salt
    hashedPassword, err := crypto.HashPassword(password, salt)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }
    
    // Store user with salt and hashed password
    user := User{
        Email:    email,
        Salt:     salt,
        Password: hashedPassword,
    }
    
    return userRepo.Create(user)
}

// User login flow
func LoginUser(email, password string) (bool, error) {
    // Retrieve user from database
    user, err := userRepo.GetByEmail(email)
    if err != nil {
        return false, err
    }
    
    // Validate the password
    isValid := crypto.ValidatePassword(password, user.Salt, user.Password)
    return isValid, nil
}
```

## Security Features

### Salt Generation
- Uses `crypto/rand` for cryptographically secure random number generation
- 16-byte salt provides sufficient entropy (2^128 possible values)
- Each user gets a unique salt to prevent rainbow table attacks

### Password Hashing
- Uses bcrypt algorithm with cost factor 12 (recommended as of 2023)
- Salt is prepended to password before hashing
- Bcrypt handles additional internal salting and iterations

### Password Validation
- Constant-time comparison prevents timing attacks
- Reconstructs salted password and compares using bcrypt's built-in verification

## Complete Registration/Login Example

```go
package main

import (
    "fmt"
    "github.com/juanMaAV92/go-utils/crypto"
)

type User struct {
    ID       int
    Email    string
    Salt     string
    Password string
}

func main() {
    // Registration
    plainPassword := "mySecurePassword123!"
    
    salt, err := crypto.GeneratePasswordSalt()
    if err != nil {
        panic(err)
    }
    
    hashedPassword, err := crypto.HashPassword(plainPassword, salt)
    if err != nil {
        panic(err)
    }
    
    user := User{
        Email:    "user@example.com",
        Salt:     salt,
        Password: hashedPassword,
    }
    
    fmt.Printf("User registered with salt: %s\n", salt)
    
    // Login validation
    loginPassword := "mySecurePassword123!"
    isValid := crypto.ValidatePassword(loginPassword, user.Salt, user.Password)
    
    if isValid {
        fmt.Println("Login successful!")
    } else {
        fmt.Println("Invalid credentials!")
    }
}
```

## Error Handling

The package returns errors in the following cases:
- `GeneratePasswordSalt()`: When the system's random number generator fails
- `HashPassword()`: When bcrypt hashing fails (typically due to memory constraints or invalid cost)

`ValidatePassword()` returns a boolean and never errors - invalid passwords simply return `false`.

## Testing

The package includes unit tests that verify:
- Salt generation produces unique values
- Password hashing works correctly
- Password validation correctly identifies valid and invalid passwords

Run tests with:
```bash
go test ./crypto
```

## Dependencies

- `golang.org/x/crypto/bcrypt`: Bcrypt password hashing implementation
- `crypto/rand`: Cryptographically secure random number generation
- `encoding/hex`: Hexadecimal encoding for salt representation

## Security Considerations

1. **Cost Factor**: The bcrypt cost of 12 provides good security as of 2023. Consider increasing it in the future as hardware improves.
2. **Salt Storage**: Always store the salt alongside the hashed password in your database.
3. **Password Requirements**: This package handles hashing securely, but implement password complexity requirements in your application logic.
4. **Timing Attacks**: The validation function uses bcrypt's constant-time comparison internally.

## License

MIT
