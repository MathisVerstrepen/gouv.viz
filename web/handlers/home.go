package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"gouv.viz/web/components"
)

func Home(ctx echo.Context) error {
	return Render(ctx, http.StatusOK, components.Root(components.Home(), "gouv.viz"))
}
