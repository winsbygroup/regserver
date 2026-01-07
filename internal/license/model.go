package license

type License struct {
	CustomerID          int64  `db:"customer_id"`
	ProductID           int64  `db:"product_id"`
	LicenseKey          string `db:"license_key"`
	LicenseCount        int    `db:"license_count"`
	IsSubscription      int    `db:"is_subscription"`
	LicenseTerm         int    `db:"license_term"`
	StartDate           string `db:"start_date"`
	ExpirationDate      string `db:"expiration_date"`
	MaintExpirationDate string `db:"maint_expiration_date"`
	MaxProductVersion   string `db:"max_product_version"`
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
