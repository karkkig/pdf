package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/karkki-hub/chromedp_pdfgen/chromedp"
)

func main() {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogMethod:    true,
		LogLatency:   true,
		LogRequestID: true,
		LogError:     true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			if v.Error != nil || v.Status >= 500 {
				level = slog.LevelError
			} else if v.Status >= 400 {
				level = slog.LevelWarn
			}
			slog.Log(c.Request().Context(), level, "request",
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
				"request_id", v.RequestID,
				"error", v.Error,
			)
			return nil
		},
	}))

	e.Static("/", "UI")
	e.POST("/v1/generatepdf", chromedp.GenerateHandler)
	e.GET("/health", chromedp.HealthHandler)

	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		slog.Error("shutting down server", "error", err)
		os.Exit(1)
	}
}
