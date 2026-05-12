package store

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
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
	{value: "groupe_asc", label: "Groupe", orderBy: "COALESCE(NULLIF(lg.groupe, ''), 'zzzz') ASC, COALESCE(NULLIF(a.alpha, ''), a.nom, a.prenom, a.uid) ASC, a.uid ASC"},
	{value: "legislature_desc", label: "Législature récente", orderBy: "COALESCE(lg.legislature, 0) DESC, COALESCE(NULLIF(a.alpha, ''), a.nom, a.prenom, a.uid) ASC, a.uid ASC"},
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

	filterOptions, err := s.deputyFilterOptions(ctx)
	if err != nil {
		return DeputiesPage{}, err
	}
	page.FilterOptions = filterOptions

	whereClause, whereArgs := deputiesWhereClause(query)
	if err := s.db.QueryRowContext(ctx, deputiesListCTE()+`
SELECT COUNT(*)
FROM acteurs a
LEFT JOIN latest_group lg ON lg.acteur_uid = a.uid
LEFT JOIN vote_totals vt ON vt.acteur_uid = a.uid
`+whereClause, whereArgs...).Scan(&page.TotalResults); err != nil {
		return DeputiesPage{}, fmt.Errorf("count deputies: %w", err)
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

	sortDefinition := deputySortByValue(page.Query.Sort)
	rowsArgs := append([]any{}, whereArgs...)
	rowsArgs = append(rowsArgs, query.PerPage, (page.Query.Page-1)*query.PerPage)
	rows, err := s.db.QueryContext(ctx, deputiesListCTE()+`
SELECT
  a.uid,
  COALESCE(NULLIF(TRIM(COALESCE(a.prenom, '') || ' ' || COALESCE(a.nom, '')), ''), a.alpha, a.uid, ''),
  COALESCE(a.alpha, ''),
  COALESCE(a.profession, ''),
  COALESCE(a.date_naissance, ''),
  COALESCE(lg.legislature, 0),
  COALESCE(lg.groupe_uid, ''),
  COALESCE(lg.groupe, ''),
  COALESCE(lg.groupe_abrege, ''),
  COALESCE(lg.groupe_abrev, ''),
  COALESCE(vt.total_votes, 0),
  COALESCE(vt.pour, 0),
  COALESCE(vt.contre, 0),
  COALESCE(vt.abstentions, 0),
  COALESCE(vt.non_votants, 0)
FROM acteurs a
LEFT JOIN latest_group lg ON lg.acteur_uid = a.uid
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

func (s *Store) deputyFilterOptions(ctx context.Context) (DeputyFilterOptions, error) {
	legislatures, err := s.deputyIntFilterOptions(ctx, `
SELECT DISTINCT legislature
FROM mandats
WHERE legislature IS NOT NULL
ORDER BY legislature DESC
`)
	if err != nil {
		return DeputyFilterOptions{}, fmt.Errorf("query deputy legislature filter options: %w", err)
	}
	groups, err := s.deputyTextFilterOptions(ctx, `
SELECT DISTINCT o.uid, COALESCE(o.libelle_abrege, o.libelle, o.uid) AS label, COALESCE(o.preseance, 999999) AS preseance
FROM mandats m
JOIN mandat_organes mo ON mo.mandat_uid = m.uid
JOIN organes o ON o.uid = mo.organe_uid
WHERE UPPER(COALESCE(o.code_type, '')) = 'GP'
ORDER BY preseance, label
`)
	if err != nil {
		return DeputyFilterOptions{}, fmt.Errorf("query deputy group filter options: %w", err)
	}

	return DeputyFilterOptions{
		Legislatures: legislatures,
		Groups:       groups,
	}, nil
}

func (s *Store) deputyIntFilterOptions(ctx context.Context, query string) ([]DeputyFilterOption, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []DeputyFilterOption
	for rows.Next() {
		var value int
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		text := strconv.Itoa(value)
		options = append(options, DeputyFilterOption{Value: text, Label: text})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return options, nil
}

func (s *Store) deputyTextFilterOptions(ctx context.Context, query string) ([]DeputyFilterOption, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []DeputyFilterOption
	for rows.Next() {
		var value string
		var label sql.NullString
		var ignoredPreseance int
		if err := rows.Scan(&value, &label, &ignoredPreseance); err != nil {
			return nil, err
		}
		if value == "" {
			continue
		}
		optionLabel := value
		if label.Valid && label.String != "" {
			optionLabel = label.String
		}
		options = append(options, DeputyFilterOption{Value: value, Label: optionLabel})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return options, nil
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
  COALESCE(lg.groupe, '') || ' ' ||
  COALESCE(lg.groupe_abrege, '') || ' ' ||
  COALESCE(lg.groupe_abrev, '')
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
		clauses = append(clauses, "lg.groupe_uid = ?")
		args = append(args, query.Group)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return "\nWHERE " + strings.Join(clauses, "\n  AND ") + "\n", args
}

func deputiesListCTE() string {
	return `
WITH ranked_groups AS (
  SELECT
    m.acteur_uid,
    m.legislature,
    o.uid AS groupe_uid,
    COALESCE(o.libelle_abrege, o.libelle, o.uid, '') AS groupe,
    COALESCE(o.libelle_abrege, '') AS groupe_abrege,
    COALESCE(o.libelle_abrev, '') AS groupe_abrev,
    ROW_NUMBER() OVER (
      PARTITION BY m.acteur_uid
      ORDER BY COALESCE(m.date_debut, '') DESC, COALESCE(m.preseance, 9999), m.uid, COALESCE(o.preseance, 9999)
    ) AS rn
  FROM mandats m
  JOIN mandat_organes mo ON mo.mandat_uid = m.uid
  JOIN organes o ON o.uid = mo.organe_uid
  WHERE UPPER(COALESCE(o.code_type, '')) = 'GP'
),
latest_group AS (
  SELECT acteur_uid, legislature, groupe_uid, groupe, groupe_abrege, groupe_abrev
  FROM ranked_groups
  WHERE rn = 1
),
vote_totals AS (
  SELECT acteur_uid, SUM(total_votes) AS total_votes, SUM(pour) AS pour, SUM(contre) AS contre, SUM(abstentions) AS abstentions, SUM(non_votants) AS non_votants
  FROM acteur_vote_stats
  GROUP BY acteur_uid
)
`
}
