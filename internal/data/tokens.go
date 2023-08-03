package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"github.com/dapetoo/greenlight/internal/validator"
	"time"
)

// Token scope for different use
const (
	ScopeActivation = "activation"
)

// Token struct for an individual tokens.
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
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
