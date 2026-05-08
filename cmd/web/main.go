package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gouv.viz/web/handlers"
)

func main() {
	handlers.Init()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))

	e.Static("/assets", os.Getenv("ASSETS_PATH"))

	e.GET("/", handlers.Home)
	e.GET("/scrutins", handlers.Scrutins)
	e.GET("/ping", handlers.Ping)

	if os.Getenv("ENV") != "prod" {
		e.GET("/ws", handlers.HotReloadWS)
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("PORT"))))
}
