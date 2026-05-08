package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func Ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}

func HotReloadWS(ctx echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			msg := ""
			if err := websocket.Message.Receive(ws, &msg); err != nil {
				return
			}
		}
	}).ServeHTTP(ctx.Response(), ctx.Request())

	return nil
}
