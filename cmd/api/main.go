package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
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

	fmt.Println("Hello World")

	mux := echo.New()
	mux.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://labstack.com", "https://labstack.net"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	mux.GET("/v1/healthcheck", healthCheckHandler)

	log.Println("Startng server on port 8000")
	mux.Logger.Fatal(mux.Start(":8000"))

}
