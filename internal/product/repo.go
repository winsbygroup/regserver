package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetAll(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id int64) (*Product, error)
	GetByGUID(ctx context.Context, guid string) (*Product, error)
	Create(ctx context.Context, tx *sqlx.Tx, p *Product) (int64, error)
	Update(ctx context.Context, tx *sqlx.Tx, p *Product) error
	Delete(ctx context.Context, tx *sqlx.Tx, id int64) error
}

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return &repo{db: db}
}

func (r *repo) GetAll(ctx context.Context) ([]Product, error) {
	var out []Product
	err := r.db.SelectContext(ctx, &out, getAllProductsSQL)
	if err != nil {
		return nil, fmt.Errorf("get all products: %w", err)
	}
	return out, nil
}

func (r *repo) Get(ctx context.Context, id int64) (*Product, error) {
	var p Product
	err := r.db.GetContext(ctx, &p, getProductSQL, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("product not found (%d)", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	return &p, nil
}

func (r *repo) GetByGUID(ctx context.Context, guid string) (*Product, error) {
	var p Product
	err := r.db.GetContext(ctx, &p, getProductByGUIDSQL, guid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("product not found (%s)", guid)
	}
	if err != nil {
		return nil, fmt.Errorf("get product by GUID: %w", err)
	}
	return &p, nil
}

func (r *repo) Create(ctx context.Context, tx *sqlx.Tx, p *Product) (int64, error) {
	res, err := tx.ExecContext(ctx, createProductSQL,
		p.ProductName,
		strings.ToLower(p.ProductGUID),
		p.LatestVersion,
		p.DownloadURL,
	)
	if err != nil {
		return 0, fmt.Errorf("create product: %w", err)
	}
	return res.LastInsertId()
}

func (r *repo) Update(ctx context.Context, tx *sqlx.Tx, p *Product) error {
	_, err := tx.ExecContext(ctx, updateProductSQL,
		p.ProductName,
		strings.ToLower(p.ProductGUID),
		p.LatestVersion,
		p.DownloadURL,
		p.ProductID,
	)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, tx *sqlx.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, deleteProductSQL, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}
