package store

import (
	"context"
	"fmt"
	"strings"
)

const DeputiesPerPage = 25

type deputySortDefinition struct {
	value   string
	label   string
	orderBy string
}

var deputySortDefinitions = []deputySortDefinition{
	{value: "alpha_asc", label: "Nom alphabétique", orderBy: "COALESCE(NULLIF(a.alpha, ''), a.nom, a.prenom, a.uid) ASC, a.uid ASC"},
	{value: "votes_desc", label: "Plus de votes", orderBy: "COALESCE(vt.total_votes, 0) DESC, COALESCE(NULLIF(a.alpha, ''), a.nom, a.prenom, a.uid) ASC, a.uid ASC"},
	{value: "groupe_asc", label: "Groupe", orderBy: "COALESCE(NULLIF(COALESCE(lg.libelle_abrege, lg.libelle, lg.uid, ''), ''), 'zzzz') ASC, COALESCE(NULLIF(a.alpha, ''), a.nom, a.prenom, a.uid) ASC, a.uid ASC"},
	{value: "legislature_desc", label: "Législature récente", orderBy: "COALESCE(alg.legislature, 0) DESC, COALESCE(NULLIF(a.alpha, ''), a.nom, a.prenom, a.uid) ASC, a.uid ASC"},
}

func NormalizeDeputiesQuery(query DeputiesQuery) DeputiesQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = DeputiesPerPage
	}
	query.Search = strings.TrimSpace(query.Search)
	query.Group = strings.TrimSpace(query.Group)
	if deputySortByValue(query.Sort).value == "" {
		query.Sort = deputySortDefinitions[0].value
	}
	return query
}

func DefaultDeputiesSort() string {
	return deputySortDefinitions[0].value
}

func (s *Store) DeputiesPage(ctx context.Context, query DeputiesQuery) (DeputiesPage, error) {
	query = NormalizeDeputiesQuery(query)
	page := DeputiesPage{
		Query:       query,
		DefaultSort: DefaultDeputiesSort(),
		SortOptions: deputySortOptions(),
	}

	cache, err := s.staticCache(ctx)
	if err != nil {
		return DeputiesPage{}, err
	}
	page.FilterOptions = cloneDeputyFilterOptions(cache.deputyFilterOptions)

	whereClause, whereArgs := deputiesWhereClause(query)
	if err := s.db.QueryRowContext(ctx, deputiesListCTE()+`
SELECT COUNT(*)
FROM acteurs a
LEFT JOIN acteur_latest_group alg ON alg.acteur_uid = a.uid
LEFT JOIN organes lg ON lg.uid = alg.groupe_uid
LEFT JOIN vote_totals vt ON vt.acteur_uid = a.uid
`+whereClause, whereArgs...).Scan(&page.TotalResults); err != nil {
		return DeputiesPage{}, fmt.Errorf("count deputies: %w", err)
	}

	window := paginate(page.TotalResults, page.Query.Page, query.PerPage)
	page.Query.Page = window.Page
	page.TotalPages = window.TotalPages
	page.StartItem = window.StartItem
	page.EndItem = window.EndItem

	sortDefinition := deputySortByValue(page.Query.Sort)
	rowsArgs := append([]any{}, whereArgs...)
	rowsArgs = append(rowsArgs, query.PerPage, window.Offset)
	rows, err := s.db.QueryContext(ctx, deputiesListCTE()+`
SELECT
  a.uid,
  COALESCE(NULLIF(TRIM(COALESCE(a.prenom, '') || ' ' || COALESCE(a.nom, '')), ''), a.alpha, a.uid, ''),
  COALESCE(a.alpha, ''),
  COALESCE(a.profession, ''),
  COALESCE(a.date_naissance, ''),
  COALESCE(alg.legislature, 0),
  COALESCE(alg.groupe_uid, ''),
  COALESCE(lg.libelle_abrege, lg.libelle, alg.groupe_uid, ''),
  COALESCE(lg.libelle_abrege, ''),
  COALESCE(lg.libelle_abrev, ''),
  COALESCE(vt.total_votes, 0),
  COALESCE(vt.pour, 0),
  COALESCE(vt.contre, 0),
  COALESCE(vt.abstentions, 0),
  COALESCE(vt.non_votants, 0)
FROM acteurs a
LEFT JOIN acteur_latest_group alg ON alg.acteur_uid = a.uid
LEFT JOIN organes lg ON lg.uid = alg.groupe_uid
LEFT JOIN vote_totals vt ON vt.acteur_uid = a.uid
`+whereClause+`
ORDER BY `+sortDefinition.orderBy+`
LIMIT ? OFFSET ?
`, rowsArgs...)
	if err != nil {
		return DeputiesPage{}, fmt.Errorf("query deputies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var deputy DeputyListItem
		if err := rows.Scan(
			&deputy.UID,
			&deputy.DisplayName,
			&deputy.Alpha,
			&deputy.Profession,
			&deputy.DateNaissance,
			&deputy.Legislature,
			&deputy.GroupUID,
			&deputy.Group,
			&deputy.GroupAbrege,
			&deputy.GroupAbrev,
			&deputy.TotalVotes,
			&deputy.Pour,
			&deputy.Contre,
			&deputy.Abstentions,
			&deputy.NonVotants,
		); err != nil {
			return DeputiesPage{}, fmt.Errorf("scan deputy: %w", err)
		}
		page.Deputies = append(page.Deputies, deputy)
	}
	if err := rows.Err(); err != nil {
		return DeputiesPage{}, fmt.Errorf("iterate deputies: %w", err)
	}

	return page, nil
}

func (s *Store) queryDeputyFilterOptions(ctx context.Context) (DeputyFilterOptions, error) {
	legislatures, err := s.intFilterOptions(ctx, `
SELECT DISTINCT legislature
FROM mandats
WHERE legislature IS NOT NULL
ORDER BY legislature DESC
`)
	if err != nil {
		return DeputyFilterOptions{}, fmt.Errorf("query deputy legislature filter options: %w", err)
	}
	groups, err := s.textFilterOptions(ctx, `
SELECT uid, label
FROM (
  SELECT DISTINCT o.uid, COALESCE(o.libelle_abrege, o.libelle, o.uid) AS label, COALESCE(o.preseance, 999999) AS preseance
  FROM mandats m
  JOIN mandat_organes mo ON mo.mandat_uid = m.uid
  JOIN organes o ON o.uid = mo.organe_uid
  WHERE UPPER(COALESCE(o.code_type, '')) = 'GP'
)
ORDER BY preseance, label
`)
	if err != nil {
		return DeputyFilterOptions{}, fmt.Errorf("query deputy group filter options: %w", err)
	}

	return DeputyFilterOptions{
		Legislatures: mapFilterOptions[DeputyFilterOption](legislatures),
		Groups:       mapFilterOptions[DeputyFilterOption](groups),
	}, nil
}

func cloneDeputyFilterOptions(options DeputyFilterOptions) DeputyFilterOptions {
	return DeputyFilterOptions{
		Legislatures: cloneSlice(options.Legislatures),
		Groups:       cloneSlice(options.Groups),
	}
}

func deputySortOptions() []DeputySortOption {
	options := make([]DeputySortOption, 0, len(deputySortDefinitions))
	for _, definition := range deputySortDefinitions {
		options = append(options, DeputySortOption{Value: definition.value, Label: definition.label})
	}
	return options
}

func deputySortByValue(value string) deputySortDefinition {
	for _, definition := range deputySortDefinitions {
		if definition.value == value {
			return definition
		}
	}
	return deputySortDefinition{}
}

func deputiesWhereClause(query DeputiesQuery) (string, []any) {
	clauses := []string{}
	args := []any{}

	if query.Search != "" {
		clauses = append(clauses, `LOWER(
  COALESCE(a.uid, '') || ' ' ||
  COALESCE(a.prenom, '') || ' ' ||
  COALESCE(a.nom, '') || ' ' ||
  COALESCE(a.alpha, '') || ' ' ||
  COALESCE(a.profession, '') || ' ' ||
  COALESCE(lg.libelle, '') || ' ' ||
  COALESCE(lg.libelle_abrege, '') || ' ' ||
  COALESCE(lg.libelle_abrev, '')
) LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(strings.ToLower(query.Search))+"%")
	}
	if query.Legislature > 0 {
		clauses = append(clauses, `EXISTS (
  SELECT 1
  FROM mandats m_filter
  WHERE m_filter.acteur_uid = a.uid AND m_filter.legislature = ?
)`)
		args = append(args, query.Legislature)
	}
	if query.Group != "" {
		clauses = append(clauses, "alg.groupe_uid = ?")
		args = append(args, query.Group)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return "\nWHERE " + strings.Join(clauses, "\n  AND ") + "\n", args
}

func deputiesListCTE() string {
	return `
WITH vote_totals AS (
  SELECT acteur_uid, SUM(total_votes) AS total_votes, SUM(pour) AS pour, SUM(contre) AS contre, SUM(abstentions) AS abstentions, SUM(non_votants) AS non_votants
  FROM acteur_vote_stats
  GROUP BY acteur_uid
)
`
}
