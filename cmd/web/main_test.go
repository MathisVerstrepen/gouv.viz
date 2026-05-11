package main

import (
	"io"
	"net/http"
	"net/http/httptest"
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
			wantContains: []string{"script-src 'self'", "connect-src 'self'"},
			wantExcludes: []string{"'unsafe-inline'", "ws:", "wss:"},
		},
		{
			name:         "dev",
			env:          "dev",
			wantContains: []string{"script-src 'self'", "connect-src 'self' ws: wss:"},
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

func assertHeader(t *testing.T, rec *httptest.ResponseRecorder, key, want string) {
	t.Helper()

	if got := rec.Header().Get(key); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}
