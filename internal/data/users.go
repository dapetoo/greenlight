package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dapetoo/greenlight/internal/validator"
	"time"
)
import "golang.org/x/crypto/bcrypt"

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// User struct to represent an individual user.
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

// UserModel struct
type UserModel struct {
	DB *sql.DB
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

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	//Check if password is not nil and validate the password
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	//Check if the password hash is nil.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// Insert a new record into the DB for the user
func (m UserModel) Insert(user *User) error {
	query := `
			INSERT INTO users (name, email, password_hash, activated)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, version
			`

	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//If the table already contains a record with this email address, return ErrDuplicateEmail error
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID, &user.CreatedAt, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err

		}
	}
	return nil
}
