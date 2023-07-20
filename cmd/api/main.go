package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

// Config Struct to hold configuration settings
type config struct {
	port int
	env  string
}

// Application struct to hold the dependencies for HTTP handlers, helpers and the middlewares
type application struct {
	config config
	logger *log.Logger
}

func main() {

	var cfg config

	//read the value of the port and env command-line flags into the config struct
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	//Init a new logger to write message to STDOUT
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	//Declare an instance of the application struct, containing the config anf the logger
	app := &application{
		config: cfg,
		logger: logger,
	}

	//Declare HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	//Start the HTTP Server
	logger.Printf("starting %s server on %s", cfg.env, cfg.port)
	err := srv.ListenAndServe()
	logger.Fatal(err)

}
