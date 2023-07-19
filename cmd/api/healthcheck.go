package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func healthCheckHandler(c echo.Context) error {
	status := "status: available"
	environment := "environment: dev"
	return c.String(http.StatusOK, environment+status)
}
