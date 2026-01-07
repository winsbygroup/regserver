package customer

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo Repository
	db   *sqlx.DB
}

func NewService(db *sqlx.DB) *Service {
	return &Service{
		db:   db,
		repo: New(db),
	}
}

func (s *Service) WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *Service) GetAll(ctx context.Context) ([]Customer, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) Get(ctx context.Context, id int64) (*Customer, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, c *Customer) (*Customer, error) {
	var id int64
	err := s.WithTx(ctx, func(tx *sqlx.Tx) error {
		var err error
		id, err = s.repo.Create(ctx, tx, c)
		return err
	})
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, c *Customer) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Update(ctx, tx, c)
	})
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Delete(ctx, tx, id)
	})
}

func (s *Service) Exists(ctx context.Context, id int64) (bool, error) {
	return s.repo.Exists(ctx, id)
}
