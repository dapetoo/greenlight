package main

import (
	"errors"
	"github.com/dapetoo/greenlight/internal/data"
	"github.com/dapetoo/greenlight/internal/validator"
	"net/http"
	"time"
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

	//After the user has been created in the DB, generate a new activation token for the user.
	token, err := app.models.Tokens.New(user.ID, 8*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	//Use the helper method to execute an anonymous function that sends welcome email
	app.background(func() {

		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}
		//Call Send() to send email passing in the template and user struct data
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}
	})

	//JSON response to the client with the user data and a 201 status code
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Activate users with the token
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	//Parse the plaintext application token from the request body
	var input struct {
		TokenPlainText string `json:"token"`
	}

	err := app.readJSON(w, r, input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	//Validate the plaintext tojen provided by the client
	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlainText); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//Retrieve the details of the user associated with the token using the GetForToken() method. If no matching record
	//found, return invalid token provided.
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//Update the user's activation status
	user.Activated = true

	//Save the updated user record in the DB, check for any edit conflicts
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	//If everything went well, delete all activation tokens for the user
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//Send the updated user details to the client in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
