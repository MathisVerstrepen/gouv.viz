package store

import (
	"context"
	"fmt"
	"strings"
)

const PoliticalGroupsPerPage = 25

type politicalGroupSortDefinition struct {
	value   string
	label   string
	orderBy string
}

var politicalGroupSortDefinitions = []politicalGroupSortDefinition{
	{value: "preseance_asc", label: "Ordre institutionnel", orderBy: "COALESCE(g.preseance, 999999) ASC, COALESCE(NULLIF(g.libelle_abrege, ''), g.libelle, g.uid) ASC"},
	{value: "name_asc", label: "Nom", orderBy: "COALESCE(NULLIF(g.libelle, ''), g.libelle_abrege, g.libelle_abrev, g.uid) ASC"},
	{value: "deputies_desc", label: "Plus de députés", orderBy: "COALESCE(md.deputies_count, 0) DESC, COALESCE(g.preseance, 999999) ASC"},
	{value: "votes_desc", label: "Plus de scrutins", orderBy: "COALESCE(vs.total_scrutins, 0) DESC, COALESCE(g.preseance, 999999) ASC"},
	{value: "legislature_desc", label: "Législature récente", orderBy: "COALESCE(g.legislature, 0) DESC, COALESCE(g.preseance, 999999) ASC"},
}

func NormalizePoliticalGroupsQuery(query PoliticalGroupsQuery) PoliticalGroupsQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = PoliticalGroupsPerPage
	}
	query.Search = strings.TrimSpace(query.Search)
	if politicalGroupSortByValue(query.Sort).value == "" {
		query.Sort = politicalGroupSortDefinitions[0].value
	}
	return query
}

func DefaultPoliticalGroupsSort() string {
	return politicalGroupSortDefinitions[0].value
}

func (s *Store) PoliticalGroupsPage(ctx context.Context, query PoliticalGroupsQuery) (PoliticalGroupsPage, error) {
	query = NormalizePoliticalGroupsQuery(query)
	page := PoliticalGroupsPage{
		Query:       query,
		DefaultSort: DefaultPoliticalGroupsSort(),
		SortOptions: politicalGroupSortOptions(),
	}

	cache, err := s.staticCache(ctx)
	if err != nil {
		return PoliticalGroupsPage{}, err
	}
	page.FilterOptions = clonePoliticalGroupFilterOptions(cache.politicalGroupFilterOptions)

	whereClause, whereArgs := politicalGroupsWhereClause(query)
	if err := s.db.QueryRowContext(ctx, politicalGroupsListCTE()+`
SELECT COUNT(*)
FROM organes g
LEFT JOIN member_counts md ON md.groupe_uid = g.uid
LEFT JOIN vote_stats vs ON vs.groupe_uid = g.uid
`+whereClause, whereArgs...).Scan(&page.TotalResults); err != nil {
		return PoliticalGroupsPage{}, fmt.Errorf("count political groups: %w", err)
	}

	window := paginate(page.TotalResults, page.Query.Page, query.PerPage)
	page.Query.Page = window.Page
	page.TotalPages = window.TotalPages
	page.StartItem = window.StartItem
	page.EndItem = window.EndItem

	sortDefinition := politicalGroupSortByValue(page.Query.Sort)
	rowsArgs := append([]any{}, whereArgs...)
	rowsArgs = append(rowsArgs, query.PerPage, window.Offset)
	rows, err := s.db.QueryContext(ctx, politicalGroupsListCTE()+`
SELECT
  g.uid,
  COALESCE(g.libelle, ''),
  COALESCE(g.libelle_abrege, ''),
  COALESCE(g.libelle_abrev, ''),
  COALESCE(g.legislature, 0),
  COALESCE(g.position_politique, ''),
  COALESCE(g.preseance, 0),
  COALESCE(g.date_debut, ''),
  COALESCE(g.date_fin, ''),
  COALESCE(md.deputies_count, 0),
  COALESCE(vs.total_scrutins, 0),
  COALESCE(vs.pour, 0),
  COALESCE(vs.contre, 0),
  COALESCE(vs.abstentions, 0),
  COALESCE(vs.non_votants, 0)
FROM organes g
LEFT JOIN member_counts md ON md.groupe_uid = g.uid
LEFT JOIN vote_stats vs ON vs.groupe_uid = g.uid
`+whereClause+`
ORDER BY `+sortDefinition.orderBy+`, g.uid ASC
LIMIT ? OFFSET ?
`, rowsArgs...)
	if err != nil {
		return PoliticalGroupsPage{}, fmt.Errorf("query political groups: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var group PoliticalGroupListItem
		if err := rows.Scan(
			&group.UID,
			&group.Libelle,
			&group.LibelleAbrege,
			&group.LibelleAbrev,
			&group.Legislature,
			&group.Position,
			&group.Preseance,
			&group.DateDebut,
			&group.DateFin,
			&group.DeputiesCount,
			&group.TotalScrutins,
			&group.Pour,
			&group.Contre,
			&group.Abstentions,
			&group.NonVotants,
		); err != nil {
			return PoliticalGroupsPage{}, fmt.Errorf("scan political group: %w", err)
		}
		page.Groups = append(page.Groups, group)
	}
	if err := rows.Err(); err != nil {
		return PoliticalGroupsPage{}, fmt.Errorf("iterate political groups: %w", err)
	}

	return page, nil
}

func (s *Store) queryPoliticalGroupFilterOptions(ctx context.Context) (PoliticalGroupFilterOptions, error) {
	legislatures, err := s.intFilterOptions(ctx, `
SELECT DISTINCT legislature
FROM organes
WHERE UPPER(COALESCE(code_type, '')) = 'GP' AND legislature IS NOT NULL
ORDER BY legislature DESC
`)
	if err != nil {
		return PoliticalGroupFilterOptions{}, fmt.Errorf("query political group legislature filter options: %w", err)
	}

	return PoliticalGroupFilterOptions{Legislatures: mapFilterOptions[PoliticalGroupFilterOption](legislatures)}, nil
}

func clonePoliticalGroupFilterOptions(options PoliticalGroupFilterOptions) PoliticalGroupFilterOptions {
	return PoliticalGroupFilterOptions{Legislatures: cloneSlice(options.Legislatures)}
}

func politicalGroupsWhereClause(query PoliticalGroupsQuery) (string, []any) {
	clauses := []string{"UPPER(COALESCE(g.code_type, '')) = 'GP'"}
	args := []any{}

	if query.Search != "" {
		clauses = append(clauses, `LOWER(
  COALESCE(g.uid, '') || ' ' ||
  COALESCE(g.libelle, '') || ' ' ||
  COALESCE(g.libelle_abrege, '') || ' ' ||
  COALESCE(g.libelle_abrev, '') || ' ' ||
  COALESCE(g.libelle_edition, '') || ' ' ||
  COALESCE(g.position_politique, '')
) LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(strings.ToLower(query.Search))+"%")
	}
	if query.Legislature > 0 {
		clauses = append(clauses, "g.legislature = ?")
		args = append(args, query.Legislature)
	}

	return "\nWHERE " + strings.Join(clauses, "\n  AND ") + "\n", args
}

func politicalGroupsListCTE() string {
	return `
WITH member_counts AS (
  SELECT mo.organe_uid AS groupe_uid, COUNT(DISTINCT m.acteur_uid) AS deputies_count
  FROM mandat_organes mo
  JOIN mandats m ON m.uid = mo.mandat_uid
  GROUP BY mo.organe_uid
),
vote_stats AS (
  SELECT groupe_uid, SUM(total_scrutins) AS total_scrutins, SUM(pour) AS pour, SUM(contre) AS contre, SUM(abstentions) AS abstentions, SUM(non_votants) AS non_votants
  FROM groupe_vote_stats
  GROUP BY groupe_uid
)
`
}

func politicalGroupSortOptions() []PoliticalGroupSortOption {
	options := make([]PoliticalGroupSortOption, 0, len(politicalGroupSortDefinitions))
	for _, definition := range politicalGroupSortDefinitions {
		options = append(options, PoliticalGroupSortOption{Value: definition.value, Label: definition.label})
	}
	return options
}

func politicalGroupSortByValue(value string) politicalGroupSortDefinition {
	for _, definition := range politicalGroupSortDefinitions {
		if definition.value == value {
			return definition
		}
	}
	return politicalGroupSortDefinition{}
}
