package machine

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

func (s *Service) Get(ctx context.Context, machineID int64) (*Machine, error) {
	return s.repo.GetByID(ctx, machineID)
}

func (s *Service) GetByCode(ctx context.Context, customerID int64, machineCode string) (*Machine, error) {
	return s.repo.GetByCode(ctx, customerID, machineCode)
}

func (s *Service) GetOrCreate(
	ctx context.Context,
	tx *sqlx.Tx,
	customerID int64,
	machineCode string,
	userName string,
) (int64, error) {

	// Lookup
	m, err := s.repo.GetByCode(ctx, customerID, machineCode)
	if err != nil {
		return 0, err
	}

	// Create if missing
	if m == nil {
		newMachine := &Machine{
			CustomerID:  customerID,
			MachineCode: machineCode,
			UserName:    userName,
		}
		return s.repo.Create(ctx, tx, newMachine)
	}

	// Update username if changed
	if m.UserName != userName {
		if err := s.repo.UpdateUserName(ctx, tx, m.MachineID, userName); err != nil {
			return 0, err
		}
	}

	return m.MachineID, nil
}

func (s *Service) GetForLicense(ctx context.Context, customerID, productID int64) ([]Machine, error) {
	return s.repo.GetForLicense(ctx, customerID, productID)
}

func (s *Service) GetActiveForLicense(ctx context.Context, customerID, productID int64) ([]Machine, error) {
	return s.repo.GetActiveForLicense(ctx, customerID, productID)
}
