package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"gouv.viz/internal/store"
	"gouv.viz/web/components"
)

func (s *Server) Deputies(ctx echo.Context) error {
	query := parseDeputiesQuery(ctx)
	page, err := s.store.DeputiesPage(ctx.Request().Context(), query)
	if err != nil {
		return fmt.Errorf("load deputies page: %w", err)
	}

	view := deputiesView(page)
	if ctx.Request().Header.Get("HX-Request") == "true" {
		return Render(ctx, http.StatusOK, components.DeputiesExplorer(view))
	}

	return Render(ctx, http.StatusOK, components.Root(components.Deputies(view), "Députés - gouv.viz", "Liste des députés de l'Assemblée nationale. Consultez les fiches parlementaires, les votes et les statistiques de chaque député."))
}

func parseDeputiesQuery(ctx echo.Context) store.DeputiesQuery {
	page := 1
	if value, err := strconv.Atoi(ctx.QueryParam("page")); err == nil && value > 0 {
		page = value
	}

	return store.NormalizeDeputiesQuery(store.DeputiesQuery{
		Search:      ctx.QueryParam("q"),
		Sort:        ctx.QueryParam("sort"),
		Page:        page,
		PerPage:     store.DeputiesPerPage,
		Legislature: parsePositiveInt(ctx.QueryParam("legislature")),
		Group:       ctx.QueryParam("group"),
	})
}
