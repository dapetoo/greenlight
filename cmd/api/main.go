package main

import (
	"context"
	"database/sql"
	"flag"
	"github.com/dapetoo/greenlight/internal/data"
	"github.com/dapetoo/greenlight/internal/jsonlog"
	"github.com/dapetoo/greenlight/internal/mailer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"os"
	"time"
)

const version = "1.0.0"

// Config Struct to hold configuration settings
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// Application struct to hold the dependencies for HTTP handlers, helpers and the middlewares
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
}

func init() {
	err := godotenv.Load()
	if err != nil {
		zlog.Fatal().Msg("Error loading .env file")
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var cfg config

	//read the value of the port and env command-line flags into the config struct
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgresSQL DSN ")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL Max Open Connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL Max Idle Connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL Max Idle Time")

	//Rate Limiter config
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable Rate Limiting")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter burst", 4, "Rate limiter maximum burst")

	//SMTP Server configuration settings
	flag.StringVar(&cfg.smtp.host, "smtp host", os.Getenv("MAIL_SERVER"), "SMTP Host")
	flag.IntVar(&cfg.smtp.port, "smtp port", 2525, "SMTP Port")
	flag.StringVar(&cfg.smtp.username, "smtp username", os.Getenv("MAIL_USERNAME"), "SMTP Username")
	flag.StringVar(&cfg.smtp.password, "smtp password", os.Getenv("MAIL_PASSWORD"), "SMTP Password")
	flag.StringVar(&cfg.smtp.sender, "smtp sender", os.Getenv("MAIL_SENDER"), "SMTP Sender")
	flag.Parse()

	//Init a new logger to write message to STDOUT
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	//Create a connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
		zlog.Fatal().Err(err)
	}

	//Defer call to db.Close()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.PrintFatal(err, nil)
			zlog.Fatal().Err(err)
		}
	}(db)

	logger.PrintInfo("database connection pool established successfully", nil)
	zlog.Info().Msg("database connection pool established successfully")

	//Declare an instance of the application struct, containing the config anf the logger
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.serve()
	logger.PrintFatal(err, nil)

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	//Set connection pool
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//PingContext() to establish a new connection to the database
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
