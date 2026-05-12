package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
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

	e := newEcho(cfg)

	e.Static("/assets", cfg.AssetsPath)

	db, err := openWebDatabase(cfg.DatabasePath)
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
	e.GET("/deputes/:uid", server.DeputyDetail)
	e.GET("/ping", handlers.Ping)

	if !cfg.IsProd() {
		e.GET("/ws", handlers.HotReloadWS)
	}

	e.Logger.Fatal(e.StartServer(newHTTPServer(cfg.Addr())))
}

func openWebDatabase(databasePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", sqliteReadOnlyDSN(databasePath))
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)
	return db, nil
}

func sqliteReadOnlyDSN(databasePath string) string {
	query := url.Values{}
	query.Set("mode", "ro")
	query.Set("immutable", "1")
	query.Add("_pragma", "busy_timeout(5000)")
	query.Add("_pragma", "query_only(1)")

	if !filepath.IsAbs(databasePath) {
		return "file:" + (&url.URL{Path: databasePath}).EscapedPath() + "?" + query.Encode()
	}

	return (&url.URL{
		Scheme:   "file",
		Path:     databasePath,
		RawQuery: query.Encode(),
	}).String()
}

func newEcho(cfg config.Config) *echo.Echo {
	e := echo.New()
	csp := "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https://www.assemblee-nationale.fr; connect-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"
	if !cfg.IsProd() {
		csp = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https://www.assemblee-nationale.fr; connect-src 'self' ws: wss:; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"
	}

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XFrameOptions:         "DENY",
		ContentTypeNosniff:    "nosniff",
		HSTSMaxAge:            31536000,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: csp,
	}))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))
	e.HTTPErrorHandler = handlers.NewHTTPErrorHandler(e)
	return e
}

func newHTTPServer(addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
