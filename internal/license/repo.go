package license

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"winsbygroup.com/regserver/internal/product"
)

type Repository interface {
	Get(ctx context.Context, customerID, productID int64) (*License, error)
	GetByLicenseKey(ctx context.Context, licenseKey string) (*License, error)
	GetForCustomer(ctx context.Context, customerID int64) ([]License, error)
	GetUnlicensed(ctx context.Context, customerID int64) ([]product.Product, error)
	GetExpiredLicenses(ctx context.Context, before string) ([]ExpiredLicense, error)

	Create(ctx context.Context, tx *sqlx.Tx, lic *License) error
	Update(ctx context.Context, tx *sqlx.Tx, lic *License) error
	Delete(ctx context.Context, tx *sqlx.Tx, customerID, productID int64) error
}

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return &repo{db: db}
}

func (r *repo) Get(ctx context.Context, customerID, productID int64) (*License, error) {
	var lic License
	err := r.db.GetContext(ctx, &lic, getLicenseSQL, customerID, productID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("license not found (%d/%d)", customerID, productID)
	}
	if err != nil {
		return nil, fmt.Errorf("get license: %w", err)
	}
	return &lic, nil
}

func (r *repo) GetByLicenseKey(ctx context.Context, licenseKey string) (*License, error) {
	var lic License
	err := r.db.GetContext(ctx, &lic, getLicenseByKeySQL, strings.ToLower(licenseKey))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("license not found: %s", licenseKey)
	}
	if err != nil {
		return nil, fmt.Errorf("get license by key: %w", err)
	}
	return &lic, nil
}

func (r *repo) GetForCustomer(ctx context.Context, customerID int64) ([]License, error) {
	var out []License
	err := r.db.SelectContext(ctx, &out, getLicensesSQL, customerID)
	if err != nil {
		return nil, fmt.Errorf("get licenses: %w", err)
	}
	return out, nil
}

func (r *repo) GetUnlicensed(ctx context.Context, customerID int64) ([]product.Product, error) {
	var out []product.Product
	err := r.db.SelectContext(ctx, &out, getUnlicensedProductsSQL, customerID)
	if err != nil {
		return nil, fmt.Errorf("get unlicensed products: %w", err)
	}
	return out, nil
}

func (r *repo) Create(ctx context.Context, tx *sqlx.Tx, lic *License) error {
	_, err := tx.ExecContext(ctx, createLicenseSQL,
		lic.CustomerID,
		lic.ProductID,
		strings.ToLower(lic.LicenseKey),
		lic.LicenseCount,
		lic.IsSubscription,
		lic.LicenseTerm,
		lic.StartDate,
		lic.ExpirationDate,
		lic.MaintExpirationDate,
		lic.MaxProductVersion,
	)
	if err != nil {
		return fmt.Errorf("create license: %w", err)
	}
	return nil
}

func (r *repo) Update(ctx context.Context, tx *sqlx.Tx, lic *License) error {
	_, err := tx.ExecContext(ctx, updateLicenseSQL,
		lic.LicenseCount,
		lic.IsSubscription,
		lic.LicenseTerm,
		lic.StartDate,
		lic.ExpirationDate,
		lic.MaintExpirationDate,
		lic.MaxProductVersion,
		lic.CustomerID,
		lic.ProductID,
	)
	if err != nil {
		return fmt.Errorf("update license: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, tx *sqlx.Tx, customerID, productID int64) error {
	_, err := tx.ExecContext(ctx, deleteLicenseSQL, customerID, productID)
	if err != nil {
		return fmt.Errorf("delete license: %w", err)
	}
	return nil
}

func (r *repo) GetExpiredLicenses(ctx context.Context, before string) ([]ExpiredLicense, error) {
	var out []ExpiredLicense
	err := r.db.SelectContext(ctx, &out, getExpiredLicensesSQL, before, before)
	if err != nil {
		return nil, fmt.Errorf("get expired licenses: %w", err)
	}
	return out, nil
}
