package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func (s *Server) setupRoutes() {
	h := s.Handler
	s.Echo.Use(middleware.CORS())
	s.Echo.Use(s.loggerMiddleware)

	userGroup := s.Echo.Group("/api/v1", RequireSession(s.MerchantStorage, s.SessionRepository))
	pubGroup := s.Echo.Group("/api/v1")

	userGroup.GET("/merchants/:id", h.HandleRetrieveMerchant)
	userGroup.POST("/keys", h.HandleCreateAPIKey)

	pubGroup.POST("/signup", h.HandleRegisterOwner)
	pubGroup.POST("/login", h.HandleLogin)
	pubGroup.POST("/auth/signin", h.HandleUniversalLogin)
	pubGroup.GET("/auth/session", h.HandleUniversalGetSession)
	userGroup.POST("/signup/employee", h.HandleRegisterEmployee)
	userGroup.POST("/signout", h.HandleLogout)

	userGroup.GET("/employees/:id", h.HandleRetrieveEmployee)
	userGroup.POST("/employees/search", h.HandleSearchEmployee)
	userGroup.POST("/employees", h.HandleCreateEmployee)
	userGroup.PUT("/employees/:id", h.HandleUpdateEmployee)
	userGroup.DELETE("/employees/:id", h.HandleDeleteEmployee)

	userGroup.GET("/locations/:id", h.HandleRetrieveLocation)
	userGroup.GET("/locations", h.HandleListLocations)
	userGroup.POST("/locations", h.HandleCreateLocation)
	userGroup.PUT("/locations/:id", h.HandleUpdateLocation)
	userGroup.DELETE("/locations/:id", h.HandleDeleteLocation)

	userGroup.GET("/customers/:id", h.HandleRetrieveCustomer)
	userGroup.GET("/customers", h.HandleListCustomers)
	userGroup.POST("/customers", h.HandleCreateCustomer)
	userGroup.PUT("/customers/:id", h.HandleUpdateCustomer)
	userGroup.DELETE("/customers/:id", h.HandleDeleteCustomer)

	userGroup.POST("/inventory/batch-change", h.HandleChangeInventory)
	userGroup.POST("/inventory/batch-retrieve-counts", h.HandleBatchRetrieveInventory)

	userGroup.GET("/categories/:id", h.HandleRetrieveCategory)
	userGroup.GET("/categories", h.HandleListCategories)
	userGroup.POST("/categories", h.HandleCreateCategory)
	userGroup.PUT("/categories/:id", h.HandleUpdateCategory)
	userGroup.DELETE("/categories/:id", h.HandleDeleteCategory)

	userGroup.GET("/items/:id", h.HandleRetrieveItem)
	userGroup.GET("/items", h.HandleListItems)
	userGroup.POST("/items", h.HandleCreateItem)
	userGroup.PUT("/items/:id", h.HandleUpdateItem)
	userGroup.DELETE("/items/:id", h.HandleDeleteItem)

	userGroup.GET("/item_variations/:id", h.HandleRetrieveItemVariation)
	userGroup.GET("/item_variations", h.HandleListItemVariations)
	userGroup.POST("/item_variations", h.HandleCreateItemVariation)
	userGroup.PUT("/item_variations/:id", h.HandleUpdateItemVariation)
	userGroup.DELETE("/item_variations/:id", h.HandleDeleteItemVariation)

	userGroup.GET("/taxes/:id", h.HandleRetrieveTax)
	userGroup.GET("/taxes", h.HandleListTaxes)
	userGroup.POST("/taxes", h.HandleCreateTax)
	userGroup.PUT("/taxes/:id", h.HandleUpdateTax)
	userGroup.DELETE("/taxes/:id", h.HandleDeleteTax)
	userGroup.POST("/taxes/batch", h.HandleBatchCreateTax)

	userGroup.GET("/discounts/:id", h.HandleRetrieveDiscount)
	userGroup.GET("/discounts", h.HandleListDiscounts)
	userGroup.POST("/discounts", h.HandleCreateDiscount)
	userGroup.PUT("/discounts/:id", h.HandleUpdateDiscount)
	userGroup.DELETE("/discounts/:id", h.HandleDeleteDiscount)

	userGroup.POST("/orders", h.HandleCreateOrder)
	userGroup.POST("/orders/search", h.HandleSearchOrders)
	userGroup.POST("/orders/:order_id/pay", h.HandlePayOrder)

	userGroup.POST("/payments", h.HandleCreatePayment)

	userGroup.POST("/reports/custom", h.HandleGenerateCustomReport)
}

func (s *Server) loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		s.Echo.Logger.Infoj(log.JSON{
			"path":   c.Path(),
			"method": c.Request().Method,
		})
		return next(c)
	}
}
