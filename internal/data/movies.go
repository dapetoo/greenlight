package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dapetoo/greenlight/internal/validator"
	"github.com/lib/pq"
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

type MockMovieModel struct{}

// Insert a new record into the movies table
func (m *MovieModel) Insert(movie *Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
			INSERT INTO movies (title, year, runtime, genres)
			VALUES ($1, $2, $3, $4) 
			RETURNING id, created_at, version
			`
	//args slice containing the values for the placeholder parameters from the movie struct. Declaring this slice immediately
	//makes it nice and clear *what values are being used where* in the query.
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get method for fetching a specific record from the movies table
func (m *MovieModel) Get(id int64) (*Movie, error) {
	//Check if there is no record in the DB
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
			SELECT id, created_at, title, year, runtime, genres, version
			FROM movies
			WHERE id = $1;
			`

	//Init a pointer to the movie
	var movie Movie

	row := m.DB.QueryRowContext(ctx, stmt, id)

	err := row.Scan(
		&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

// Update method update a specific record in the movies table
func (m *MovieModel) Update(movie *Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
			UPDATE movies
			SET title = $1, year = $2, runtime = $3, genres = $4, version = version +1
			WHERE id = $5 AND version = $6
			RETURNING version;
			`
	args := []interface{}{
		movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version,
	}

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Delete method delete a specific record in the movies table
func (m *MovieModel) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
			DELETE FROM movies
			WHERE id = $1
			`
	result, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	//Get the number of Rows() affected by the query
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	//Check if no rows were affected and return no record found error
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// GetAll() to return a slice of movies
func (m *MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
			SELECT id, created_at, title, year, runtime, genres, version
			FROM movies
			ORDER BY id
			`
	//QueryContext to execute the query
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	//Initialize an empty slice to hold the movie data
	var movies []*Movie
	for rows.Next() {
		var movie Movie
		//Scan the values from the row into the movie struct
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)

		if err != nil {
			return nil, err
		}
		//Add the movie struct to the slice
		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}

// Insert a new record into the movies table
func (m *MockMovieModel) Insert(movie *Movie) error {
	return nil
}

// Get method for fetching a specific record from the movies table
func (m *MockMovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Update method update a specific record in the movies table
func (m *MockMovieModel) Update(movie *Movie) error {
	return nil
}

// Delete method delete a specific record in the movies table
func (m *MockMovieModel) Delete(id int64) error {
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
