package featurevalue

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetFeatureValues(ctx context.Context, customerID, productID int64) ([]FeatureValue, error)
	Update(ctx context.Context, tx *sqlx.Tx, fv *FeatureValue) error
}

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return &repo{db: db}
}

func (r *repo) GetFeatureValues(ctx context.Context, customerID, productID int64) ([]FeatureValue, error) {
	var out []FeatureValue
	err := r.db.SelectContext(ctx, &out, getFeatureValuesSQL,
		customerID, productID,
	)
	if err != nil {
		return nil, fmt.Errorf("get feature values: %w", err)
	}
	return out, nil
}

func (r *repo) Update(ctx context.Context, tx *sqlx.Tx, fv *FeatureValue) error {
	_, err := tx.ExecContext(ctx, updateFeatureValueSQL,
		fv.CustomerID,
		fv.ProductID,
		fv.FeatureID,
		fv.FeatureValue,
	)
	if err != nil {
		return fmt.Errorf("update feature value: %w", err)
	}
	return nil
}
