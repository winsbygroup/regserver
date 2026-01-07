package admin

import (
	"context"

	"github.com/google/uuid"

	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/registration"
)

type Service struct {
	customers     *customer.Service
	products      *product.Service
	licenses      *license.Service
	features      *feature.Service
	featureValues *featurevalue.Service
	machines      *machine.Service
	registrations *registration.Service
}

func NewService(
	c *customer.Service,
	p *product.Service,
	lic *license.Service,
	f *feature.Service,
	fv *featurevalue.Service,
	m *machine.Service,
	r *registration.Service,
) *Service {
	return &Service{
		customers:     c,
		products:      p,
		licenses:      lic,
		features:      f,
		featureValues: fv,
		machines:      m,
		registrations: r,
	}
}

// -------------------------
// Customers
// -------------------------

func (s *Service) GetCustomers(ctx context.Context) ([]customer.Customer, error) {
	return s.customers.GetAll(ctx)
}

func (s *Service) GetCustomer(ctx context.Context, id int64) (*customer.Customer, error) {
	return s.customers.Get(ctx, id)
}

func (s *Service) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*customer.Customer, error) {
	c := &customer.Customer{
		CustomerName: req.CustomerName,
		ContactName:  req.ContactName,
		Phone:        req.Phone,
		Email:        req.Email,
		Notes:        req.Notes,
	}
	return s.customers.Create(ctx, c)
}

func (s *Service) UpdateCustomer(ctx context.Context, id int64, req *UpdateCustomerRequest) error {
	c := &customer.Customer{
		CustomerID:   id,
		CustomerName: req.CustomerName,
		ContactName:  req.ContactName,
		Phone:        req.Phone,
		Email:        req.Email,
		Notes:        req.Notes,
	}
	return s.customers.Update(ctx, c)
}

func (s *Service) DeleteCustomer(ctx context.Context, id int64) error {
	return s.customers.Delete(ctx, id)
}

func (s *Service) CustomerExists(ctx context.Context, id int64) (bool, error) {
	return s.customers.Exists(ctx, id)
}

// -------------------------
// Products
// -------------------------

func (s *Service) GetProducts(ctx context.Context) ([]product.Product, error) {
	return s.products.GetAll(ctx)
}

func (s *Service) GetProduct(ctx context.Context, id int64) (*product.Product, error) {
	return s.products.Get(ctx, id)
}

func (s *Service) CreateProduct(ctx context.Context, req *CreateProductRequest) (*product.Product, error) {
	p := &product.Product{
		ProductName:   req.Name,
		ProductGUID:   req.Guid,
		LatestVersion: req.LatestVersion,
		DownloadURL:   req.DownloadURL,
	}
	return s.products.Create(ctx, p)
}

func (s *Service) UpdateProduct(ctx context.Context, id int64, req *UpdateProductRequest) error {
	p := &product.Product{
		ProductID:     id,
		ProductName:   req.Name,
		ProductGUID:   req.Guid,
		LatestVersion: req.LatestVersion,
		DownloadURL:   req.DownloadURL,
	}
	return s.products.Update(ctx, p)
}

func (s *Service) DeleteProduct(ctx context.Context, id int64) error {
	return s.products.Delete(ctx, id)
}

// -------------------------
// Licenses
// -------------------------

func (s *Service) GetLicenses(ctx context.Context, customerID int64) ([]license.License, error) {
	return s.licenses.GetForCustomer(ctx, customerID)
}

func (s *Service) GetUnlicensedProducts(ctx context.Context, customerID int64) ([]product.Product, error) {
	return s.licenses.GetUnlicensed(ctx, customerID)
}

func (s *Service) CreateLicense(ctx context.Context, customerID int64, req *CreateLicenseRequest) (*license.License, error) {
	lic := &license.License{
		CustomerID:          customerID,
		ProductID:           req.ProductID,
		LicenseKey:          uuid.New().String(),
		LicenseCount:        req.LicenseCount,
		IsSubscription:      req.IsSubscription,
		LicenseTerm:         req.LicenseTerm,
		StartDate:           req.StartDate,
		ExpirationDate:      req.ExpirationDate,
		MaintExpirationDate: req.MaintExpirationDate,
		MaxProductVersion:   req.MaxProductVersion,
	}
	return s.licenses.Create(ctx, lic)
}

func (s *Service) UpdateLicense(ctx context.Context, customerID, productID int64, req *UpdateLicenseRequest) error {
	lic := &license.License{
		CustomerID:          customerID,
		ProductID:           productID,
		LicenseCount:        req.LicenseCount,
		IsSubscription:      req.IsSubscription,
		LicenseTerm:         req.LicenseTerm,
		StartDate:           req.StartDate,
		ExpirationDate:      req.ExpirationDate,
		MaintExpirationDate: req.MaintExpirationDate,
		MaxProductVersion:   req.MaxProductVersion,
	}
	return s.licenses.Update(ctx, lic)
}

func (s *Service) DeleteLicense(ctx context.Context, customerID, productID int64) error {
	return s.licenses.Delete(ctx, customerID, productID)
}

// -------------------------
// Feature Definitions (per product)
// -------------------------

func (s *Service) GetFeatures(ctx context.Context, productID int64) ([]feature.Feature, error) {
	return s.features.GetForProduct(ctx, productID)
}

func (s *Service) CreateFeature(ctx context.Context, productID int64, req *CreateFeatureRequest) (*feature.Feature, error) {
	f := &feature.Feature{
		ProductID:     productID,
		FeatureName:   req.FeatureName,
		FeatureType:   feature.ToInt(req.FeatureType),
		AllowedValues: req.AllowedValues,
		DefaultValue:  req.DefaultValue,
	}
	return s.features.Create(ctx, f)
}

func (s *Service) UpdateFeature(ctx context.Context, featureID int64, req *UpdateFeatureRequest) error {
	f := &feature.Feature{
		FeatureID:     featureID,
		FeatureName:   req.FeatureName,
		FeatureType:   feature.ToInt(req.FeatureType),
		AllowedValues: req.AllowedValues,
		DefaultValue:  req.DefaultValue,
	}
	return s.features.Update(ctx, f)
}

func (s *Service) DeleteFeature(ctx context.Context, featureID int64) error {
	return s.features.Delete(ctx, featureID)
}

// -------------------------
// Product Feature Values (customer-specific overrides)
// -------------------------

func (s *Service) GetProductFeatures(ctx context.Context, customerID, productID int64) ([]featurevalue.FeatureValue, error) {
	return s.featureValues.GetFeatureValues(ctx, customerID, productID)
}

func (s *Service) UpdateProductFeature(ctx context.Context, customerID, productID, featureID int64, req *UpdateProductFeatureRequest) error {
	fv := &featurevalue.FeatureValue{
		CustomerID:   customerID,
		ProductID:    productID,
		FeatureID:    featureID,
		FeatureValue: req.Value,
	}
	return s.featureValues.Update(ctx, fv)
}

// -------------------------
// Machine Registrations
// -------------------------

func (s *Service) GetMachineRegistrations(ctx context.Context, customerID, productID int64, activeOnly bool) ([]machine.Machine, error) {
	if activeOnly {
		return s.machines.GetActiveForLicense(ctx, customerID, productID)
	}
	return s.machines.GetForLicense(ctx, customerID, productID)
}

func (s *Service) DeleteMachineRegistration(ctx context.Context, machineID, productID int64) error {
	return s.registrations.Delete(ctx, machineID, productID)
}

// -------------------------
// Expirations
// -------------------------

func (s *Service) GetExpiredLicenses(ctx context.Context, before string) ([]license.ExpiredLicense, error) {
	return s.licenses.GetExpiredLicenses(ctx, before)
}
