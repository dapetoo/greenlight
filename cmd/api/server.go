package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	//Declare HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%v", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	//Shutdown Error Channel to receive any errors returned by the graceful Shutdown() function

	//Start a background goroutine
	go func() {
		//Create a quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)

		//Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals anf=d relay them to the quit channel
		// Any other signals will not be caught and will retain their default behaviours
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		//Read the signal from the quit channel, the code will be block until a signal is received
		s := <-quit

		//Log a message to notify that the signal has been caught
		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		//Exit the application with a 0 (success) status code
		os.Exit(0)
	}()

	//Log a starting server message
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})
	return srv.ListenAndServe()
}
