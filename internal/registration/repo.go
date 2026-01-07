package registration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Get(ctx context.Context, machineID, productID int64) (*Registration, error)
	GetForMachine(ctx context.Context, machineID int64) ([]Registration, error)
	Create(ctx context.Context, tx *sqlx.Tx, r *Registration) error
	Update(ctx context.Context, tx *sqlx.Tx, r *Registration) error
	Upsert(ctx context.Context, tx *sqlx.Tx, r *Registration) error
	Delete(ctx context.Context, tx *sqlx.Tx, machineID, productID int64) error
	UpdateInstalledVersion(ctx context.Context, machineID, productID int64, version string) error
}

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	return &repo{db: db}
}

func (r *repo) Get(ctx context.Context, machineID, productID int64) (*Registration, error) {
	var reg Registration
	err := r.db.GetContext(ctx, &reg, getRegistrationSQL, machineID, productID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("registration not found (%d/%d)", machineID, productID)
	}
	if err != nil {
		return nil, fmt.Errorf("get registration: %w", err)
	}
	return &reg, nil
}

func (r *repo) GetForMachine(ctx context.Context, machineID int64) ([]Registration, error) {
	var out []Registration
	err := r.db.SelectContext(ctx, &out, getRegistrationsForMachineSQL, machineID)
	if err != nil {
		return nil, fmt.Errorf("get registrations for machine: %w", err)
	}
	return out, nil
}

func (r *repo) Create(ctx context.Context, tx *sqlx.Tx, reg *Registration) error {
	_, err := tx.ExecContext(ctx, createRegistrationSQL,
		reg.MachineID,
		reg.ProductID,
		reg.ExpirationDate,
		reg.RegistrationHash,
		reg.FirstRegistrationDate,
		reg.LastRegistrationDate,
	)
	if err != nil {
		return fmt.Errorf("create registration: %w", err)
	}
	return nil
}

func (r *repo) Update(ctx context.Context, tx *sqlx.Tx, reg *Registration) error {
	_, err := tx.ExecContext(ctx, updateRegistrationSQL,
		reg.ExpirationDate,
		reg.RegistrationHash,
		reg.FirstRegistrationDate,
		reg.LastRegistrationDate,
		reg.MachineID,
		reg.ProductID,
	)
	if err != nil {
		return fmt.Errorf("update registration: %w", err)
	}
	return nil
}

func (r *repo) Upsert(ctx context.Context, tx *sqlx.Tx, reg *Registration) error {
	_, err := tx.ExecContext(ctx, upsertRegistrationSQL,
		reg.MachineID,
		reg.ProductID,
		reg.ExpirationDate,
		reg.RegistrationHash,
		reg.FirstRegistrationDate,
		reg.LastRegistrationDate,
	)
	if err != nil {
		return fmt.Errorf("upsert registration: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, tx *sqlx.Tx, machineID, productID int64) error {
	_, err := tx.ExecContext(ctx, deleteRegistrationSQL, machineID, productID)
	if err != nil {
		return fmt.Errorf("delete registration: %w", err)
	}
	return nil
}

func (r *repo) UpdateInstalledVersion(ctx context.Context, machineID, productID int64, version string) error {
	result, err := r.db.ExecContext(ctx, updateInstalledVersionSQL, version, machineID, productID)
	if err != nil {
		return fmt.Errorf("update installed version: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("registration not found")
	}
	return nil
}
