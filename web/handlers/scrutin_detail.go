package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"gouv.viz/internal/store"
	"gouv.viz/web/components"
)

func (s *Server) ScrutinDetail(ctx echo.Context) error {
	page, err := s.store.ScrutinDetailPage(ctx.Request().Context(), ctx.Param("uid"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "scrutin introuvable")
		}
		return fmt.Errorf("load scrutin detail page: %w", err)
	}

	view := scrutinDetailView(page)
	title := fmt.Sprintf("Scrutin n%d - gouv.viz", view.Scrutin.Numero)
	return Render(ctx, http.StatusOK, components.Root(components.ScrutinDetail(view), title))
}
