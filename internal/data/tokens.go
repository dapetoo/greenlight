package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"github.com/dapetoo/greenlight/internal/validator"
	"time"
)

// ScopeActivation Token scope for different use
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

// Token struct for an individual tokens.
type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"scope"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	//Create a token instance containing the UserID and all other information
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	//Init a zero-valued byte slice with a lenght of 16bytes
	randomBytes := make([]byte, 16)

	//Read() to fill the byte slice with random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	//Encode the byte slice to a base-32 encoded string and assign it to the token Plaintext field.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	//Generate a SHA-256 hash of the plaintext token string. this will be stored in the hash field of the DB table
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

// ValidateTokenPlaintext Check that the Plaintext token has been provided and is exactly 26 bytes long
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

// TokenModel Define the TokenModel type
type TokenModel struct {
	DB *sql.DB
}

// New creates a new token struct and then inserts the data in the tokens table
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert adds the data for a specific token to the tokens table
func (m TokenModel) Insert(token *Token) error {
	query := `
			INSERT INTO tokens (hash, user_id, expiry, scope)
			VALUES ($1, $2, $3, $4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAllForUser deletes all tokens for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens 
		WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
