package main

import (
	"fmt"
	"github.com/cajun-pro-llc/open-match/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"net/http"
	"os"
)

func main() {
	utils.LogEnv()
	log.Info().Msg("Starting service")

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(&Context{c})
		}
	})
	echoLogger := zerolog.New(os.Stdout)
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			echoLogger.Info().
				Str("URI", v.URI).
				Int("status", v.Status).
				Msg("request")

			return nil
		},
	}))
	e.Use(middleware.RequestID())
	e.Use(middleware.Decompress())
	e.IPExtractor = echo.ExtractIPDirect()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Open Match Frontend")
	})
	v1 := e.Group("/v1")
	tickets := v1.Group("/tickets")
	tickets.POST("", createTicket)
	tickets.GET("/:ticketId", getTicket)
	tickets.DELETE("/:ticketId", deleteTicket)

	// Serve on the edgegap environment variable defined port
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("HTTP_SERVE_PORT"))))
}
