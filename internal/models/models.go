package models

import "time"

// Customer represents a customer in the system
type Customer struct {
	CustomerID   int64  `db:"customer_id"`
	CustomerName string `db:"customer_name"`
	ContactName  string `db:"contact_name"`
	Phone        string `db:"phone"`
	Email        string `db:"email"`
	Notes        string `db:"notes"`
}

// Machine represents a machine belonging to a customer
type Machine struct {
	MachineID   int64  `db:"machine_id"`
	CustomerID  int64  `db:"customer_id"`
	MachineCode string `db:"machine_code"`
	UserName    string `db:"user_name"`

	// Related entity
	Customer *Customer `db:"-"`
}

// Product represents a software product
type Product struct {
	ProductID     int64  `db:"product_id"`
	ProductName   string `db:"product_name"`
	ProductGUID   string `db:"product_guid"`
	LatestVersion string `db:"latest_version"`
	DownloadURL   string `db:"download_url"`
}

// Registration represents a product registered to a machine
type Registration struct {
	MachineID             int64  `db:"machine_id"`
	ProductID             int64  `db:"product_id"`
	ExpirationDate        string `db:"expiration_date"`
	RegistrationHash      string `db:"registration_hash"`
	FirstRegistrationDate string `db:"first_registration_date"`
	LastRegistrationDate  string `db:"last_registration_date"`

	// Related entities
	Machine *Machine `db:"-"`
	Product *Product `db:"-"`
}

// License represents a product licensed to a customer
type License struct {
	CustomerID          int64  `db:"customer_id"`
	ProductID           int64  `db:"product_id"`
	LicenseKey          string `db:"license_key"`
	LicenseCount        int    `db:"license_count"`
	IsSubscription      bool   `db:"is_subscription"`
	LicenseTerm         int    `db:"license_term"`
	StartDate           string `db:"start_date"`
	ExpirationDate      string `db:"expiration_date"`
	MaintExpirationDate string `db:"maint_expiration_date"`
	MaxProductVersion   string `db:"max_product_version"`

	// Related entities
	Customer *Customer         `db:"-"`
	Product  *Product          `db:"-"`
	Features []*LicenseFeature `db:"-"`
}

// Feature represents a feature of a product
type Feature struct {
	FeatureID     int64  `db:"feature_id"`
	ProductID     int64  `db:"product_id"`
	FeatureName   string `db:"feature_name"`
	FeatureType   int    `db:"feature_type"` // 0, 1, or 2
	AllowedValues string `db:"allowed_values"`
	DefaultValue  string `db:"default_value"`

	// Related entity
	Product *Product `db:"-"`
}

// LicenseFeature represents a feature configuration for a customer's licensed product
type LicenseFeature struct {
	CustomerID   int64  `db:"customer_id"`
	ProductID    int64  `db:"product_id"`
	FeatureID    int64  `db:"feature_id"`
	FeatureValue string `db:"feature_value"`

	// Related entities
	License *License `db:"-"`
	Feature *Feature `db:"-"`
}

// Helper methods for date handling

// ParseDate converts a string date to a time.Time
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", dateStr)
}

// FormatDate converts a time.Time to a string date
func FormatDate(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format("2006-01-02")
}
