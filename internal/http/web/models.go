package web

import (
	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/featurevalue"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/product"
	vm "winsbygroup.com/regserver/internal/viewmodels"
)

// Re-export types for convenience
type (
	Customer            = vm.Customer
	Product             = vm.Product
	License             = vm.License
	Feature             = vm.Feature
	ProductFeature      = vm.ProductFeature
	MachineRegistration = vm.MachineRegistration
	ExpiredLicense      = vm.ExpiredLicense
	FeatureType         = vm.FeatureType
)

// Re-export constants
const (
	FeatureTypeInteger = vm.FeatureTypeInteger
	FeatureTypeString  = vm.FeatureTypeString
	FeatureTypeValues  = vm.FeatureTypeValues
)

// FromDomainCustomer converts a domain customer to view model
func FromDomainCustomer(c customer.Customer) vm.Customer {
	return vm.Customer{
		CustomerID:   c.CustomerID,
		CustomerName: c.CustomerName,
		ContactName:  c.ContactName,
		Phone:        c.Phone,
		Email:        c.Email,
		Notes:        c.Notes,
	}
}

// FromDomainCustomers converts a slice of domain customers to view models
func FromDomainCustomers(customers []customer.Customer) []vm.Customer {
	result := make([]vm.Customer, len(customers))
	for i, c := range customers {
		result[i] = FromDomainCustomer(c)
	}
	return result
}

// FromDomainProduct converts a domain product to view model
func FromDomainProduct(p product.Product) vm.Product {
	return vm.Product{
		ProductID:     p.ProductID,
		ProductName:   p.ProductName,
		ProductGUID:   p.ProductGUID,
		LatestVersion: p.LatestVersion,
		DownloadURL:   p.DownloadURL,
	}
}

// FromDomainProducts converts a slice of domain products to view models
func FromDomainProducts(products []product.Product) []vm.Product {
	result := make([]vm.Product, len(products))
	for i, p := range products {
		result[i] = FromDomainProduct(p)
	}
	return result
}

// FromDomainLicense converts a domain license to view model
func FromDomainLicense(lic license.License, productName string) vm.License {
	return vm.License{
		CustomerID:          lic.CustomerID,
		ProductID:           lic.ProductID,
		ProductName:         productName,
		LicenseKey:          lic.LicenseKey,
		LicenseCount:        lic.LicenseCount,
		IsSubscription:      lic.IsSubscription,
		LicenseTerm:         lic.LicenseTerm,
		StartDate:           lic.StartDate,
		ExpirationDate:      lic.ExpirationDate,
		MaintExpirationDate: lic.MaintExpirationDate,
		MaxProductVersion:   lic.MaxProductVersion,
	}
}

// FromDomainFeature converts a domain feature to view model
func FromDomainFeature(f feature.Feature) vm.Feature {
	return vm.Feature{
		FeatureID:     f.FeatureID,
		ProductID:     f.ProductID,
		FeatureName:   f.FeatureName,
		FeatureType:   vm.FeatureType(f.FeatureType),
		AllowedValues: f.AllowedValues,
		DefaultValue:  f.DefaultValue,
	}
}

// FromDomainFeatures converts a slice of domain features to view models
func FromDomainFeatures(features []feature.Feature) []vm.Feature {
	result := make([]vm.Feature, len(features))
	for i, f := range features {
		result[i] = FromDomainFeature(f)
	}
	return result
}

// FromDomainFeatureValue converts domain feature value + feature to view model
func FromDomainFeatureValue(fv featurevalue.FeatureValue, f feature.Feature) vm.ProductFeature {
	return vm.ProductFeature{
		CustomerID:    fv.CustomerID,
		ProductID:     fv.ProductID,
		FeatureID:     fv.FeatureID,
		FeatureName:   f.FeatureName,
		FeatureType:   vm.FeatureType(f.FeatureType),
		FeatureValue:  fv.FeatureValue,
		AllowedValues: f.AllowedValues,
		DefaultValue:  f.DefaultValue,
	}
}

// FromDomainMachine converts a domain machine to view model
func FromDomainMachine(m machine.Machine, productID int64, regHash, expDate, firstRegDate, lastRegDate, installedVersion string) vm.MachineRegistration {
	return vm.MachineRegistration{
		MachineID:        m.MachineID,
		CustomerID:       m.CustomerID,
		ProductID:        productID,
		MachineCode:      m.MachineCode,
		UserName:         m.UserName,
		RegHash:          regHash,
		ExpDate:          expDate,
		FirstRegDate:     firstRegDate,
		LastRegDate:      lastRegDate,
		InstalledVersion: installedVersion,
	}
}

// FromDomainExpiredLicense converts a domain expired license to view model
func FromDomainExpiredLicense(el license.ExpiredLicense) vm.ExpiredLicense {
	return vm.ExpiredLicense{
		CustomerName:        el.CustomerName,
		ContactName:         el.ContactName,
		Email:               el.Email,
		ProductName:         el.ProductName,
		ExpirationDate:      el.ExpirationDate,
		MaintExpirationDate: el.MaintExpirationDate,
	}
}

// FromDomainExpiredLicenses converts a slice of domain expired licenses to view models
func FromDomainExpiredLicenses(licenses []license.ExpiredLicense) []vm.ExpiredLicense {
	result := make([]vm.ExpiredLicense, len(licenses))
	for i, l := range licenses {
		result[i] = FromDomainExpiredLicense(l)
	}
	return result
}
