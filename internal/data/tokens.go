package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
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
	randowmBytes := make([]byte, 16)

	//Read() to fill the byte slice with random bytes
	_, err := rand.Read(randowmBytes)
	if err != nil {
		return nil, err
	}

	//Encode the byte slice to a base-32 encoded string and assign it to the token Plaintext field.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randowmBytes)

	//Generate a SHA-256 hash of the plaintext token string. this will be stored in the hash field of the DB table
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}
