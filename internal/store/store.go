package store

import "database/sql"

var ErrNotFound = sql.ErrNoRows

// Store owns read access to the generated SQLite database.
type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}
