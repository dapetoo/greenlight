package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	//Initialize a new httpRouter instance
	router := httprouter.New()

	//Custom notfound response for http.handler
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	//Register the methods using HandlerFunc
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)
	//Require authenticated user
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requireActivatedUser())
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requireActivatedUser())
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requireActivatedUser())
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requireActivatedUser())
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requireActivatedUser())

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
