package data

import (
	"database/sql"
	"errors"
)

// ErrRecordNotFound Custom Error Implementation
var (
	ErrRecordNotFound = errors.New("models: no matching record found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
		GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error)
	}
	Users UserModel
}

// NewModels returns a Models struct containing the init MovieModel
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: &MovieModel{
			DB: db,
		},
		Users: UserModel{
			DB: db,
		},
	}
}

// NewMockModels returns a Models instance containing the mock models
func NewMockModels() Models {
	return Models{
		Movies: &MockMovieModel{},
	}
}
