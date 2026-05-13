package store

import (
	"context"
	"fmt"
)

func (s *Store) HomePage(ctx context.Context) (HomePage, error) {
	cache, err := s.staticCache(ctx)
	if err != nil {
		return HomePage{}, err
	}
	return cloneHomePage(cache.homePage), nil
}

func (s *Store) queryHomePage(ctx context.Context) (HomePage, error) {
	var page HomePage
	if err := s.db.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*) FROM scrutins),
  (SELECT COUNT(*) FROM acteurs),
  (SELECT COUNT(*) FROM organes WHERE UPPER(COALESCE(code_type, '')) = 'GP'),
  COALESCE((SELECT MIN(date_scrutin) FROM scrutins), ''),
  COALESCE((SELECT MAX(date_scrutin) FROM scrutins), '')
`).Scan(&page.TotalScrutins, &page.TotalDeputies, &page.TotalGroups, &page.FirstScrutinDate, &page.LastScrutinDate); err != nil {
		return HomePage{}, fmt.Errorf("query scrutin totals: %w", err)
	}

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
ORDER BY s.date_scrutin DESC, s.numero DESC
LIMIT 50
`)
	if err != nil {
		return HomePage{}, fmt.Errorf("query scrutins: %w", err)
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
			return HomePage{}, fmt.Errorf("scan scrutin: %w", err)
		}
		page.Scrutins = append(page.Scrutins, scrutin)
	}
	if err := rows.Err(); err != nil {
		return HomePage{}, fmt.Errorf("iterate scrutins: %w", err)
	}

	return page, nil
}

func cloneHomePage(page HomePage) HomePage {
	page.Scrutins = cloneSlice(page.Scrutins)
	return page
}
