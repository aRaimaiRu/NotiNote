package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher implements password hashing using bcrypt
type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher creates a new bcrypt password hasher
func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{
		cost: bcrypt.DefaultCost, // Cost 10
	}
}

// HashPassword hashes a plain text password using bcrypt
func (h *BcryptPasswordHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a plain text password with a bcrypt hash
func (h *BcryptPasswordHasher) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
