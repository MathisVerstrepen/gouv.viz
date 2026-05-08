package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"

	"gouv.viz/web/components"
)

func Home(ctx echo.Context) error {
	page, err := loadHomePage(os.Getenv("DATABASE_PATH"))
	if err != nil {
		return fmt.Errorf("load homepage data: %w", err)
	}

	return Render(ctx, http.StatusOK, components.Root(components.Home(page), "gouv.viz"))
}

func loadHomePage(databasePath string) (components.HomePage, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return components.HomePage{}, fmt.Errorf("open sqlite database: %w", err)
	}
	defer db.Close()

	var page components.HomePage
	if err := db.QueryRow(`
SELECT COUNT(*), MIN(date_scrutin), MAX(date_scrutin)
FROM scrutins
`).Scan(&page.TotalScrutins, &page.FirstScrutinDate, &page.LastScrutinDate); err != nil {
		return components.HomePage{}, fmt.Errorf("query scrutin totals: %w", err)
	}

	rows, err := db.Query(`
SELECT
  s.uid,
  s.numero,
  s.date_scrutin,
  s.titre,
  s.sort_code,
  s.libelle_type_vote,
  COALESCE(o.libelle_abrege, o.libelle, s.organe_uid, ''),
  COALESCE(s.pour, 0),
  COALESCE(s.contre, 0),
  COALESCE(s.abstentions, 0),
  COALESCE(s.nombre_votants, 0)
FROM scrutins s
LEFT JOIN organes o ON o.uid = s.organe_uid
ORDER BY s.date_scrutin DESC, s.numero DESC
LIMIT 50
`)
	if err != nil {
		return components.HomePage{}, fmt.Errorf("query scrutins: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var scrutin components.ScrutinListItem
		if err := rows.Scan(
			&scrutin.UID,
			&scrutin.Numero,
			&scrutin.Date,
			&scrutin.Titre,
			&scrutin.SortCode,
			&scrutin.TypeVote,
			&scrutin.Organe,
			&scrutin.Pour,
			&scrutin.Contre,
			&scrutin.Abstentions,
			&scrutin.NombreVotants,
		); err != nil {
			return components.HomePage{}, fmt.Errorf("scan scrutin: %w", err)
		}
		page.Scrutins = append(page.Scrutins, scrutin)
	}
	if err := rows.Err(); err != nil {
		return components.HomePage{}, fmt.Errorf("iterate scrutins: %w", err)
	}

	return page, nil
}
