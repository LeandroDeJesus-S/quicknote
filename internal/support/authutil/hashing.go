// Package authutil provides authentication utilities.
package authutil

import "golang.org/x/crypto/bcrypt"

// PasswordHasher is an interface for hashing and verifying passwords.
type PasswordHasher interface {
	// HashPassword hashes a password using a secure algorithm.
	HashPassword(password string) (string, error)
	// CheckPassword compares a plaintext password with a hash to see if they match.
	CheckPassword(pw, hash string) (bool, error)
}

// bcryptHasher is a PasswordHasher implementation that uses the bcrypt algorithm.
type bcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new bcryptHasher with the given cost.
// The cost parameter determines the complexity of the hashing algorithm.
func NewBcryptHasher(opts ...HasherOpt) PasswordHasher {
	h := &bcryptHasher{cost: bcrypt.DefaultCost}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			panic(err)
		}
	}

	return h
}

// HashPassword hashes a password using bcrypt.
func (p *bcryptHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	return string(bytes), err
}

// CheckPassword compares a plaintext password with a bcrypt hash.
// It returns true if the password matches the hash, and false otherwise.
func (p *bcryptHasher) CheckPassword(pw, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
	return err == nil, err
}

type HasherOpt func(h PasswordHasher) error

func WithCost(cost int) HasherOpt {
	return func(h PasswordHasher) error {
		switch hasher := h.(type) {
		case *bcryptHasher:
			hasher.cost = cost
		}
		return nil
	}
}
