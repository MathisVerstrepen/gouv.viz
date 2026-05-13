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
	title := fmt.Sprintf("Scrutin n°%d : %s - gouv.viz", view.Scrutin.Numero, view.Scrutin.Titre)
	description := fmt.Sprintf("%s - Scrutin public n°%d de la %de législature. Résultat : %s.", view.Scrutin.Titre, view.Scrutin.Numero, view.Scrutin.Legislature, view.Scrutin.SortLibelle)
	return Render(ctx, http.StatusOK, components.Root(components.ScrutinDetail(view), title, description))
}
