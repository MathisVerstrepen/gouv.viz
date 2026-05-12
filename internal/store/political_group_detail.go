package store

import (
	"context"
	"fmt"
	"math"
	"strings"
)

const PoliticalGroupVotesPerPage = 50

type politicalGroupVoteSortDefinition struct {
	value   string
	orderBy string
}

var politicalGroupVoteSortDefinitions = []politicalGroupVoteSortDefinition{
	{value: "date_desc", orderBy: "s.date_scrutin DESC, s.numero DESC, s.uid"},
	{value: "date_asc", orderBy: "s.date_scrutin ASC, s.numero ASC, s.uid"},
}

func NormalizePoliticalGroupDetailQuery(query PoliticalGroupDetailQuery) PoliticalGroupDetailQuery {
	if query.VotesPage < 1 {
		query.VotesPage = 1
	}
	if query.VotesPerPage < 1 {
		query.VotesPerPage = PoliticalGroupVotesPerPage
	}
	query.VotesSearch = strings.TrimSpace(query.VotesSearch)
	query.VotesPosition = strings.TrimSpace(query.VotesPosition)
	if politicalGroupVoteSortByValue(query.VotesSort).value == "" {
		query.VotesSort = politicalGroupVoteSortDefinitions[0].value
	}
	return query
}

func (s *Store) PoliticalGroupDetailPage(ctx context.Context, uid string, query PoliticalGroupDetailQuery) (PoliticalGroupDetailPage, error) {
	query = NormalizePoliticalGroupDetailQuery(query)
	page := PoliticalGroupDetailPage{Query: query}

	if err := s.db.QueryRowContext(ctx, `
SELECT
  uid,
  COALESCE(code_type, ''),
  COALESCE(libelle, ''),
  COALESCE(libelle_abrege, ''),
  COALESCE(libelle_abrev, ''),
  COALESCE(libelle_edition, ''),
  COALESCE(legislature, 0),
  COALESCE(chambre, ''),
  COALESCE(regime, ''),
  COALESCE(position_politique, ''),
  COALESCE(couleur_associee, ''),
  COALESCE(preseance, 0),
  COALESCE(date_debut, ''),
  COALESCE(date_fin, ''),
  COALESCE(source_file, '')
FROM organes
WHERE uid = ? AND UPPER(COALESCE(code_type, '')) = 'GP'
`, uid).Scan(
		&page.Group.UID,
		&page.Group.CodeType,
		&page.Group.Libelle,
		&page.Group.LibelleAbrege,
		&page.Group.LibelleAbrev,
		&page.Group.LibelleEdition,
		&page.Group.Legislature,
		&page.Group.Chambre,
		&page.Group.Regime,
		&page.Group.PositionPolitique,
		&page.Group.CouleurAssociee,
		&page.Group.Preseance,
		&page.Group.DateDebut,
		&page.Group.DateFin,
		&page.Group.SourceFile,
	); err != nil {
		return PoliticalGroupDetailPage{}, fmt.Errorf("query political group: %w", err)
	}

	stats, err := s.politicalGroupVoteStats(ctx, uid)
	if err != nil {
		return PoliticalGroupDetailPage{}, err
	}
	page.Stats = stats

	deputies, err := s.politicalGroupDeputies(ctx, uid)
	if err != nil {
		return PoliticalGroupDetailPage{}, err
	}
	page.Deputies = deputies

	votesTotal, err := s.politicalGroupVotesCount(ctx, uid, query)
	if err != nil {
		return PoliticalGroupDetailPage{}, err
	}
	page.VotesTotalResults = votesTotal
	if page.VotesTotalResults > 0 {
		page.VotesTotalPages = int(math.Ceil(float64(page.VotesTotalResults) / float64(query.VotesPerPage)))
		if page.Query.VotesPage > page.VotesTotalPages {
			page.Query.VotesPage = page.VotesTotalPages
		}
		page.VotesStartItem = ((page.Query.VotesPage - 1) * query.VotesPerPage) + 1
		page.VotesEndItem = page.VotesStartItem + query.VotesPerPage - 1
		if page.VotesEndItem > page.VotesTotalResults {
			page.VotesEndItem = page.VotesTotalResults
		}
	} else {
		page.Query.VotesPage = 1
		page.VotesTotalPages = 1
	}

	votes, err := s.politicalGroupVotes(ctx, uid, page.Query)
	if err != nil {
		return PoliticalGroupDetailPage{}, err
	}
	page.Votes = votes

	return page, nil
}

func (s *Store) politicalGroupVoteStats(ctx context.Context, uid string) ([]PoliticalGroupVoteStat, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT legislature, total_scrutins, pour, contre, abstentions, non_votants
FROM groupe_vote_stats
WHERE groupe_uid = ?
ORDER BY legislature DESC
`, uid)
	if err != nil {
		return nil, fmt.Errorf("query political group vote stats: %w", err)
	}
	defer rows.Close()

	var stats []PoliticalGroupVoteStat
	for rows.Next() {
		var stat PoliticalGroupVoteStat
		if err := rows.Scan(&stat.Legislature, &stat.TotalScrutins, &stat.Pour, &stat.Contre, &stat.Abstentions, &stat.NonVotants); err != nil {
			return nil, fmt.Errorf("scan political group vote stat: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate political group vote stats: %w", err)
	}
	return stats, nil
}

func (s *Store) politicalGroupDeputies(ctx context.Context, uid string) ([]PoliticalGroupDeputy, error) {
	rows, err := s.db.QueryContext(ctx, `
WITH group_mandats AS (
  SELECT
    m.uid,
    m.acteur_uid,
    m.legislature,
    m.date_debut,
    m.date_fin,
    COALESCE(m.lib_qualite, m.lib_qualite_sex, m.code_qualite, '') AS qualite,
    ROW_NUMBER() OVER (
      PARTITION BY m.acteur_uid
      ORDER BY COALESCE(m.date_debut, '') DESC, COALESCE(m.preseance, 9999), m.uid
    ) AS rn
  FROM mandats m
  JOIN mandat_organes mo ON mo.mandat_uid = m.uid
  WHERE mo.organe_uid = ?
)
SELECT
  a.uid,
  COALESCE(NULLIF(TRIM(COALESCE(a.prenom, '') || ' ' || COALESCE(a.nom, '')), ''), a.alpha, a.uid, ''),
  COALESCE(a.alpha, ''),
  COALESCE(gm.legislature, 0),
  gm.uid,
  COALESCE(gm.date_debut, ''),
  COALESCE(gm.date_fin, ''),
  COALESCE(gm.qualite, '')
FROM group_mandats gm
JOIN acteurs a ON a.uid = gm.acteur_uid
WHERE gm.rn = 1
ORDER BY COALESCE(NULLIF(a.alpha, ''), a.nom, a.prenom, a.uid), a.uid
`, uid)
	if err != nil {
		return nil, fmt.Errorf("query political group deputies: %w", err)
	}
	defer rows.Close()

	var deputies []PoliticalGroupDeputy
	for rows.Next() {
		var deputy PoliticalGroupDeputy
		if err := rows.Scan(&deputy.UID, &deputy.DisplayName, &deputy.Alpha, &deputy.Legislature, &deputy.MandatUID, &deputy.DateDebut, &deputy.DateFin, &deputy.Qualite); err != nil {
			return nil, fmt.Errorf("scan political group deputy: %w", err)
		}
		deputies = append(deputies, deputy)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate political group deputies: %w", err)
	}
	return deputies, nil
}

func (s *Store) politicalGroupVotesCount(ctx context.Context, uid string, query PoliticalGroupDetailQuery) (int, error) {
	whereClause, whereArgs := politicalGroupVotesWhereClause(uid, query)
	var total int
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM scrutin_groupe_votes sgv
JOIN scrutins s ON s.uid = sgv.scrutin_uid
`+whereClause, whereArgs...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count political group votes: %w", err)
	}
	return total, nil
}

func (s *Store) politicalGroupVotes(ctx context.Context, uid string, query PoliticalGroupDetailQuery) ([]PoliticalGroupVote, error) {
	whereClause, whereArgs := politicalGroupVotesWhereClause(uid, query)
	sortDefinition := politicalGroupVoteSortByValue(query.VotesSort)
	rowsArgs := append([]any{}, whereArgs...)
	rowsArgs = append(rowsArgs, query.VotesPerPage, (query.VotesPage-1)*query.VotesPerPage)
	rows, err := s.db.QueryContext(ctx, `
SELECT
  s.uid,
  s.numero,
  s.legislature,
  s.date_scrutin,
  COALESCE(s.titre, ''),
  COALESCE(s.sort_code, ''),
  COALESCE(s.sort_libelle, ''),
  COALESCE(s.libelle_type_vote, ''),
  COALESCE(sgv.position_majoritaire, ''),
  COALESCE(sgv.nombre_membres_groupe, 0),
  COALESCE(sgv.pour, 0),
  COALESCE(sgv.contre, 0),
  COALESCE(sgv.abstentions, 0),
  COALESCE(sgv.non_votants, 0),
  COALESCE(sgv.non_votants_volontaires, 0)
FROM scrutin_groupe_votes sgv
JOIN scrutins s ON s.uid = sgv.scrutin_uid
`+whereClause+`
ORDER BY `+sortDefinition.orderBy+`
LIMIT ? OFFSET ?
`, rowsArgs...)
	if err != nil {
		return nil, fmt.Errorf("query political group votes: %w", err)
	}
	defer rows.Close()

	var votes []PoliticalGroupVote
	for rows.Next() {
		var vote PoliticalGroupVote
		if err := rows.Scan(
			&vote.ScrutinUID,
			&vote.Numero,
			&vote.Legislature,
			&vote.Date,
			&vote.Titre,
			&vote.SortCode,
			&vote.SortLibelle,
			&vote.TypeVote,
			&vote.PositionMajoritaire,
			&vote.NombreMembres,
			&vote.Pour,
			&vote.Contre,
			&vote.Abstentions,
			&vote.NonVotants,
			&vote.NonVotantsVolontaires,
		); err != nil {
			return nil, fmt.Errorf("scan political group vote: %w", err)
		}
		votes = append(votes, vote)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate political group votes: %w", err)
	}
	return votes, nil
}

func politicalGroupVotesWhereClause(uid string, query PoliticalGroupDetailQuery) (string, []any) {
	clauses := []string{"sgv.groupe_uid = ?"}
	args := []any{uid}
	if query.VotesPosition != "" {
		clauses = append(clauses, "sgv.position_majoritaire = ?")
		args = append(args, query.VotesPosition)
	}
	if query.VotesSearch != "" {
		clauses = append(clauses, `LOWER(
  COALESCE(s.titre, '') || ' ' ||
  COALESCE(s.numero, '') || ' ' ||
  COALESCE(s.sort_code, '') || ' ' ||
  COALESCE(s.sort_libelle, '') || ' ' ||
  COALESCE(s.libelle_type_vote, '') || ' ' ||
  COALESCE(s.objet_libelle, '') || ' ' ||
  COALESCE(s.demandeur_texte, '')
) LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(strings.ToLower(query.VotesSearch))+"%")
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func politicalGroupVoteSortByValue(value string) politicalGroupVoteSortDefinition {
	for _, definition := range politicalGroupVoteSortDefinitions {
		if definition.value == value {
			return definition
		}
	}
	return politicalGroupVoteSortDefinition{}
}
