package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"gouv.viz/web/components"
)

func NewHTTPErrorHandler(e *echo.Echo) echo.HTTPErrorHandler {
	return func(err error, ctx echo.Context) {
		if ctx.Response().Committed {
			return
		}

		code := http.StatusInternalServerError
		message := "Une erreur inattendue est survenue."
		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			code = httpError.Code
			message = publicErrorMessage(code)
		}

		if code >= http.StatusInternalServerError {
			ctx.Logger().Error(err)
		}

		if ctx.Request().Method == http.MethodHead {
			if renderErr := ctx.NoContent(code); renderErr != nil {
				e.DefaultHTTPErrorHandler(renderErr, ctx)
			}
			return
		}

		page := components.ErrorPageData{
			StatusCode: code,
			Title:      http.StatusText(code),
			Message:    message,
		}
		if renderErr := Render(ctx, code, components.Root(components.ErrorPage(page), fmt.Sprintf("%d - gouv.viz", code), "Une erreur est survenue sur gouv.viz.")); renderErr != nil {
			e.DefaultHTTPErrorHandler(renderErr, ctx)
		}
	}
}

func publicErrorMessage(code int) string {
	switch code {
	case http.StatusNotFound:
		return "La page demandée est introuvable."
	case http.StatusMethodNotAllowed:
		return "Cette méthode HTTP n'est pas autorisée pour cette page."
	case http.StatusBadRequest:
		return "La requête ne peut pas être traitée en l'état."
	default:
		if code >= http.StatusInternalServerError {
			return "Une erreur inattendue est survenue."
		}
		return http.StatusText(code)
	}
}
