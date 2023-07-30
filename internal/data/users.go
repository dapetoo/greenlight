package data

import "time"

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
