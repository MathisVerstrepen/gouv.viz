package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"gouv.viz/web/components"
)

func (s *Server) Home(ctx echo.Context) error {
	page, err := s.store.HomePage(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("load homepage data: %w", err)
	}

	return Render(ctx, http.StatusOK, components.Root(components.Home(homeView(page)), "Gouv.viz - Visualisation des scrutins publics", "Visualisation des scrutins publics de l'Assemblée nationale. Explorez les votes, les députés et les groupes politiques à partir des données ouvertes."))
}
