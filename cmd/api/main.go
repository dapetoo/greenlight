package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
	db   struct{ dsn string }
}

// Application struct to hold the dependencies for HTTP handlers, helpers and the middlewares
type application struct {
	config config
	logger *log.Logger
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	var cfg config

	//read the value of the port and env command-line flags into the config struct
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgresSQL DSN ")
	flag.Parse()

	//Init a new logger to write message to STDOUT
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	//Create a connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	//Defer call to db.Close()
	defer db.Close()

	logger.Printf("database connection pool established successfully")

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
	logger.Printf("starting %s server on port %d", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//PingContext() to establish a new connection to the database
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
