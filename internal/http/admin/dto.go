package admin

// -------------------------
// Customer DTOs
// -------------------------

type CreateCustomerRequest struct {
	CustomerName string `json:"customerName"`
	ContactName  string `json:"contactName"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Notes        string `json:"notes"`
}

type UpdateCustomerRequest struct {
	CustomerName string `json:"customerName"`
	ContactName  string `json:"contactName"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Notes        string `json:"notes"`
}

// -------------------------
// Product DTOs
// -------------------------

type CreateProductRequest struct {
	Name          string `json:"name"`
	Guid          string `json:"guid"`
	LatestVersion string `json:"latestVersion"`
	DownloadURL   string `json:"downloadUrl"`
}

type UpdateProductRequest struct {
	Name          string `json:"name"`
	Guid          string `json:"guid"`
	LatestVersion string `json:"latestVersion"`
	DownloadURL   string `json:"downloadUrl"`
}

// -------------------------
// License DTOs
// -------------------------

type CreateLicenseRequest struct {
	CustomerID          int64  `json:"customerId"`
	ProductID           int64  `json:"productId"`
	LicenseCount        int    `json:"licenseCount"`
	IsSubscription      int    `json:"isSubscription"`
	LicenseTerm         int    `json:"licenseTerm"`
	StartDate           string `json:"startDate"`
	ExpirationDate      string `json:"expirationDate"`
	MaintExpirationDate string `json:"maintExpirationDate"`
	MaxProductVersion   string `json:"maxProductVersion"`
}

type UpdateLicenseRequest struct {
	LicenseCount        int    `json:"licenseCount"`
	IsSubscription      int    `json:"isSubscription"`
	LicenseTerm         int    `json:"licenseTerm"`
	StartDate           string `json:"startDate"`
	ExpirationDate      string `json:"expirationDate"`
	MaintExpirationDate string `json:"maintExpirationDate"`
	MaxProductVersion   string `json:"maxProductVersion"`
}

// -------------------------
// Feature Definition DTOs
// -------------------------

type CreateFeatureRequest struct {
	FeatureName   string `json:"featureName"`
	FeatureType   string `json:"featureType"`
	AllowedValues string `json:"allowedValues"`
	DefaultValue  string `json:"defaultValue"`
}

type UpdateFeatureRequest struct {
	FeatureName   string `json:"featureName"`
	FeatureType   string `json:"featureType"`
	AllowedValues string `json:"allowedValues"`
	DefaultValue  string `json:"defaultValue"`
}

// -------------------------
// Feature Value DTOs (customer-specific)
// -------------------------

type UpdateProductFeatureRequest struct {
	Value string `json:"value"`
}
