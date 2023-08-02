package main

import (
	"context"
	"errors"
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
	shutdownError := make(chan error)

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
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		//Create a context with a 5 second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		//Call shutdown on the server, passing in the context
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		//Log a message to say that we're waiting for any background goroutines to complete their tasks
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		//Call Wait() to block until out WaitGroup counter is zero, essentially blocking until the background
		//goroutines have finished. Return nil on the shutdownError channel to indicate shutdown completed without any issues
		app.wg.Wait()
		shutdownError <- nil
	}()

	//Log a starting server message
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	//Calling Shutdown() on the server will cause ListenAndServe() to immediately return a http.ErrServerClosed error.
	//Return the error if it's not http.ErrServerClosed.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	//Otherwise, wait to receive return value from Shutdown() on the shutdownError channel. If return value is an error,
	//return the error
	err = <-shutdownError
	if err != nil {
		return err
	}

	//Graceful shutdown completed successfully, and we logged a stopped server message
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
