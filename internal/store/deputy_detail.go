package store

import (
	"context"
	"fmt"
	"math"
	"strings"
)

const DeputyVotesPerPage = 50

type deputyVoteSortDefinition struct {
	value   string
	orderBy string
}

var deputyVoteSortDefinitions = []deputyVoteSortDefinition{
	{value: "date_desc", orderBy: "s.date_scrutin DESC, s.numero DESC, s.uid"},
	{value: "date_asc", orderBy: "s.date_scrutin ASC, s.numero ASC, s.uid"},
}

func NormalizeDeputyDetailQuery(query DeputyDetailQuery) DeputyDetailQuery {
	if query.VotesPage < 1 {
		query.VotesPage = 1
	}
	if query.VotesPerPage < 1 {
		query.VotesPerPage = DeputyVotesPerPage
	}
	query.VotesSearch = strings.TrimSpace(query.VotesSearch)
	query.VotesPosition = strings.TrimSpace(query.VotesPosition)
	if deputyVoteSortByValue(query.VotesSort).value == "" {
		query.VotesSort = deputyVoteSortDefinitions[0].value
	}
	return query
}

func (s *Store) DeputyDetailPage(ctx context.Context, uid string, query DeputyDetailQuery) (DeputyDetailPage, error) {
	query = NormalizeDeputyDetailQuery(query)
	var page DeputyDetailPage
	page.Query = query
	var deputy DeputyDetailData
	if err := s.db.QueryRowContext(ctx, `
SELECT
  uid,
  COALESCE(civilite, ''),
  COALESCE(prenom, ''),
  COALESCE(nom, ''),
  COALESCE(alpha, ''),
  COALESCE(NULLIF(TRIM(COALESCE(prenom, '') || ' ' || COALESCE(nom, '')), ''), alpha, uid, ''),
  COALESCE(date_naissance, ''),
  COALESCE(ville_naissance, ''),
  COALESCE(dep_naissance, ''),
  COALESCE(pays_naissance, ''),
  COALESCE(date_deces, ''),
  COALESCE(profession, ''),
  COALESCE(uri_hatvp, ''),
  COALESCE(source_file, '')
FROM acteurs
WHERE uid = ?
`, uid).Scan(
		&deputy.UID,
		&deputy.Civilite,
		&deputy.Prenom,
		&deputy.Nom,
		&deputy.Alpha,
		&deputy.DisplayName,
		&deputy.DateNaissance,
		&deputy.VilleNaissance,
		&deputy.DepNaissance,
		&deputy.PaysNaissance,
		&deputy.DateDeces,
		&deputy.Profession,
		&deputy.URIHATVP,
		&deputy.SourceFile,
	); err != nil {
		return DeputyDetailPage{}, fmt.Errorf("query deputy: %w", err)
	}
	page.Deputy = deputy

	mandats, err := s.deputyMandats(ctx, uid)
	if err != nil {
		return DeputyDetailPage{}, err
	}
	page.Mandats = mandats

	stats, err := s.deputyVoteStats(ctx, uid)
	if err != nil {
		return DeputyDetailPage{}, err
	}
	page.Stats = stats

	votesTotal, err := s.deputyVotesCount(ctx, uid, query)
	if err != nil {
		return DeputyDetailPage{}, err
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

	votes, err := s.deputyVotes(ctx, uid, page.Query)
	if err != nil {
		return DeputyDetailPage{}, err
	}
	page.Votes = votes

	return page, nil
}

func (s *Store) deputyMandats(ctx context.Context, uid string) ([]DeputyMandat, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
  m.uid,
  COALESCE(m.legislature, 0),
  COALESCE(m.type_organe, ''),
  COALESCE(m.date_debut, ''),
  COALESCE(m.date_fin, ''),
  COALESCE(m.date_publication, ''),
  COALESCE(m.preseance, 0),
  COALESCE(m.nomin_principale, 0),
  COALESCE(m.code_qualite, ''),
  COALESCE(m.lib_qualite, ''),
  COALESCE(m.lib_qualite_sex, '')
FROM mandats m
WHERE m.acteur_uid = ?
ORDER BY COALESCE(m.date_debut, '') DESC, COALESCE(m.preseance, 9999), m.uid
`, uid)
	if err != nil {
		return nil, fmt.Errorf("query deputy mandats: %w", err)
	}
	defer rows.Close()

	mandats := make([]DeputyMandat, 0)
	mandatIndexes := make(map[string]int)
	for rows.Next() {
		var mandat DeputyMandat
		var nominPrincipale int
		if err := rows.Scan(
			&mandat.UID,
			&mandat.Legislature,
			&mandat.TypeOrgane,
			&mandat.DateDebut,
			&mandat.DateFin,
			&mandat.DatePublication,
			&mandat.Preseance,
			&nominPrincipale,
			&mandat.CodeQualite,
			&mandat.LibQualite,
			&mandat.LibQualiteSex,
		); err != nil {
			return nil, fmt.Errorf("scan deputy mandat: %w", err)
		}
		mandat.NominPrincipale = nominPrincipale == 1
		mandatIndexes[mandat.UID] = len(mandats)
		mandats = append(mandats, mandat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deputy mandats: %w", err)
	}

	organes, err := s.deputyMandatOrganes(ctx, uid)
	if err != nil {
		return nil, err
	}
	for mandatUID, values := range organes {
		if index, ok := mandatIndexes[mandatUID]; ok {
			mandats[index].Organes = values
		}
	}

	return mandats, nil
}

func (s *Store) deputyMandatOrganes(ctx context.Context, uid string) (map[string][]DeputyMandatOrgane, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
  mo.mandat_uid,
  mo.organe_uid,
  COALESCE(o.libelle_abrege, o.libelle, mo.organe_uid, '')
FROM mandats m
JOIN mandat_organes mo ON mo.mandat_uid = m.uid
LEFT JOIN organes o ON o.uid = mo.organe_uid
WHERE m.acteur_uid = ?
ORDER BY mo.mandat_uid, COALESCE(o.preseance, 9999), COALESCE(o.libelle, o.libelle_abrege, mo.organe_uid)
`, uid)
	if err != nil {
		return nil, fmt.Errorf("query deputy mandat organes: %w", err)
	}
	defer rows.Close()

	organes := make(map[string][]DeputyMandatOrgane)
	for rows.Next() {
		var mandatUID string
		var organe DeputyMandatOrgane
		if err := rows.Scan(&mandatUID, &organe.UID, &organe.Libelle); err != nil {
			return nil, fmt.Errorf("scan deputy mandat organe: %w", err)
		}
		organes[mandatUID] = append(organes[mandatUID], organe)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deputy mandat organes: %w", err)
	}

	return organes, nil
}

func (s *Store) deputyVoteStats(ctx context.Context, uid string) ([]DeputyVoteStat, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT legislature, total_votes, pour, contre, abstentions, non_votants
FROM acteur_vote_stats
WHERE acteur_uid = ?
ORDER BY legislature DESC
`, uid)
	if err != nil {
		return nil, fmt.Errorf("query deputy vote stats: %w", err)
	}
	defer rows.Close()

	var stats []DeputyVoteStat
	for rows.Next() {
		var stat DeputyVoteStat
		if err := rows.Scan(&stat.Legislature, &stat.TotalVotes, &stat.Pour, &stat.Contre, &stat.Abstentions, &stat.NonVotants); err != nil {
			return nil, fmt.Errorf("scan deputy vote stat: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deputy vote stats: %w", err)
	}

	return stats, nil
}

func (s *Store) deputyVotesCount(ctx context.Context, uid string, query DeputyDetailQuery) (int, error) {
	whereClause, whereArgs := deputyVotesWhereClause(uid, query)
	var total int
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
LEFT JOIN organes o ON o.uid = s.organe_uid
LEFT JOIN organes g ON g.uid = v.groupe_uid
`+whereClause, whereArgs...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count deputy votes: %w", err)
	}
	return total, nil
}

func (s *Store) deputyVotes(ctx context.Context, uid string, query DeputyDetailQuery) ([]DeputyVote, error) {
	whereClause, whereArgs := deputyVotesWhereClause(uid, query)
	sortDefinition := deputyVoteSortByValue(query.VotesSort)
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
  COALESCE(o.libelle_abrege, o.libelle, s.organe_uid, ''),
  COALESCE(v.groupe_uid, ''),
  COALESCE(g.libelle_abrege, g.libelle, v.groupe_uid, ''),
  COALESCE(v.mandat_uid, ''),
  v.position,
  COALESCE(v.par_delegation, 0),
  COALESCE(v.num_place, '')
FROM votes v
JOIN scrutins s ON s.uid = v.scrutin_uid
LEFT JOIN organes o ON o.uid = s.organe_uid
LEFT JOIN organes g ON g.uid = v.groupe_uid
`+whereClause+`
ORDER BY `+sortDefinition.orderBy+`
LIMIT ? OFFSET ?
`, rowsArgs...)
	if err != nil {
		return nil, fmt.Errorf("query deputy votes: %w", err)
	}
	defer rows.Close()

	var votes []DeputyVote
	for rows.Next() {
		var vote DeputyVote
		var parDelegation int
		if err := rows.Scan(
			&vote.ScrutinUID,
			&vote.Numero,
			&vote.Legislature,
			&vote.Date,
			&vote.Titre,
			&vote.SortCode,
			&vote.SortLibelle,
			&vote.TypeVote,
			&vote.Organe,
			&vote.GroupeUID,
			&vote.Groupe,
			&vote.MandatUID,
			&vote.Position,
			&parDelegation,
			&vote.NumPlace,
		); err != nil {
			return nil, fmt.Errorf("scan deputy vote: %w", err)
		}
		vote.ParDelegation = parDelegation == 1
		votes = append(votes, vote)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deputy votes: %w", err)
	}

	return votes, nil
}

func deputyVotesWhereClause(uid string, query DeputyDetailQuery) (string, []any) {
	clauses := []string{"v.acteur_uid = ?"}
	args := []any{uid}
	if query.VotesPosition != "" {
		clauses = append(clauses, "v.position = ?")
		args = append(args, query.VotesPosition)
	}
	if query.VotesSearch != "" {
		clauses = append(clauses, `LOWER(
  COALESCE(s.titre, '') || ' ' ||
  COALESCE(s.numero, '') || ' ' ||
  COALESCE(s.sort_code, '') || ' ' ||
  COALESCE(s.sort_libelle, '') || ' ' ||
  COALESCE(s.libelle_type_vote, '') || ' ' ||
  COALESCE(o.libelle_abrege, '') || ' ' ||
  COALESCE(o.libelle, '') || ' ' ||
  COALESCE(g.libelle_abrege, '') || ' ' ||
  COALESCE(g.libelle, '')
) LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(strings.ToLower(query.VotesSearch))+"%")
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func deputyVoteSortByValue(value string) deputyVoteSortDefinition {
	for _, definition := range deputyVoteSortDefinitions {
		if definition.value == value {
			return definition
		}
	}
	return deputyVoteSortDefinition{}
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
}
