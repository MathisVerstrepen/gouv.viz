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

	view := scrutinsView(page)
	if ctx.Request().Header.Get("HX-Request") == "true" {
		return Render(ctx, http.StatusOK, components.ScrutinsExplorer(view))
	}

	return Render(ctx, http.StatusOK, components.Root(components.Scrutins(view), "Scrutins publics - gouv.viz"))
}

func parseScrutinsQuery(ctx echo.Context) store.ScrutinsQuery {
	page := 1
	if value, err := strconv.Atoi(ctx.QueryParam("page")); err == nil && value > 0 {
		page = value
	}

	return store.NormalizeScrutinsQuery(store.ScrutinsQuery{
		Search:      ctx.QueryParam("q"),
		Sort:        ctx.QueryParam("sort"),
		Page:        page,
		PerPage:     store.ScrutinsPerPage,
		Legislature: parsePositiveInt(ctx.QueryParam("legislature")),
		Result:      ctx.QueryParam("result"),
		VoteType:    ctx.QueryParam("vote_type"),
		Organe:      ctx.QueryParam("organe"),
		DateFrom:    ctx.QueryParam("date_from"),
		DateTo:      ctx.QueryParam("date_to"),
		CloseVotes:  ctx.QueryParam("close_votes") == "1",
	})
}

func parsePositiveInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0
	}
	return parsed
}
