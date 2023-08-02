package main

import (
	"errors"
	"github.com/dapetoo/greenlight/internal/data"
	"github.com/dapetoo/greenlight/internal/validator"
	"net/http"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	//Anonymous struct
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	//Parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//Copy the data from the request body into a new user struct.
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	//Use the Password.Set() to generate and stored hashed nad plaintext passwords
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	//Validate the user struct and return error messages if any of the check fail
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//Insert the user data into the DB
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//Call Send() to send email passing in the template and user struct data
	//_, err = os.ReadFile("./templates/user_welcome.tmpl")
	//if err != nil {
	//	log.Println("Unable to read file")
	//}
	err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//JSON response to the client with the user data and a 201 status code
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
