package main

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"gouv.viz/internal/config"
)

func TestNewEchoAddsSecurityHeadersAndRequestID(t *testing.T) {
	e := newEcho(config.Config{Env: "prod"})
	e.Logger.SetOutput(io.Discard)
	e.GET("/ok", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Header().Get(echo.HeaderXRequestID) == "" {
		t.Fatal("X-Request-ID header is empty")
	}
	assertHeader(t, rec, echo.HeaderXFrameOptions, "DENY")
	assertHeader(t, rec, echo.HeaderXContentTypeOptions, "nosniff")
	assertHeader(t, rec, echo.HeaderReferrerPolicy, "strict-origin-when-cross-origin")
	if got := rec.Header().Get(echo.HeaderContentSecurityPolicy); !strings.Contains(got, "default-src 'self'") {
		t.Fatalf("Content-Security-Policy = %q, want default-src 'self'", got)
	}
}

func TestNewEchoContentSecurityPolicy(t *testing.T) {
	tests := []struct {
		name         string
		env          string
		wantContains []string
		wantExcludes []string
	}{
		{
			name:         "prod",
			env:          "prod",
			wantContains: []string{"script-src 'self'", "img-src 'self' data: https://www.assemblee-nationale.fr", "connect-src 'self'"},
			wantExcludes: []string{"'unsafe-inline'", "ws:", "wss:"},
		},
		{
			name:         "dev",
			env:          "dev",
			wantContains: []string{"script-src 'self'", "img-src 'self' data: https://www.assemblee-nationale.fr", "connect-src 'self' ws: wss:"},
			wantExcludes: []string{"'unsafe-inline'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEcho(config.Config{Env: tt.env})
			e.Logger.SetOutput(io.Discard)
			e.GET("/ok", func(ctx echo.Context) error {
				return ctx.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/ok", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			got := rec.Header().Get(echo.HeaderContentSecurityPolicy)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Fatalf("Content-Security-Policy = %q, want %q", got, want)
				}
			}
			for _, excluded := range tt.wantExcludes {
				if strings.Contains(got, excluded) {
					t.Fatalf("Content-Security-Policy = %q, should not contain %q", got, excluded)
				}
			}
		})
	}
}

func TestNewHTTPServerSetsTimeouts(t *testing.T) {
	server := newHTTPServer(":9456")

	if server.Addr != ":9456" {
		t.Fatalf("Addr = %q, want :9456", server.Addr)
	}
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 10*time.Second || server.WriteTimeout != 30*time.Second || server.IdleTimeout != 120*time.Second {
		t.Fatalf("timeouts = readHeader:%s read:%s write:%s idle:%s", server.ReadHeaderTimeout, server.ReadTimeout, server.WriteTimeout, server.IdleTimeout)
	}
}

func TestSQLiteReadOnlyDSN(t *testing.T) {
	dsn := sqliteReadOnlyDSN("data/processed/gouv viz.sqlite")
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", dsn, err)
	}

	if u.Scheme != "file" {
		t.Fatalf("scheme = %q, want file", u.Scheme)
	}
	if u.Host != "" {
		t.Fatalf("host = %q, want empty host for relative SQLite URI", u.Host)
	}
	if u.Opaque == "" {
		t.Fatalf("opaque path is empty, want relative SQLite URI path")
	}
	query := u.Query()
	if got := query.Get("mode"); got != "ro" {
		t.Fatalf("mode = %q, want ro", got)
	}
	if got := query.Get("immutable"); got != "1" {
		t.Fatalf("immutable = %q, want 1", got)
	}
	assertQueryValue(t, query["_pragma"], "busy_timeout(5000)")
	assertQueryValue(t, query["_pragma"], "query_only(1)")
}

func TestOpenWebDatabaseIsReadOnly(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "web.sqlite")
	writable, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if _, err := writable.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create fixture table error = %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("close fixture database error = %v", err)
	}

	db, err := openWebDatabase(dbPath)
	if err != nil {
		t.Fatalf("openWebDatabase() error = %v", err)
	}
	defer db.Close()

	var queryOnly int
	if err := db.QueryRow("PRAGMA query_only").Scan(&queryOnly); err != nil {
		t.Fatalf("PRAGMA query_only error = %v", err)
	}
	if queryOnly != 1 {
		t.Fatalf("query_only = %d, want 1", queryOnly)
	}

	var busyTimeout int
	if err := db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatalf("PRAGMA busy_timeout error = %v", err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", busyTimeout)
	}

	if _, err := db.Exec("INSERT INTO items (id) VALUES (1)"); err == nil {
		t.Fatal("write through web database succeeded, want read-only error")
	}
}

func TestOpenWebDatabaseSupportsRelativePath(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(filepath.Join("data", "processed"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	dbPath := filepath.Join("data", "processed", "web.sqlite")
	writable, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if _, err := writable.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create fixture table error = %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("close fixture database error = %v", err)
	}

	db, err := openWebDatabase(dbPath)
	if err != nil {
		t.Fatalf("openWebDatabase() error = %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		t.Fatalf("relative read-only database query error = %v", err)
	}
}

func assertHeader(t *testing.T, rec *httptest.ResponseRecorder, key, want string) {
	t.Helper()

	if got := rec.Header().Get(key); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func assertQueryValue(t *testing.T, values []string, want string) {
	t.Helper()

	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("query values = %v, want %q", values, want)
}
