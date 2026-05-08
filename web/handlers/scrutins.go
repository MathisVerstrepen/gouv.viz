package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"gouv.viz/internal/store"
	"gouv.viz/web/components"
)

func (s *Server) Scrutins(ctx echo.Context) error {
	query := parseScrutinsQuery(ctx)
	page, err := s.store.ScrutinsPage(ctx.Request().Context(), query)
	if err != nil {
		return fmt.Errorf("load scrutins page: %w", err)
	}

	return Render(ctx, http.StatusOK, components.Root(components.Scrutins(scrutinsView(page)), "Scrutins publics - gouv.viz"))
}

func parseScrutinsQuery(ctx echo.Context) store.ScrutinsQuery {
	page := 1
	if value, err := strconv.Atoi(ctx.QueryParam("page")); err == nil && value > 0 {
		page = value
	}

	return store.NormalizeScrutinsQuery(store.ScrutinsQuery{
		Search:  ctx.QueryParam("q"),
		Sort:    ctx.QueryParam("sort"),
		Page:    page,
		PerPage: store.ScrutinsPerPage,
	})
}
