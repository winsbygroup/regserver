package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Customers

func (h *Handler) GetCustomers(c echo.Context) error {
	out, err := h.svc.GetCustomers(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) GetCustomer(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	out, err := h.svc.GetCustomer(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateCustomer(c echo.Context) error {
	var req CreateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	out, err := h.svc.CreateCustomer(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) UpdateCustomer(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req UpdateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err := h.svc.UpdateCustomer(c.Request().Context(), id, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteCustomer(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := h.svc.DeleteCustomer(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) CustomerExists(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	exists, err := h.svc.CustomerExists(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, map[string]bool{"exists": exists})
}

// Products

func (h *Handler) GetProducts(c echo.Context) error {
	out, err := h.svc.GetProducts(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) GetProduct(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	out, err := h.svc.GetProduct(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateProduct(c echo.Context) error {
	var req CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	out, err := h.svc.CreateProduct(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) UpdateProduct(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	err := h.svc.UpdateProduct(c.Request().Context(), id, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteProduct(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := h.svc.DeleteProduct(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Customer Products (Licenses)

func (h *Handler) GetLicenses(c echo.Context) error {
	custID, _ := strconv.ParseInt(c.Param("customerId"), 10, 64)
	out, err := h.svc.GetLicenses(c.Request().Context(), custID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) GetUnlicensedProducts(c echo.Context) error {
	custID, _ := strconv.ParseInt(c.Param("customerId"), 10, 64)
	out, err := h.svc.GetUnlicensedProducts(c.Request().Context(), custID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateLicense(c echo.Context) error {
	custID, _ := strconv.ParseInt(c.Param("customerId"), 10, 64)
	var req CreateLicenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	out, err := h.svc.CreateLicense(c.Request().Context(), custID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) UpdateLicense(c echo.Context) error {
	custID, _ := strconv.ParseInt(c.Param("customerId"), 10, 64)
	prodID, _ := strconv.ParseInt(c.Param("productId"), 10, 64)
	var req UpdateLicenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	err := h.svc.UpdateLicense(c.Request().Context(), custID, prodID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteLicense(c echo.Context) error {
	custID, _ := strconv.ParseInt(c.Param("customerId"), 10, 64)
	prodID, _ := strconv.ParseInt(c.Param("productId"), 10, 64)
	err := h.svc.DeleteLicense(c.Request().Context(), custID, prodID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Feature Definitions

func (h *Handler) GetFeatures(c echo.Context) error {
	prodID, _ := strconv.ParseInt(c.Param("productId"), 10, 64)
	out, err := h.svc.GetFeatures(c.Request().Context(), prodID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateFeature(c echo.Context) error {
	prodID, _ := strconv.ParseInt(c.Param("productId"), 10, 64)
	var req CreateFeatureRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	out, err := h.svc.CreateFeature(c.Request().Context(), prodID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) UpdateFeature(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req UpdateFeatureRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	err := h.svc.UpdateFeature(c.Request().Context(), id, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteFeature(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := h.svc.DeleteFeature(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Product Feature Values

func (h *Handler) GetProductFeatures(c echo.Context) error {
	custID, _ := strconv.ParseInt(c.Param("customerId"), 10, 64)
	prodID, _ := strconv.ParseInt(c.Param("productId"), 10, 64)
	out, err := h.svc.GetProductFeatures(c.Request().Context(), custID, prodID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) UpdateProductFeature(c echo.Context) error {
	custID, _ := strconv.ParseInt(c.Param("customerId"), 10, 64)
	prodID, _ := strconv.ParseInt(c.Param("productId"), 10, 64)
	featID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req UpdateProductFeatureRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err := h.svc.UpdateProductFeature(c.Request().Context(), custID, prodID, featID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Machine Registrations

func (h *Handler) GetMachineRegistrations(c echo.Context) error {
	custID, _ := strconv.ParseInt(c.Param("customerId"), 10, 64)
	prodID, _ := strconv.ParseInt(c.Param("productId"), 10, 64)

	active := c.QueryParam("active") == "true"

	out, err := h.svc.GetMachineRegistrations(c.Request().Context(), custID, prodID, active)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) DeleteMachineRegistration(c echo.Context) error {
	machineID, _ := strconv.ParseInt(c.Param("machineId"), 10, 64)
	prodID, _ := strconv.ParseInt(c.Param("productId"), 10, 64)

	err := h.svc.DeleteMachineRegistration(c.Request().Context(), machineID, prodID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Expirations

func (h *Handler) GetExpirations(c echo.Context) error {
	before := c.QueryParam("before")
	if before == "" {
		before = time.Now().Format("2006-01-02")
	}

	out, err := h.svc.GetExpiredLicenses(c.Request().Context(), before)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}
