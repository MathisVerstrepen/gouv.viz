package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"gouv.viz/internal/store"
	"gouv.viz/web/components"
)

func (s *Server) PoliticalGroups(ctx echo.Context) error {
	query := parsePoliticalGroupsQuery(ctx)
	page, err := s.store.PoliticalGroupsPage(ctx.Request().Context(), query)
	if err != nil {
		return fmt.Errorf("load political groups page: %w", err)
	}

	view := politicalGroupsView(page)
	if ctx.Request().Header.Get("HX-Request") == "true" {
		return Render(ctx, http.StatusOK, components.PoliticalGroupsExplorer(view))
	}

	return Render(ctx, http.StatusOK, components.Root(components.PoliticalGroups(view), "Groupes politiques - gouv.viz"))
}

func parsePoliticalGroupsQuery(ctx echo.Context) store.PoliticalGroupsQuery {
	page := 1
	if value, err := strconv.Atoi(ctx.QueryParam("page")); err == nil && value > 0 {
		page = value
	}

	return store.NormalizePoliticalGroupsQuery(store.PoliticalGroupsQuery{
		Search:      ctx.QueryParam("q"),
		Sort:        ctx.QueryParam("sort"),
		Page:        page,
		PerPage:     store.PoliticalGroupsPerPage,
		Legislature: parsePositiveInt(ctx.QueryParam("legislature")),
	})
}
