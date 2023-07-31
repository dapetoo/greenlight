package data

import (
	"errors"
	"time"
)
import "golang.org/x/crypto/bcrypt"

//User struct to represent an individual user.

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// Custom password type containing the hashed and the plaintext.
type password struct {
	plaintext *string
	hash      []byte
}

// Set method calculates the bcrypt hash of a plaintext password and stores both the hash anf the plaintext versions
// in the struct
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// Matches method checks the provided plaintext password matches the hashed password stored in the struct, returning true if
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err

		}
	}
	return true, nil

}
