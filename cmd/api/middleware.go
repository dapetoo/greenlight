package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
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
	//Declare a mutex and a map to hold the client's IP addresses and rate limitters
	var (
		mu      sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)

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
			clients[ip] = rate.NewLimiter(2, 4)
		}
		//Call limiter.Allow() on the rate limiter for the current's IP address. If the request is nt allowed,
		// unlock the mutex and send a 429 Too many requests response
		if !clients[ip].Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		//Unlock the mutex before calling the next handler in the chain.
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
