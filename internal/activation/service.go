package activation

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/registration"
)

type Service struct {
	db                 *sqlx.DB
	registrationSecret string
	customerSvc        *customer.Service
	machineSvc         *machine.Service
	regSvc             *registration.Service
	licenseSvc         *license.Service
	productSvc         *product.Service
	featureSvc         *feature.Service
	featureValueSvc    *featurevalue.Service
}

func NewService(
	db *sqlx.DB,
	registrationSecret string,
	customerSvc *customer.Service,
	machineSvc *machine.Service,
	regSvc *registration.Service,
	licenseSvc *license.Service,
	productSvc *product.Service,
	featureSvc *feature.Service,
	featureValueSvc *featurevalue.Service,
) *Service {
	return &Service{
		db:                 db,
		registrationSecret: registrationSecret,
		customerSvc:        customerSvc,
		machineSvc:         machineSvc,
		regSvc:             regSvc,
		licenseSvc:         licenseSvc,
		productSvc:         productSvc,
		featureSvc:         featureSvc,
		featureValueSvc:    featureValueSvc,
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

func (s *Service) Activate(
	ctx context.Context,
	customerID, productID int64,
	req *Request,
) (*Response, error) {

	now := time.Now().Format("2006-01-02")
	var machineID int64

	// License - check early for license validity
	lic, err := s.licenseSvc.Get(ctx, customerID, productID)
	if err != nil {
		return nil, err
	}
	if lic == nil {
		return nil, fmt.Errorf("no license for customer %d product %d", customerID, productID)
	}

	// License count check - get active machines and verify we haven't exceeded the limit
	activeMachines, err := s.machineSvc.GetActiveForLicense(ctx, customerID, productID)
	if err != nil {
		return nil, err
	}

	// Check if this machine is already active (allow re-activation)
	isExistingMachine := false
	for _, m := range activeMachines {
		if m.MachineCode == req.MachineCode {
			isExistingMachine = true
			break
		}
	}

	// If this is a new machine and we're at the license limit, reject
	if !isExistingMachine && len(activeMachines) >= lic.LicenseCount {
		return nil, fmt.Errorf("license count exceeded: %d of %d licenses in use", len(activeMachines), lic.LicenseCount)
	}

	// Fetch customer (for CustomerName)
	cust, err := s.customerSvc.Get(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// Product info
	prod, err := s.productSvc.Get(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Features - fetch before transaction so we can compute the hash
	defs, err := s.featureSvc.GetForProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	vals, err := s.featureValueSvc.GetFeatureValues(ctx, customerID, productID)
	if err != nil {
		return nil, err
	}

	merged := feature.MergeWithOverrides(defs, vals)

	// Compute registration hash (includes MaxProductVersion for tamper detection)
	regHash, err := s.computeHash(req.MachineCode, lic.ExpirationDate, lic.MaintExpirationDate, lic.MaxProductVersion, merged)
	if err != nil {
		return nil, fmt.Errorf("compute registration hash: %w", err)
	}

	// Machine + registration writes in a single transaction
	err = s.WithTx(ctx, func(tx *sqlx.Tx) error {

		// Machine
		mid, err := s.machineSvc.GetOrCreate(ctx, tx, customerID, req.MachineCode, req.UserName)
		if err != nil {
			return err
		}
		machineID = mid

		// Registration upsert
		reg := &registration.Registration{
			MachineID:             machineID,
			ProductID:             productID,
			ExpirationDate:        lic.ExpirationDate,
			RegistrationHash:      regHash,
			FirstRegistrationDate: now,
			LastRegistrationDate:  now,
		}

		return s.regSvc.Upsert(ctx, tx, reg)
	})

	if err != nil {
		return nil, err
	}

	// Build response
	return &Response{
		UserName:            req.UserName,
		UserCompany:         cust.CustomerName,
		MachineCode:         req.MachineCode,
		ExpirationDate:      lic.ExpirationDate,
		MaintExpirationDate: lic.MaintExpirationDate,
		MaxProductVersion:   lic.MaxProductVersion,
		LatestVersion:       prod.LatestVersion,
		ProductGUID:         prod.ProductGUID,
		LicenseKey:          lic.LicenseKey,
		RegistrationHash:    regHash,
		Features:            merged,
	}, nil
}

// computeHash builds the registration string and computes its HMAC hash
func (s *Service) computeHash(machineCode, expDate, maintExpDate, maxVersion string, features map[string]any) (string, error) {
	// Convert features to string map
	featStr := make(map[string]string, len(features))
	for k, v := range features {
		featStr[k] = fmt.Sprintf("%v", v)
	}

	// Build registration string and compute HMAC hash
	regStr := buildRegistrationString(machineCode, expDate, maintExpDate, maxVersion, featStr)
	return computeRegistrationHash(regStr, s.registrationSecret)
}

