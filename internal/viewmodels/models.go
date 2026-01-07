package viewmodels

import (
	"strings"
	"time"
)

// FeatureType represents the type of a feature
type FeatureType int

const (
	FeatureTypeInteger FeatureType = 0
	FeatureTypeString  FeatureType = 1
	FeatureTypeValues  FeatureType = 2
)

func (ft FeatureType) String() string {
	switch ft {
	case FeatureTypeInteger:
		return "Integer"
	case FeatureTypeString:
		return "String"
	case FeatureTypeValues:
		return "Values"
	default:
		return "Unknown"
	}
}

// Customer is a view model for customer display
type Customer struct {
	CustomerID   int64
	CustomerName string
	ContactName  string
	Phone        string
	Email        string
	Notes        string
}

// Product is a view model for product display
type Product struct {
	ProductID     int64
	ProductName   string
	ProductGUID   string
	LatestVersion string
	DownloadURL   string
}

// License is a view model for license display
type License struct {
	CustomerID          int64
	ProductID           int64
	ProductName         string
	LicenseKey          string
	LicenseCount        int
	IsSubscription      bool
	LicenseTerm         int
	StartDate           string
	ExpirationDate      string
	MaintExpirationDate string
	MaxProductVersion   string
}

// SubscriptionText returns "Yes" or "No" for subscription status
func (lic License) SubscriptionText() string {
	if lic.IsSubscription {
		return "Yes"
	}
	return "No"
}

// IsExpired checks if the license has expired
func (lic License) IsExpired() bool {
	if lic.ExpirationDate == "" {
		return false
	}
	t, err := time.Parse("2006-01-02", lic.ExpirationDate)
	if err != nil {
		return false
	}
	return t.Before(time.Now())
}

// IsMaintExpired checks if the maintenance has expired
func (lic License) IsMaintExpired() bool {
	if lic.MaintExpirationDate == "" {
		return false
	}
	t, err := time.Parse("2006-01-02", lic.MaintExpirationDate)
	if err != nil {
		return false
	}
	return t.Before(time.Now())
}

// Feature is a view model for feature definition display
type Feature struct {
	FeatureID     int64
	ProductID     int64
	FeatureName   string
	FeatureType   FeatureType
	AllowedValues string
	DefaultValue  string
}

// ProductFeature is a view model for customer-specific feature values
type ProductFeature struct {
	CustomerID    int64
	ProductID     int64
	FeatureID     int64
	FeatureName   string
	FeatureType   FeatureType
	FeatureValue  string
	AllowedValues string
	DefaultValue  string
}

// AllowedValuesList returns the allowed values as a slice (pipe-delimited)
func (pf ProductFeature) AllowedValuesList() []string {
	if pf.AllowedValues == "" {
		return nil
	}
	return strings.Split(pf.AllowedValues, "|")
}

// MachineRegistration is a view model for machine registration display
type MachineRegistration struct {
	MachineID        int64
	CustomerID       int64
	ProductID        int64
	MachineCode      string
	UserName         string
	RegHash          string
	ExpDate          string
	FirstRegDate     string
	LastRegDate      string
	InstalledVersion string
}

// IsExpired checks if the machine registration has expired
func (mr MachineRegistration) IsExpired() bool {
	if mr.ExpDate == "" {
		return false
	}
	t, err := time.Parse("2006-01-02", mr.ExpDate)
	if err != nil {
		return false
	}
	return t.Before(time.Now())
}

// ExpiredLicense is a view model for expired license display
type ExpiredLicense struct {
	CustomerName        string
	ContactName         string
	Email               string
	ProductName         string
	ExpirationDate      string
	MaintExpirationDate string
}
