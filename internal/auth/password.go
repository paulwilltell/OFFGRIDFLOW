// Package auth - password.go provides secure password hashing using bcrypt.
package auth

import (
"errors"
"unicode"

"golang.org/x/crypto/bcrypt"
)

// -----------------------------------------------------------------------------
// Password Policy
// -----------------------------------------------------------------------------

// Password policy constants.
const (
// MinPasswordLength is the minimum allowed password length.
MinPasswordLength = 8

// MaxPasswordLength is the maximum allowed password length.
// bcrypt has a 72-byte limit; we use a lower value for safety.
MaxPasswordLength = 64

// DefaultBcryptCost is the bcrypt work factor (2^12 iterations).
DefaultBcryptCost = bcrypt.DefaultCost
)

// Password policy errors.
var (
// ErrPasswordTooShort indicates the password is below minimum length.
ErrPasswordTooShort = errors.New("auth: password must be at least 8 characters")

// ErrPasswordTooLong indicates the password exceeds maximum length.
ErrPasswordTooLong = errors.New("auth: password must not exceed 64 characters")

// ErrPasswordNoUpper indicates the password lacks an uppercase letter.
ErrPasswordNoUpper = errors.New("auth: password must contain an uppercase letter")

// ErrPasswordNoLower indicates the password lacks a lowercase letter.
ErrPasswordNoLower = errors.New("auth: password must contain a lowercase letter")

// ErrPasswordNoDigit indicates the password lacks a digit.
ErrPasswordNoDigit = errors.New("auth: password must contain a digit")

// ErrPasswordNoSpecial indicates the password lacks a special character.
ErrPasswordNoSpecial = errors.New("auth: password must contain a special character")

// ErrPasswordEmpty indicates an empty password was provided.
ErrPasswordEmpty = errors.New("auth: password cannot be empty")
)

// -----------------------------------------------------------------------------
// Password Validation
// -----------------------------------------------------------------------------

// PasswordStrength represents the complexity requirements for passwords.
type PasswordStrength int

const (
// PasswordStrengthBasic requires only minimum length.
PasswordStrengthBasic PasswordStrength = iota

// PasswordStrengthMedium requires length + mixed case.
PasswordStrengthMedium

// PasswordStrengthStrong requires length + mixed case + digits + special.
PasswordStrengthStrong
)

// ValidatePassword checks if a password meets the specified strength requirements.
// Returns nil if valid, or an error describing the first failing requirement.
func ValidatePassword(password string, strength PasswordStrength) error {
if password == "" {
return ErrPasswordEmpty
}

if len(password) < MinPasswordLength {
return ErrPasswordTooShort
}

if len(password) > MaxPasswordLength {
return ErrPasswordTooLong
}

if strength == PasswordStrengthBasic {
return nil
}

// Check for uppercase and lowercase
var hasUpper, hasLower, hasDigit, hasSpecial bool
for _, r := range password {
switch {
case unicode.IsUpper(r):
hasUpper = true
case unicode.IsLower(r):
hasLower = true
case unicode.IsDigit(r):
hasDigit = true
case unicode.IsPunct(r) || unicode.IsSymbol(r):
hasSpecial = true
}
}

if !hasUpper {
return ErrPasswordNoUpper
}
if !hasLower {
return ErrPasswordNoLower
}

if strength == PasswordStrengthMedium {
return nil
}

// PasswordStrengthStrong requires digits and special characters
if !hasDigit {
return ErrPasswordNoDigit
}
if !hasSpecial {
return ErrPasswordNoSpecial
}

return nil
}

// ValidatePasswordBasic validates a password with basic requirements (length only).
func ValidatePasswordBasic(password string) error {
return ValidatePassword(password, PasswordStrengthBasic)
}

// ValidatePasswordStrong validates a password with strong requirements.
func ValidatePasswordStrong(password string) error {
return ValidatePassword(password, PasswordStrengthStrong)
}

// -----------------------------------------------------------------------------
// Password Hashing
// -----------------------------------------------------------------------------

// HashPassword creates a bcrypt hash of the plaintext password.
// Uses the default cost factor (currently 10, which is 2^10 iterations).
//
// Example:
//
//hash, err := auth.HashPassword("secretPassword123!")
//if err != nil {
//    // Handle error
//}
//// Store hash in database
func HashPassword(plaintext string) (string, error) {
if plaintext == "" {
return "", ErrPasswordEmpty
}

hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), DefaultBcryptCost)
if err != nil {
return "", err
}
return string(hashed), nil
}

// HashPasswordWithCost creates a bcrypt hash with a custom cost factor.
// Higher costs are more secure but slower. Valid range is 4-31.
//
// Recommended costs:
//   - 10: Default, suitable for most applications
//   - 12: Higher security for sensitive systems
//   - 14+: Very high security (may be slow)
func HashPasswordWithCost(plaintext string, cost int) (string, error) {
if plaintext == "" {
return "", ErrPasswordEmpty
}

if cost < bcrypt.MinCost {
cost = bcrypt.MinCost
}
if cost > bcrypt.MaxCost {
cost = bcrypt.MaxCost
}

hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), cost)
if err != nil {
return "", err
}
return string(hashed), nil
}

// CheckPassword verifies that a plaintext password matches the stored hash.
// Returns true if the password is correct, false otherwise.
//
// Example:
//
//if auth.CheckPassword(storedHash, inputPassword) {
//    // Password is correct
//}
func CheckPassword(hash, plaintext string) bool {
if hash == "" || plaintext == "" {
return false
}
return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// NeedsRehash checks if a password hash should be upgraded to a higher cost.
// Returns true if the hash was created with a lower cost than specified.
//
// Use this during login to gradually upgrade password hashes:
//
//if auth.NeedsRehash(storedHash, 12) {
//    newHash, _ := auth.HashPasswordWithCost(plaintext, 12)
//    // Update stored hash
//}
func NeedsRehash(hash string, desiredCost int) bool {
cost, err := bcrypt.Cost([]byte(hash))
if err != nil {
return true // If we can't determine cost, rehash to be safe
}
return cost < desiredCost
}

// GetHashCost returns the bcrypt cost factor used to generate a hash.
// Returns -1 if the hash is invalid.
func GetHashCost(hash string) int {
cost, err := bcrypt.Cost([]byte(hash))
if err != nil {
return -1
}
return cost
}
