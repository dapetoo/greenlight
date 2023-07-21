package main

import (
	"fmt"
	"greenlight/internal/data"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	movie := &data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Movies",
		Year:      2023,
		Runtime:   12,
		Genres:    []string{"drama"},
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, http.Header{})
	if err != nil {
		app.logger.Printf("the error is %s", err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
