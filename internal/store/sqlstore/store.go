package sqlstore

import (
	"effectiveMobile/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db              *pgxpool.Pool
	humanRepository *HumanRepository
}

func New(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Human() store.HumanRepository {
	if s.humanRepository != nil {
		return s.humanRepository
	}
	s.humanRepository = &HumanRepository{
		store: s,
	}
	return s.humanRepository
}
