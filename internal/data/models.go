package data

import (
	"database/sql"
	"errors"
)

// /Custom Error Implementation
var (
	ErrRecordNotFound = errors.New("models: no matching record found")
)

type Models struct {
	Movies MovieModel
}

// NewModels returns a Models struct containing the init MovieModel
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{
			DB: db,
		},
	}
}
