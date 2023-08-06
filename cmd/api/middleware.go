package main

import (
	"errors"
	"expvar"
	"fmt"
	"github.com/dapetoo/greenlight/internal/data"
	"github.com/dapetoo/greenlight/internal/validator"
	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				//If there's a panic, set a "Connection: close" header on the response. This automatically close the
				//current connection after a response has been sent
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {

	//Define a client struct to hold the rate limiter and last seen for each client
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	//Declare a mutex and a map to hold the client's IP addresses and rate limitters
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	//Launch a background goroutine which removes old entries from the clients map once every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			//Lock the mutex to prevent any rate limiter checks from happening while the cleanup is happening
			mu.Lock()

			//Loop through all clients. If not seen within the last 3 minutes, delete the corresponding entry
			//from the map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			//Unlock the mutex when the cleanup is complete
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			//Use the realip.FromRequest() function to get the client's real IP address
			ip := realip.FromRequest(r)

			//Lock the mutex to prevent this code from being executed concurrently
			mu.Lock()

			//Check to see if the IP address already exists in the map. If it doesnt, init a new rate limiter
			// and add the IP address and limiter to the map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			//Update the last seen time for the client
			clients[ip].lastSeen = time.Now()

			//Call limiter.Allow() on the rate limiter for the current's IP address. If the request is nt allowed,
			// unlock the mutex and send a 429 Too many requests response
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			//Unlock the mutex before calling the next handler in the chain.
			mu.Unlock()

		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Add the "Vary: Authentication" header to the response. This indicates to any caches that the response
		//may vary based on the value of the Authorization header in the request.
		w.Header().Add("Vary", "Authorization")

		//Retrive the value of the Authorization Header from the request. This will return the empty string "", if there
		// is no such header found
		authorizationHeader := r.Header.Get("Authorization")

		//If there's no Authorization found, use the contextSetUser() helper to add the AnonymousUser to the request
		// context.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		//Otherwise, we expect the value of the Authorization header to be in the format "Bearer <token>".
		// Split this into the constituents part and if the header isnt in the expected format, return 401 response
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		//Extract the actual authentication token from the header parts
		token := headerParts[1]

		//Validate the token to make sure it's in a sensible format
		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		//Retrieve the details of the user associated with the authentication token.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
				return
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		//Call the contextSetUser() helper method to add the user information to the request context
		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		//Check that a user is activated.
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
	})

	//Wrap fn with the requiredAuthenticatedUser middleware before returning it.
	return app.requireAuthenticatedUser(fn)
}

// Check that a user is not anonymous
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Permission middleware
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		//Retrieve the user from the request context
		user := app.contextGetUser(r)

		//Get the slice of permissions for the user
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		//Check if the slice includes the required permission.
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
	return app.requireActivatedUser(fn)

}

// enableCORs set AccessControlAllowOrigin header
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")

		//Get the value of the request's Origin header
		origin := r.Header.Get("Origin")

		//Run this if there's an Origin request header present
		if origin != "" {
			//Loop through the listed of trusted origins, checking to see if request origin exactly matches
			//one of them. if no match, the loop wont be iterated
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					//If there's a match, set a "Access-Control-Allow-Origin" response header with the request
					// origin as the value and break out of the loop
					w.Header().Set("Access-Control-Allow-Origin", origin)

					//Check if the request has the HTTP method OPTIONS and contains the
					// "Access-Control-Request-Method" header. If it does, treas as a preflight request
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						//Set the necessary preflight responses headers
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.Header().Set("Access-Control-Max-Age", "-1")

						//Write the headers along with a 200 OK status code
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Expose Request-Level Metrics
func (app *application) metrics(next http.Handler) http.Handler {
	//Initialize the new expvar variables when the middleware chain is first built
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_µs")
	totalResponsesSentByStatus := expvar.NewMap("total_response_sent_by_status")

	//The code will be run for every requests
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Use the ADD() method to increment the number of requests received by 1
		totalRequestsReceived.Add(1)

		//Call the httpsnoop.CaptureMetrics function, passing in the next handler in the chain along with existing
		// http.ResponseWriter and http.Request. This returns the metrics struct
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		//Use the ADD() method to increment the number of requests received by 1
		totalResponsesSent.Add(1)

		//Get the request processing time in microseconds from httpsnoop and increment the cumulative processing time
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())

		//Use the Add() to increment the count for the given status code by 1.
		// Convert the expvar map string-keyed to a string from the integer using strconv.Itoa()
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
