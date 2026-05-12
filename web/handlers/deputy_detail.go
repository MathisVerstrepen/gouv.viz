package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"gouv.viz/internal/store"
	"gouv.viz/web/components"
)

func (s *Server) DeputyDetail(ctx echo.Context) error {
	page, err := s.store.DeputyDetailPage(ctx.Request().Context(), ctx.Param("uid"), parseDeputyDetailQuery(ctx))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "député introuvable")
		}
		return fmt.Errorf("load deputy detail page: %w", err)
	}

	view := deputyDetailView(page)
	if ctx.Request().Header.Get("HX-Request") == "true" {
		return Render(ctx, http.StatusOK, components.DeputyVotesPanel(view))
	}

	title := fmt.Sprintf("%s - gouv.viz", view.Deputy.DisplayName)
	return Render(ctx, http.StatusOK, components.Root(components.DeputyDetail(view), title))
}

func parseDeputyDetailQuery(ctx echo.Context) store.DeputyDetailQuery {
	return store.NormalizeDeputyDetailQuery(store.DeputyDetailQuery{
		VotesPage:     parsePositiveInt(ctx.QueryParam("votes_page")),
		VotesPerPage:  store.DeputyVotesPerPage,
		VotesSearch:   ctx.QueryParam("votes_q"),
		VotesSort:     ctx.QueryParam("votes_sort"),
		VotesPosition: ctx.QueryParam("votes_position"),
	})
}
