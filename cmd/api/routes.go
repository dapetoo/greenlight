package main

import (
	"expvar"
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
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	//Endpoint to the expvar handler
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	//return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
	return app.recoverPanic(app.enableCORS(app.authenticate(router)))
}
