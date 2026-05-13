package store

import (
	"context"
	"database/sql"
	"sync"
)

var ErrNotFound = sql.ErrNoRows

// Store owns read access to the generated SQLite database.
type Store struct {
	db *sql.DB

	cacheMu     sync.Mutex
	cacheLoaded bool
	cache       staticStoreCache
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

type staticStoreCache struct {
	homePage                    HomePage
	scrutinFilterOptions        ScrutinFilterOptions
	deputyFilterOptions         DeputyFilterOptions
	politicalGroupFilterOptions PoliticalGroupFilterOptions
}

func (s *Store) staticCache(ctx context.Context) (staticStoreCache, error) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	if s.cacheLoaded {
		return s.cache, nil
	}

	cache, err := s.loadStaticCache(ctx)
	if err != nil {
		return staticStoreCache{}, err
	}
	s.cache = cache
	s.cacheLoaded = true
	return cache, nil
}

func (s *Store) loadStaticCache(ctx context.Context) (staticStoreCache, error) {
	homePage, err := s.queryHomePage(ctx)
	if err != nil {
		return staticStoreCache{}, err
	}
	scrutinFilterOptions, err := s.queryScrutinFilterOptions(ctx)
	if err != nil {
		return staticStoreCache{}, err
	}
	deputyFilterOptions, err := s.queryDeputyFilterOptions(ctx)
	if err != nil {
		return staticStoreCache{}, err
	}
	politicalGroupFilterOptions, err := s.queryPoliticalGroupFilterOptions(ctx)
	if err != nil {
		return staticStoreCache{}, err
	}

	return staticStoreCache{
		homePage:                    homePage,
		scrutinFilterOptions:        scrutinFilterOptions,
		deputyFilterOptions:         deputyFilterOptions,
		politicalGroupFilterOptions: politicalGroupFilterOptions,
	}, nil
}

func cloneSlice[S ~[]E, E any](values S) S {
	if values == nil {
		return nil
	}
	cloned := make(S, len(values))
	copy(cloned, values)
	return cloned
}
