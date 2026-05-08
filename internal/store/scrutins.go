package store

import (
	"context"
	"fmt"
	"math"
	"strings"

	"gouv.viz/web/components"
)

const ScrutinsPerPage = 25

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

func NormalizeScrutinsQuery(query components.ScrutinsQuery) components.ScrutinsQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = ScrutinsPerPage
	}
	query.Search = strings.TrimSpace(query.Search)
	if scrutinSortByValue(query.Sort).value == "" {
		query.Sort = scrutinSortDefinitions[0].value
	}
	return query
}

func (s *Store) ScrutinsPage(ctx context.Context, query components.ScrutinsQuery) (components.ScrutinsPage, error) {
	query = NormalizeScrutinsQuery(query)
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
	if err := s.db.QueryRowContext(ctx, countQuery, whereArgs...).Scan(&page.TotalResults); err != nil {
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
	rows, err := s.db.QueryContext(ctx, `
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
