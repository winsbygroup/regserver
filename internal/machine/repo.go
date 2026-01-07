package machine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetByID(ctx context.Context, machineID int64) (*Machine, error)
	GetByCode(ctx context.Context, customerID int64, machineCode string) (*Machine, error)
	Create(ctx context.Context, tx *sqlx.Tx, m *Machine) (int64, error)
	UpdateUserName(ctx context.Context, tx *sqlx.Tx, machineID int64, userName string) error
	GetForLicense(ctx context.Context, customerID, productID int64) ([]Machine, error)
	GetActiveForLicense(ctx context.Context, customerID, productID int64) ([]Machine, error)
}

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return &repo{db: db}
}

func (r *repo) GetByID(ctx context.Context, machineID int64) (*Machine, error) {
	var m Machine
	err := r.db.GetContext(ctx, &m, getMachineByIDSQL, machineID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get machine by id: %w", err)
	}
	return &m, nil
}

func (r *repo) GetByCode(ctx context.Context, customerID int64, machineCode string) (*Machine, error) {
	var m Machine
	err := r.db.GetContext(ctx, &m, getMachineSQL, customerID, machineCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get machine: %w", err)
	}
	return &m, nil
}

func (r *repo) Create(ctx context.Context, tx *sqlx.Tx, m *Machine) (int64, error) {
	res, err := tx.ExecContext(ctx, createMachineSQL,
		m.CustomerID,
		m.MachineCode,
		m.UserName,
	)
	if err != nil {
		return 0, fmt.Errorf("create machine: %w", err)
	}
	return res.LastInsertId()
}

func (r *repo) UpdateUserName(ctx context.Context, tx *sqlx.Tx, machineID int64, userName string) error {
	_, err := tx.ExecContext(ctx, updateUserNameSQL, userName, machineID)
	if err != nil {
		return fmt.Errorf("update machine user_name: %w", err)
	}
	return nil
}

func (r *repo) GetForLicense(ctx context.Context, customerID, productID int64) ([]Machine, error) {
	var machines []Machine
	err := r.db.SelectContext(ctx, &machines, getForLicenseSQL, customerID, productID)
	return machines, err
}

func (r *repo) GetActiveForLicense(ctx context.Context, customerID, productID int64) ([]Machine, error) {
	var machines []Machine
	err := r.db.SelectContext(ctx, &machines, getActiveForLicenseSQL, customerID, productID)
	return machines, err
}
