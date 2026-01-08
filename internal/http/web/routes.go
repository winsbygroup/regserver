package web

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all web UI routes
func RegisterRoutes(e *echo.Group, h *Handler) {
	// Authentication
	e.GET("/login", h.LoginPage)
	e.POST("/login", h.Login)
	e.POST("/logout", h.Logout)

	// Dashboard
	e.GET("/", h.Index)
	e.GET("", h.Index)

	// Customers
	e.GET("/customers", h.ListCustomers)
	e.GET("/customers/new", h.NewCustomerForm)
	e.POST("/customers", h.CreateCustomer)
	e.GET("/customers/:id/edit", h.EditCustomerForm)
	e.PUT("/customers/:id", h.UpdateCustomer)
	e.DELETE("/customers/:id", h.DeleteCustomer)

	// Products
	e.GET("/products", h.ListProducts)
	e.GET("/products/new", h.NewProductForm)
	e.POST("/products", h.CreateProduct)
	e.GET("/products/:id/edit", h.EditProductForm)
	e.PUT("/products/:id", h.UpdateProduct)
	e.DELETE("/products/:id", h.DeleteProduct)

	// Product Features (definitions)
	e.GET("/products/:id/features", h.ProductFeaturesManager)
	e.GET("/products/:id/features/new", h.NewFeatureForm)
	e.POST("/products/:id/features", h.CreateFeature)
	e.GET("/products/:id/features/:featureId/edit", h.EditFeatureForm)
	e.PUT("/products/:id/features/:featureId", h.UpdateFeature)
	e.DELETE("/products/:id/features/:featureId", h.DeleteFeature)

	// Licenses
	e.GET("/licenses/:customerID", h.GetLicenses)
	e.GET("/licenses/:customerID/new", h.NewLicenseForm)
	e.POST("/licenses/:customerID", h.CreateLicense)
	e.GET("/licenses/:customerID/:productID/edit", h.EditLicenseForm)
	e.PUT("/licenses/:customerID/:productID", h.UpdateLicense)
	e.DELETE("/licenses/:customerID/:productID", h.DeleteLicense)

	// Product Features (customer values)
	e.GET("/features/:customerID/:productID", h.GetProductFeatures)
	e.GET("/features/:customerID/:productID/:featureID/edit", h.EditFeatureValueForm)
	e.PUT("/features/:customerID/:productID/:featureID", h.UpdateFeatureValue)

	// Machine Registrations
	e.GET("/machines/:customerID/:productID", h.GetMachineRegistrations)
	e.GET("/machines/:customerID/:productID/add", h.ManualRegistrationForm)
	e.POST("/machines/:customerID/:productID", h.CreateManualRegistration)
	e.GET("/machines/:machineID/:productID/export", h.ExportMachineRegistration)
	e.DELETE("/machines/:machineID/:productID", h.DeleteMachineRegistration)

	// Expirations
	e.GET("/expirations", h.ListExpirations)
	e.GET("/expirations/csv", h.ExportExpirationsCSV)
}
