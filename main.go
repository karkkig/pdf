package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"your_module/internal/handler"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "UI")

	e.POST("/generate-pdf", handler.GeneratePDF)

	e.Logger.Fatal(e.Start(":8080"))
}
