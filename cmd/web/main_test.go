package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestNewEchoAddsSecurityHeadersAndRequestID(t *testing.T) {
	e := newEcho()
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
