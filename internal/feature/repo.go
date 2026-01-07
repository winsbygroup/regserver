package feature

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetForProduct(ctx context.Context, productID int64) ([]Feature, error)
	Get(ctx context.Context, id int64) (*Feature, error)
	Create(ctx context.Context, tx *sqlx.Tx, f *Feature) (int64, error)
	Update(ctx context.Context, tx *sqlx.Tx, f *Feature) error
	Delete(ctx context.Context, tx *sqlx.Tx, id int64) error
}

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return &repo{db: db}
}

func (r *repo) GetForProduct(ctx context.Context, productID int64) ([]Feature, error) {
	var out []Feature
	err := r.db.SelectContext(ctx, &out, getFeaturesForProductSQL, productID)
	if err != nil {
		return nil, fmt.Errorf("get features for product: %w", err)
	}
	return out, nil
}

func (r *repo) Get(ctx context.Context, id int64) (*Feature, error) {
	var f Feature
	err := r.db.GetContext(ctx, &f, getFeatureSQL, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("feature not found (%d)", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get feature: %w", err)
	}
	return &f, nil
}

func (r *repo) Create(ctx context.Context, tx *sqlx.Tx, f *Feature) (int64, error) {
	res, err := tx.ExecContext(ctx, createFeatureSQL,
		f.ProductID,
		f.FeatureName,
		f.FeatureType,
		f.AllowedValues,
		f.DefaultValue,
	)
	if err != nil {
		return 0, fmt.Errorf("create feature: %w", err)
	}
	return res.LastInsertId()
}

func (r *repo) Update(ctx context.Context, tx *sqlx.Tx, f *Feature) error {
	_, err := tx.ExecContext(ctx, updateFeatureSQL,
		f.FeatureName,
		f.FeatureType,
		f.AllowedValues,
		f.DefaultValue,
		f.FeatureID,
	)
	if err != nil {
		return fmt.Errorf("update feature: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, tx *sqlx.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, deleteFeatureSQL, id)
	if err != nil {
		return fmt.Errorf("delete feature: %w", err)
	}
	return nil
}
