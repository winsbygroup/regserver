package license

import (
	"context"

	"github.com/jmoiron/sqlx"

	"winsbygroup.com/regserver/internal/product"
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

func (s *Service) Get(ctx context.Context, customerID, productID int64) (*License, error) {
	return s.repo.Get(ctx, customerID, productID)
}

func (s *Service) GetByLicenseKey(ctx context.Context, licenseKey string) (*License, error) {
	return s.repo.GetByLicenseKey(ctx, licenseKey)
}

func (s *Service) GetForCustomer(ctx context.Context, customerID int64) ([]License, error) {
	return s.repo.GetForCustomer(ctx, customerID)
}

func (s *Service) GetUnlicensed(ctx context.Context, customerID int64) ([]product.Product, error) {
	return s.repo.GetUnlicensed(ctx, customerID)
}

func (s *Service) Create(ctx context.Context, lic *License) (*License, error) {
	err := s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Create(ctx, tx, lic)
	})
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Get(ctx, lic.CustomerID, lic.ProductID)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, lic *License) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Update(ctx, tx, lic)
	})
}

func (s *Service) Delete(ctx context.Context, customerID, productID int64) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Delete(ctx, tx, customerID, productID)
	})
}

func (s *Service) GetExpiredLicenses(ctx context.Context, before string) ([]ExpiredLicense, error) {
	return s.repo.GetExpiredLicenses(ctx, before)
}
