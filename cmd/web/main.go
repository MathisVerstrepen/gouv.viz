package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "modernc.org/sqlite"

	"gouv.viz/internal/config"
	"gouv.viz/internal/store"
	"gouv.viz/web/handlers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))

	e.Static("/assets", cfg.AssetsPath)

	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	st := store.New(db)
	validationCtx, cancelValidation := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelValidation()
	if err := st.Validate(validationCtx); err != nil {
		e.Logger.Fatal(err)
	}

	server := handlers.NewServer(st)

	e.GET("/", server.Home)
	e.GET("/scrutins", server.Scrutins)
	e.GET("/scrutins/:uid", server.ScrutinDetail)
	e.GET("/ping", handlers.Ping)

	if !cfg.IsProd() {
		e.GET("/ws", handlers.HotReloadWS)
	}

	e.Logger.Fatal(e.Start(cfg.Addr()))
}
