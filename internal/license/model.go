package license

import (
	"errors"

	"winsbygroup.com/regserver/internal/product"
)

// Validation errors
var (
	ErrSubscriptionRequiresTerm = errors.New("subscription licenses require a term greater than 0")
	ErrInvalidMaxVersion        = errors.New("max product version must be empty or in #.#.# format")
)

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
}

// Validate checks business rules for a license
func (l *License) Validate() error {
	if l.IsSubscription && l.LicenseTerm <= 0 {
		return ErrSubscriptionRequiresTerm
	}
	if !product.IsValidVersion(l.MaxProductVersion) {
		return ErrInvalidMaxVersion
	}
	return nil
}

// ExpiredLicense represents an expired license with customer/product details
type ExpiredLicense struct {
	CustomerName        string `db:"customer_name" json:"customerName"`
	ContactName         string `db:"contact_name" json:"contactName"`
	Email               string `db:"email" json:"email"`
	ProductName         string `db:"product_name" json:"productName"`
	ExpirationDate      string `db:"expiration_date" json:"expirationDate"`
	MaintExpirationDate string `db:"maint_expiration_date" json:"maintExpirationDate"`
}
