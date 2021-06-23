package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func (s *Server) setupRoutes() {
	s.Echo.Use(middleware.CORS())
	s.Echo.Use(s.loggerMiddleware)

	userGroup := s.Echo.Group("/api/v1", s.Handler.Authenticate)
	pubGroup := s.Echo.Group("/api/v1")

	userGroup.GET("/merchants/:id", s.Handler.RetrieveMerchant)
	userGroup.GET("/merchants", s.Handler.ListMerchants)
	userGroup.POST("/merchants", s.Handler.CreateMerchant)
	userGroup.PUT("/merchants/:id", s.Handler.UpdateMerchant)

	pubGroup.POST("/signup", s.Handler.Signup)
	pubGroup.POST("/login", s.Handler.Login)
	userGroup.POST("/signout", s.Handler.Signout)

	userGroup.GET("/locations/:id", s.Handler.RetrieveLocation)
	userGroup.GET("/locations", s.Handler.ListLocations)
	userGroup.POST("/locations", s.Handler.CreateLocation)
	userGroup.PUT("/locations/:id", s.Handler.UpdateLocation)
	userGroup.DELETE("/locations/:id", s.Handler.DeleteLocation)

	userGroup.GET("/customers/:id", s.Handler.RetrieveCustomer)
	userGroup.GET("/customers", s.Handler.ListCustomers)
	userGroup.POST("/customers", s.Handler.CreateCustomer)
	userGroup.PUT("/customers/:id", s.Handler.UpdateCustomer)
	userGroup.DELETE("/customers/:id", s.Handler.DeleteCustomer)

	userGroup.GET("/categories/:id", s.Handler.RetrieveCategory)
	userGroup.GET("/categories", s.Handler.ListCategories)
	userGroup.POST("/categories", s.Handler.CreateCategory)
	userGroup.PUT("/categories/:id", s.Handler.UpdateCategory)
	userGroup.DELETE("/categories/:id", s.Handler.DeleteCategory)

	userGroup.GET("/items/:id", s.Handler.RetrieveItem)
	userGroup.GET("/items", s.Handler.ListItems)
	userGroup.POST("/items", s.Handler.CreateItem)
	userGroup.PUT("/items/:id", s.Handler.UpdateItem)
	userGroup.DELETE("/items/:id", s.Handler.DeleteItem)

	userGroup.GET("/item_variations/:id", s.Handler.RetrieveItemVariation)
	userGroup.GET("/item_variations", s.Handler.ListItemVariations)
	userGroup.POST("/item_variations", s.Handler.CreateItemVariation)
	userGroup.PUT("/item_variations/:id", s.Handler.UpdateItemVariation)
	userGroup.DELETE("/item_variations/:id", s.Handler.DeleteItemVariation)

	userGroup.GET("/taxes/:id", s.Handler.RetrieveTax)
	userGroup.GET("/taxes", s.Handler.ListTaxes)
	userGroup.POST("/taxes", s.Handler.CreateTax)
	userGroup.PUT("/taxes/:id", s.Handler.UpdateTax)
	userGroup.DELETE("/taxes/:id", s.Handler.DeleteTax)
	userGroup.POST("/taxes/batch", s.Handler.BatchCreateTax)

	userGroup.GET("/discounts/:id", s.Handler.RetrieveDiscount)
	userGroup.GET("/discounts", s.Handler.ListDiscounts)
	userGroup.POST("/discounts", s.Handler.CreateDiscount)
	userGroup.PUT("/discounts/:id", s.Handler.UpdateDiscount)
	userGroup.DELETE("/discounts/:id", s.Handler.DeleteDiscount)
	userGroup.POST("/discounts/batch", s.Handler.BatchCreateDiscount)

	userGroup.POST("/orders", s.Handler.CreateOrder)
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
