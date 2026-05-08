package handlers

import "gouv.viz/internal/store"

type Server struct {
	store *store.Store
}

func NewServer(store *store.Store) *Server {
	return &Server{store: store}
}
