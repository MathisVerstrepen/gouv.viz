package store

import (
	"context"
	"fmt"
)

func (s *Store) SitemapScrutinUIDs(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT uid FROM scrutins ORDER BY uid`)
	if err != nil {
		return nil, fmt.Errorf("query scrutin uids: %w", err)
	}
	defer rows.Close()

	var uids []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan scrutin uid: %w", err)
		}
		uids = append(uids, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scrutin uids: %w", err)
	}
	return uids, nil
}

func (s *Store) SitemapDeputyUIDs(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT uid FROM acteurs ORDER BY uid`)
	if err != nil {
		return nil, fmt.Errorf("query deputy uids: %w", err)
	}
	defer rows.Close()

	var uids []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan deputy uid: %w", err)
		}
		uids = append(uids, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deputy uids: %w", err)
	}
	return uids, nil
}

func (s *Store) SitemapGroupUIDs(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT uid FROM organes WHERE UPPER(COALESCE(code_type, '')) = 'GP' ORDER BY uid`)
	if err != nil {
		return nil, fmt.Errorf("query group uids: %w", err)
	}
	defer rows.Close()

	var uids []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan group uid: %w", err)
		}
		uids = append(uids, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group uids: %w", err)
	}
	return uids, nil
}
