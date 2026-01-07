package featurevalue

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

func (s *Service) GetFeatureValues(ctx context.Context, customerID, productID int64) ([]FeatureValue, error) {
	return s.repo.GetFeatureValues(ctx, customerID, productID)
}

func (s *Service) Update(ctx context.Context, fv *FeatureValue) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Update(ctx, tx, fv)
	})
}
