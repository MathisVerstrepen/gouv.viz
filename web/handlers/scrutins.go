package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"gouv.viz/web/components"
)

const scrutinsPerPage = 25

type scrutinSortDefinition struct {
	value   string
	label   string
	orderBy string
}

var scrutinSortDefinitions = []scrutinSortDefinition{
	{value: "date_desc", label: "Date recente", orderBy: "s.date_scrutin DESC, s.numero DESC"},
	{value: "date_asc", label: "Date ancienne", orderBy: "s.date_scrutin ASC, s.numero ASC"},
	{value: "numero_desc", label: "Numero decroissant", orderBy: "s.numero DESC"},
	{value: "numero_asc", label: "Numero croissant", orderBy: "s.numero ASC"},
	{value: "votants_desc", label: "Plus de votants", orderBy: "COALESCE(s.nombre_votants, 0) DESC, s.date_scrutin DESC, s.numero DESC"},
	{value: "pour_desc", label: "Plus de pour", orderBy: "COALESCE(s.pour, 0) DESC, s.date_scrutin DESC, s.numero DESC"},
	{value: "contre_desc", label: "Plus de contre", orderBy: "COALESCE(s.contre, 0) DESC, s.date_scrutin DESC, s.numero DESC"},
}

func Scrutins(ctx echo.Context) error {
	query := parseScrutinsQuery(ctx)
	page, err := loadScrutinsPage(os.Getenv("DATABASE_PATH"), query)
	if err != nil {
		return fmt.Errorf("load scrutins page: %w", err)
	}

	return Render(ctx, http.StatusOK, components.Root(components.Scrutins(page), "Scrutins publics - gouv.viz"))
}

func parseScrutinsQuery(ctx echo.Context) components.ScrutinsQuery {
	page := 1
	if value, err := strconv.Atoi(ctx.QueryParam("page")); err == nil && value > 0 {
		page = value
	}

	sort := ctx.QueryParam("sort")
	if scrutinSortByValue(sort).value == "" {
		sort = scrutinSortDefinitions[0].value
	}

	return components.ScrutinsQuery{
		Search:  strings.TrimSpace(ctx.QueryParam("q")),
		Sort:    sort,
		Page:    page,
		PerPage: scrutinsPerPage,
	}
}

func loadScrutinsPage(databasePath string, query components.ScrutinsQuery) (components.ScrutinsPage, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return components.ScrutinsPage{}, fmt.Errorf("open sqlite database: %w", err)
	}
	defer db.Close()

	page := components.ScrutinsPage{
		Query:       query,
		SortOptions: scrutinSortOptions(),
	}

	whereClause, whereArgs := scrutinsSearchClause(query.Search)
	countQuery := `
SELECT COUNT(*)
FROM scrutins s
LEFT JOIN organes o ON o.uid = s.organe_uid
` + whereClause
	if err := db.QueryRow(countQuery, whereArgs...).Scan(&page.TotalResults); err != nil {
		return components.ScrutinsPage{}, fmt.Errorf("count scrutins: %w", err)
	}

	if page.TotalResults > 0 {
		page.TotalPages = int(math.Ceil(float64(page.TotalResults) / float64(query.PerPage)))
		if page.Query.Page > page.TotalPages {
			page.Query.Page = page.TotalPages
		}
		page.StartItem = ((page.Query.Page - 1) * query.PerPage) + 1
		page.EndItem = page.StartItem + query.PerPage - 1
		if page.EndItem > page.TotalResults {
			page.EndItem = page.TotalResults
		}
	} else {
		page.Query.Page = 1
		page.TotalPages = 1
	}

	sortDefinition := scrutinSortByValue(page.Query.Sort)
	rowsArgs := append([]any{}, whereArgs...)
	rowsArgs = append(rowsArgs, query.PerPage, (page.Query.Page-1)*query.PerPage)
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
`+whereClause+`
ORDER BY `+sortDefinition.orderBy+`
LIMIT ? OFFSET ?
`, rowsArgs...)
	if err != nil {
		return components.ScrutinsPage{}, fmt.Errorf("query scrutins: %w", err)
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
			return components.ScrutinsPage{}, fmt.Errorf("scan scrutin: %w", err)
		}
		page.Scrutins = append(page.Scrutins, scrutin)
	}
	if err := rows.Err(); err != nil {
		return components.ScrutinsPage{}, fmt.Errorf("iterate scrutins: %w", err)
	}

	return page, nil
}

func scrutinSortOptions() []components.ScrutinSortOption {
	options := make([]components.ScrutinSortOption, 0, len(scrutinSortDefinitions))
	for _, definition := range scrutinSortDefinitions {
		options = append(options, components.ScrutinSortOption{
			Value: definition.value,
			Label: definition.label,
		})
	}
	return options
}

func scrutinSortByValue(value string) scrutinSortDefinition {
	for _, definition := range scrutinSortDefinitions {
		if definition.value == value {
			return definition
		}
	}
	return scrutinSortDefinition{}
}

func scrutinsSearchClause(search string) (string, []any) {
	if search == "" {
		return "", nil
	}

	pattern := "%" + escapeSQLiteLike(search) + "%"
	args := make([]any, 0, 9)
	for range 9 {
		args = append(args, pattern)
	}

	return `
WHERE s.titre LIKE ? ESCAPE '\'
   OR s.objet_libelle LIKE ? ESCAPE '\'
   OR s.demandeur_texte LIKE ? ESCAPE '\'
   OR s.sort_code LIKE ? ESCAPE '\'
   OR s.sort_libelle LIKE ? ESCAPE '\'
   OR s.libelle_type_vote LIKE ? ESCAPE '\'
   OR o.libelle LIKE ? ESCAPE '\'
   OR o.libelle_abrege LIKE ? ESCAPE '\'
   OR CAST(s.numero AS TEXT) LIKE ? ESCAPE '\'
`, args
}

func escapeSQLiteLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
}
