package domain

import "errors"

// User represents a user in the system.
// It contains no tags for serialization, keeping it pure.
type User struct {
	ID    string
	Email string
}

// NewUser is a constructor for the User entity.
func NewUser(id string) (*User, error) {
	if id == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return &User{ID: id}, nil
}