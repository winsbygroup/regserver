package admin

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, h *Handler) {

	// Customers
	g.GET("/customers", h.GetCustomers)
	g.GET("/customers/:id", h.GetCustomer)
	g.POST("/customers", h.CreateCustomer)
	g.PUT("/customers/:id", h.UpdateCustomer)
	g.DELETE("/customers/:id", h.DeleteCustomer)
	g.GET("/customers/:id/exists", h.CustomerExists)

	// Products
	g.GET("/products", h.GetProducts)
	g.GET("/products/:id", h.GetProduct)
	g.POST("/products", h.CreateProduct)
	g.PUT("/products/:id", h.UpdateProduct)
	g.DELETE("/products/:id", h.DeleteProduct)

	// Licenses (customer products)
	g.GET("/customers/:customerId/products", h.GetLicenses)
	g.GET("/customers/:customerId/unlicensed-products", h.GetUnlicensedProducts)
	g.POST("/customers/:customerId/products", h.CreateLicense)
	g.PUT("/customers/:customerId/products/:productId", h.UpdateLicense)
	g.DELETE("/customers/:customerId/products/:productId", h.DeleteLicense)

	// Feature definitions (per product)
	g.GET("/products/:productId/features", h.GetFeatures)
	g.POST("/products/:productId/features", h.CreateFeature)
	g.PUT("/features/:id", h.UpdateFeature)
	g.DELETE("/features/:id", h.DeleteFeature)

	// Product features (customer-specific feature values)
	g.GET("/customers/:customerId/products/:productId/features", h.GetProductFeatures)
	g.PUT("/customers/:customerId/products/:productId/features/:id", h.UpdateProductFeature)

	// Machine registrations
	g.GET("/customers/:customerId/products/:productId/registrations", h.GetMachineRegistrations)
	g.DELETE("/registrations/:machineId/:productId", h.DeleteMachineRegistration)

	// Expirations
	g.GET("/expirations", h.GetExpirations)

	// Backup
	g.POST("/backup", h.BackupDatabase)
}
