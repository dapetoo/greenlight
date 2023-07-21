package main

import "net/http"

func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

// errorResponse() to send JSON-formatted error message to the client with a given status code
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	//Write the error using the writeJSON helper method
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
