package data

import (
	"database/sql"
	"github.com/dapetoo/greenlight/internal/validator"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

// MovieModel struct which wraps a sql.DB connection pool
type MovieModel struct {
	DB *sql.DB
}

// Insert a new record into the movies table
func (m *MovieModel) Insert(movie *Movie) error {
	return nil
}

// Get method for fetching a specific record from the movies table
func (m *MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Update method update a specific record in the movies table
func (m *MovieModel) Update(movie *Movie) error {
	return nil
}

// Delete method delete a specific record in the movies table
func (m *MovieModel) Delete(id int64) error {
	return nil
}

// ValidateMovie runs validation checks on the Movie type.
func ValidateMovie(v *validator.Validator, movie *Movie) {
	// Check movie.Title
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	// Check movie.Year
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	// Check movie.Runtime
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	// Check movie.Genres
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}