package registration

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

func (s *Service) Get(ctx context.Context, machineID, productID int64) (*Registration, error) {
	return s.repo.Get(ctx, machineID, productID)
}

func (s *Service) GetForMachine(ctx context.Context, machineID int64) ([]Registration, error) {
	return s.repo.GetForMachine(ctx, machineID)
}

func (s *Service) Create(ctx context.Context, r *Registration) (*Registration, error) {
	err := s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Create(ctx, tx, r)
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Service) Update(ctx context.Context, reg *Registration) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Update(ctx, tx, reg)
	})
}

func (s *Service) Upsert(ctx context.Context, tx *sqlx.Tx, reg *Registration) error {
	return s.repo.Upsert(ctx, tx, reg)
}

func (s *Service) Delete(ctx context.Context, machineID, productID int64) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Delete(ctx, tx, machineID, productID)
	})
}

func (s *Service) UpdateInstalledVersion(ctx context.Context, machineID, productID int64, version string) error {
	return s.repo.UpdateInstalledVersion(ctx, machineID, productID, version)
}
