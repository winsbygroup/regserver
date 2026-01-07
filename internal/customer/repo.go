package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetAll(ctx context.Context) ([]Customer, error)
	Get(ctx context.Context, id int64) (*Customer, error)
	Create(ctx context.Context, tx *sqlx.Tx, c *Customer) (int64, error)
	Update(ctx context.Context, tx *sqlx.Tx, c *Customer) error
	Delete(ctx context.Context, tx *sqlx.Tx, id int64) error
	Exists(ctx context.Context, id int64) (bool, error)
}

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return &repo{db: db}
}

func (r *repo) GetAll(ctx context.Context) ([]Customer, error) {
	var out []Customer
	err := r.db.SelectContext(ctx, &out, getAllCustomersSQL)
	if err != nil {
		return nil, fmt.Errorf("get all customers: %w", err)
	}
	return out, nil
}

func (r *repo) Get(ctx context.Context, id int64) (*Customer, error) {
	var c Customer
	err := r.db.GetContext(ctx, &c, getCustomerSQL, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("customer not found (%d)", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return &c, nil
}

func (r *repo) Create(ctx context.Context, tx *sqlx.Tx, c *Customer) (int64, error) {
	res, err := tx.ExecContext(ctx, createCustomerSQL,
		c.CustomerName,
		c.ContactName,
		c.Phone,
		c.Email,
		c.Notes,
	)
	if err != nil {
		return 0, fmt.Errorf("create customer: %w", err)
	}
	return res.LastInsertId()
}

func (r *repo) Update(ctx context.Context, tx *sqlx.Tx, c *Customer) error {
	_, err := tx.ExecContext(ctx, updateCustomerSQL,
		c.CustomerName,
		c.ContactName,
		c.Phone,
		c.Email,
		c.Notes,
		c.CustomerID,
	)
	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, tx *sqlx.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, deleteCustomerSQL, id)
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}

func (r *repo) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, customerExistsSQL, id)
	if err != nil {
		return false, fmt.Errorf("customer exists: %w", err)
	}
	return exists, nil
}
