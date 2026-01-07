package product

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jmoiron/sqlx"
)

// versionRegex validates version format: #.#.# (e.g., "1.0.0", "2.3.4")
var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// IsValidVersion returns true if version is empty or matches #.#.# format
func IsValidVersion(version string) bool {
	return version == "" || versionRegex.MatchString(version)
}

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

func (s *Service) GetAll(ctx context.Context) ([]Product, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) Get(ctx context.Context, id int64) (*Product, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) GetByGUID(ctx context.Context, guid string) (*Product, error) {
	return s.repo.GetByGUID(ctx, guid)
}

// validate checks product fields for validity
func (s *Service) validate(p *Product) error {
	if p.LatestVersion != "" && !versionRegex.MatchString(p.LatestVersion) {
		return fmt.Errorf("latest version must be in #.#.# format (e.g., 1.0.0)")
	}
	return nil
}

func (s *Service) Create(ctx context.Context, p *Product) (*Product, error) {
	if err := s.validate(p); err != nil {
		return nil, err
	}

	var id int64
	err := s.WithTx(ctx, func(tx *sqlx.Tx) error {
		var err error
		id, err = s.repo.Create(ctx, tx, p)
		return err
	})
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, p *Product) error {
	if err := s.validate(p); err != nil {
		return err
	}

	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Update(ctx, tx, p)
	})
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.repo.Delete(ctx, tx, id)
	})
}
