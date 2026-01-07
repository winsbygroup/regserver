package feature

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

func (s *Service) GetForProduct(ctx context.Context, productID int64) ([]Feature, error) {
	return s.repo.GetForProduct(ctx, productID)
}

func (s *Service) Get(ctx context.Context, id int64) (*Feature, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, f *Feature) (*Feature, error) {
	var id int64

	err := s.WithTx(ctx, func(tx *sqlx.Tx) error {
		var err error
		id, err = s.repo.Create(ctx, tx, f)
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

func (s *Service) Update(ctx context.Context, f *Feature) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Update(ctx, tx, f)
	})
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Delete(ctx, tx, id)
	})
}
