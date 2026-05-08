package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "modernc.org/sqlite"

	"gouv.viz/internal/store"
	"gouv.viz/web/handlers"
)

func main() {
	handlers.Init()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))

	e.Static("/assets", os.Getenv("ASSETS_PATH"))

	db, err := sql.Open("sqlite", os.Getenv("DATABASE_PATH"))
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	server := handlers.NewServer(store.New(db))

	e.GET("/", server.Home)
	e.GET("/scrutins", server.Scrutins)
	e.GET("/scrutins/:uid", server.ScrutinDetail)
	e.GET("/ping", handlers.Ping)

	if os.Getenv("ENV") != "prod" {
		e.GET("/ws", handlers.HotReloadWS)
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("PORT"))))
}
