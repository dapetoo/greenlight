package main

import (
	"errors"
	"fmt"
	"github.com/dapetoo/greenlight/internal/data"
	"github.com/dapetoo/greenlight/internal/validator"
	"golang.org/x/time/rate"
	"net"
	"net/http"
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
		//Extract the client's IP address from the request
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		//Lock the mutex to prevent this code from being executed concurrently
		mu.Lock()

		//Check to see if the IP address already exists in the map. If it doesnt, init a new rate limiter
		// and add the IP address and limiter to the map
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
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
