package store

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

const ScrutinsPerPage = 25

type scrutinSortDefinition struct {
	value   string
	label   string
	orderBy string
}

var scrutinSortDefinitions = []scrutinSortDefinition{
	{value: "date_desc", label: "Date récente", orderBy: "s.date_scrutin DESC, s.numero DESC"},
	{value: "date_asc", label: "Date ancienne", orderBy: "s.date_scrutin ASC, s.numero ASC"},
	{value: "closest", label: "Votes les plus serrés", orderBy: "CASE WHEN s.pour IS NULL OR s.contre IS NULL THEN 1 ELSE 0 END ASC, ABS(COALESCE(s.pour, 0) - COALESCE(s.contre, 0)) ASC, s.date_scrutin DESC, s.numero DESC"},
	{value: "votants_desc", label: "Plus de votants", orderBy: "COALESCE(s.nombre_votants, 0) DESC, s.date_scrutin DESC, s.numero DESC"},
	{value: "pour_desc", label: "Plus de pour", orderBy: "COALESCE(s.pour, 0) DESC, s.date_scrutin DESC, s.numero DESC"},
	{value: "contre_desc", label: "Plus de contre", orderBy: "COALESCE(s.contre, 0) DESC, s.date_scrutin DESC, s.numero DESC"},
}

var scrutinSearchTokenRE = regexp.MustCompile(`[\p{L}\p{N}]+`)
var scrutinDateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func NormalizeScrutinsQuery(query ScrutinsQuery) ScrutinsQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = ScrutinsPerPage
	}
	query.Search = strings.TrimSpace(query.Search)
	query.Result = strings.TrimSpace(query.Result)
	query.VoteType = strings.TrimSpace(query.VoteType)
	query.Organe = strings.TrimSpace(query.Organe)
	query.DateFrom = normalizeScrutinDateFilter(query.DateFrom)
	query.DateTo = normalizeScrutinDateFilter(query.DateTo)
	if scrutinSortByValue(query.Sort).value == "" {
		query.Sort = scrutinSortDefinitions[0].value
	}
	return query
}

func DefaultScrutinsSort() string {
	return scrutinSortDefinitions[0].value
}

func (s *Store) ScrutinsPage(ctx context.Context, query ScrutinsQuery) (ScrutinsPage, error) {
	query = NormalizeScrutinsQuery(query)
	page := ScrutinsPage{
		Query:       query,
		DefaultSort: DefaultScrutinsSort(),
		SortOptions: scrutinSortOptions(),
	}
	cache, err := s.staticCache(ctx)
	if err != nil {
		return ScrutinsPage{}, err
	}
	page.FilterOptions = cloneScrutinFilterOptions(cache.scrutinFilterOptions)

	whereClause, whereArgs := scrutinsWhereClause(query)
	countQuery := `
SELECT COUNT(*)
FROM scrutins s
LEFT JOIN organes o ON o.uid = s.organe_uid
` + whereClause
	if err := s.db.QueryRowContext(ctx, countQuery, whereArgs...).Scan(&page.TotalResults); err != nil {
		return ScrutinsPage{}, fmt.Errorf("count scrutins: %w", err)
	}

	window := paginate(page.TotalResults, page.Query.Page, query.PerPage)
	page.Query.Page = window.Page
	page.TotalPages = window.TotalPages
	page.StartItem = window.StartItem
	page.EndItem = window.EndItem

	sortDefinition := scrutinSortByValue(page.Query.Sort)
	rowsArgs := append([]any{}, whereArgs...)
	rowsArgs = append(rowsArgs, query.PerPage, window.Offset)
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
		return ScrutinsPage{}, fmt.Errorf("query scrutins: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var scrutin ScrutinListItem
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
			return ScrutinsPage{}, fmt.Errorf("scan scrutin: %w", err)
		}
		page.Scrutins = append(page.Scrutins, scrutin)
	}
	if err := rows.Err(); err != nil {
		return ScrutinsPage{}, fmt.Errorf("iterate scrutins: %w", err)
	}

	return page, nil
}

func (s *Store) queryScrutinFilterOptions(ctx context.Context) (ScrutinFilterOptions, error) {
	legislatures, err := s.intFilterOptions(ctx, `
SELECT DISTINCT legislature
FROM scrutins
WHERE legislature IS NOT NULL
ORDER BY legislature DESC
`)
	if err != nil {
		return ScrutinFilterOptions{}, fmt.Errorf("query legislature filter options: %w", err)
	}
	results, err := s.textFilterOptions(ctx, `
SELECT DISTINCT sort_code, COALESCE(sort_libelle, sort_code)
FROM scrutins
WHERE sort_code IS NOT NULL AND sort_code <> ''
ORDER BY COALESCE(sort_libelle, sort_code)
`)
	if err != nil {
		return ScrutinFilterOptions{}, fmt.Errorf("query result filter options: %w", err)
	}
	voteTypes, err := s.textFilterOptions(ctx, `
SELECT DISTINCT code_type_vote, COALESCE(libelle_type_vote, code_type_vote)
FROM scrutins
WHERE code_type_vote IS NOT NULL AND code_type_vote <> ''
ORDER BY COALESCE(libelle_type_vote, code_type_vote)
`)
	if err != nil {
		return ScrutinFilterOptions{}, fmt.Errorf("query vote type filter options: %w", err)
	}
	organes, err := s.textFilterOptions(ctx, `
SELECT uid, label
FROM (
  SELECT DISTINCT o.uid, COALESCE(o.libelle_abrege, o.libelle, o.uid) AS label, COALESCE(o.preseance, 999999) AS preseance
  FROM organes o
  JOIN scrutins s ON s.organe_uid = o.uid
  UNION
  SELECT DISTINCT o.uid, COALESCE(o.libelle_abrege, o.libelle, o.uid) AS label, COALESCE(o.preseance, 999999) AS preseance
  FROM organes o
  JOIN scrutin_groupe_votes gv ON gv.groupe_uid = o.uid
)
ORDER BY preseance, label
	`)
	if err != nil {
		return ScrutinFilterOptions{}, fmt.Errorf("query organe filter options: %w", err)
	}

	return ScrutinFilterOptions{
		Legislatures: mapFilterOptions[ScrutinFilterOption](legislatures),
		Results:      mapFilterOptions[ScrutinFilterOption](results),
		VoteTypes:    mapFilterOptions[ScrutinFilterOption](voteTypes),
		Organes:      mapFilterOptions[ScrutinFilterOption](organes),
	}, nil
}

func cloneScrutinFilterOptions(options ScrutinFilterOptions) ScrutinFilterOptions {
	return ScrutinFilterOptions{
		Legislatures: cloneSlice(options.Legislatures),
		Results:      cloneSlice(options.Results),
		VoteTypes:    cloneSlice(options.VoteTypes),
		Organes:      cloneSlice(options.Organes),
	}
}

func scrutinSortOptions() []ScrutinSortOption {
	options := make([]ScrutinSortOption, 0, len(scrutinSortDefinitions))
	for _, definition := range scrutinSortDefinitions {
		options = append(options, ScrutinSortOption{
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

func scrutinsWhereClause(query ScrutinsQuery) (string, []any) {
	clauses := []string{}
	args := []any{}

	if query.Search != "" {
		ftsQuery := scrutinSearchQuery(query.Search)
		if ftsQuery == "" {
			clauses = append(clauses, "1 = 0")
		} else {
			clauses = append(clauses, `s.uid IN (
  SELECT uid
  FROM scrutin_search
  WHERE scrutin_search MATCH ?
)`)
			args = append(args, ftsQuery)
		}
	}
	if query.Legislature > 0 {
		clauses = append(clauses, "s.legislature = ?")
		args = append(args, query.Legislature)
	}
	if query.Result != "" {
		clauses = append(clauses, "s.sort_code = ?")
		args = append(args, query.Result)
	}
	if query.VoteType != "" {
		clauses = append(clauses, "s.code_type_vote = ?")
		args = append(args, query.VoteType)
	}
	if query.Organe != "" {
		clauses = append(clauses, `(s.organe_uid = ? OR EXISTS (
  SELECT 1
  FROM scrutin_groupe_votes gv
  WHERE gv.scrutin_uid = s.uid AND gv.groupe_uid = ?
))`)
		args = append(args, query.Organe, query.Organe)
	}
	if query.DateFrom != "" {
		clauses = append(clauses, "s.date_scrutin >= ?")
		args = append(args, query.DateFrom)
	}
	if query.DateTo != "" {
		clauses = append(clauses, "s.date_scrutin <= ?")
		args = append(args, query.DateTo)
	}
	if query.CloseVotes {
		clauses = append(clauses, "s.pour IS NOT NULL AND s.contre IS NOT NULL AND ABS(s.pour - s.contre) <= 10")
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return "\nWHERE " + strings.Join(clauses, "\n  AND ") + "\n", args
}

func scrutinSearchQuery(search string) string {
	tokens := scrutinSearchTokenRE.FindAllString(search, -1)
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		parts = append(parts, `"`+strings.ReplaceAll(token, `"`, `""`)+`"*`)
	}
	return strings.Join(parts, " AND ")
}

func normalizeScrutinDateFilter(value string) string {
	value = strings.TrimSpace(value)
	if !scrutinDateRE.MatchString(value) {
		return ""
	}
	return value
}
