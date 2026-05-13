package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"gouv.viz/internal/store"
	"gouv.viz/web/components"
)

func (s *Server) PoliticalGroupDetail(ctx echo.Context) error {
	page, err := s.store.PoliticalGroupDetailPage(ctx.Request().Context(), ctx.Param("uid"), parsePoliticalGroupDetailQuery(ctx))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "groupe politique introuvable")
		}
		return fmt.Errorf("load political group detail page: %w", err)
	}

	view := politicalGroupDetailView(page)
	if ctx.Request().Header.Get("HX-Request") == "true" {
		return Render(ctx, http.StatusOK, components.PoliticalGroupVotesPanel(view))
	}

	label := components.PoliticalGroupLabel(view.Group)
	title := fmt.Sprintf("%s - Groupe politique - gouv.viz", label)
	description := fmt.Sprintf("Groupe politique %s à l'Assemblée nationale. Découvrez sa composition, ses députés et ses positions de vote.", label)
	return Render(ctx, http.StatusOK, components.Root(components.PoliticalGroupDetail(view), title, description))
}

func parsePoliticalGroupDetailQuery(ctx echo.Context) store.PoliticalGroupDetailQuery {
	return store.NormalizePoliticalGroupDetailQuery(store.PoliticalGroupDetailQuery{
		VotesPage:     parsePositiveInt(ctx.QueryParam("votes_page")),
		VotesPerPage:  store.PoliticalGroupVotesPerPage,
		VotesSearch:   ctx.QueryParam("votes_q"),
		VotesSort:     ctx.QueryParam("votes_sort"),
		VotesPosition: ctx.QueryParam("votes_position"),
	})
}
