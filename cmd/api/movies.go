package main

import (
	"errors"
	"fmt"
	"github.com/dapetoo/greenlight/internal/data"
	"github.com/dapetoo/greenlight/internal/validator"
	"net/http"
	"reflect"
	"strconv"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	//Initialize a new json.Decoder which reads the request body
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	// Call the ValidateMovie() function and return a response containing the errors if any of the checks fail
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//Call the Insert() on Movies model passing in a pointer to the validated movie struct
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	//Send JSON response with a 201 created status code, the movie data in the response body and the location header
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, http.Header{})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// UpdateMovieHandler update a movie
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	//Extract the movie ID from the URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	//Fetch the existing movie records from the database
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//If the request contains a X-Expected-Version header, verify that the movie version in the DB matches the version
	//specified in the header
	expectedVersion := r.Header.Get("X-Expected-Version")
	if expectedVersion != "" && strconv.Itoa(int(movie.Version)) != expectedVersion {
		app.editConflictResponse(w, r)
		return
	}

	//Declare an input struct to hold the expected data from the client
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	//Read the JSON request body into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	vMovie := reflect.ValueOf(movie).Elem()
	vInput := reflect.ValueOf(input)

	for i := 0; i < vInput.NumField(); i++ {
		inputField := vInput.Field(i)
		if inputField.IsValid() && !inputField.IsNil() {
			movieField := vMovie.FieldByName(vInput.Type().Field(i).Name)
			if movieField.IsValid() && movieField.CanSet() {
				movieField.Set(inputField.Elem())
			}
		}
	}

	//Validate the updated movie record, send a 422 response if any check fail
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated movie record to the Update method
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//Write the updated movie record in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// DeleteMovieHandler to delete movie
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	//Delete the movie from the database, send a 404 response to the client if there's not a matching record
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//Return a 200 OK Status code along with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	//Declare an input struct to hold the expected data from the client
	var input struct {
		Title    string
		Genres   []string
		Page     int
		PageSize int
		Sort     string
	}

	v := validator.New()

	//Get the url.Values map containing the query string data
	qs := r.URL.Query()

	//Extract the title and genres string values
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	//Get the page and page size query string values as integers
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 20, v)

	//Extract the sort query string value
	input.Sort = app.readString(qs, "sort", "id")

	//Check the validator instance for any errors and use the failedValidationResponse()
	//helper to send the client a response if necessary.
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
	}

	fmt.Fprintf(w, "%+v\n", input)
}
