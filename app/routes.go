package app

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func (s *Server) setupRoutes() {
	s.Echo.Use(middleware.CORS())
	s.Echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			s.Echo.Logger.Infoj(log.JSON{
				"path":   c.Path(),
				"method": c.Request().Method,
			})
			return next(c)
		}
	})
	s.Echo.GET("/api/v1/merchants/:id", s.merchantHandler.Retrieve, s.authHandler.Authenticate)
	s.Echo.GET("/api/v1/merchants", s.merchantHandler.ListAll, s.authHandler.Authenticate)
	s.Echo.POST("/api/v1/merchants", s.merchantHandler.Create, s.authHandler.Authenticate)
	s.Echo.PUT("/api/v1/merchants/:id", s.merchantHandler.Update, s.authHandler.Authenticate)

	s.Echo.POST("/api/v1/signup", s.authHandler.Signup)
	s.Echo.POST("/api/v1/login", s.authHandler.Login)
	s.Echo.POST("/api/v1/signout", s.authHandler.Signout, s.authHandler.Authenticate)

	s.Echo.GET("/api/v1/locations/:id", s.locationHandler.Retrieve, s.authHandler.Authenticate)
	s.Echo.GET("/api/v1/locations", s.locationHandler.ListAll, s.authHandler.Authenticate)
	s.Echo.POST("/api/v1/locations", s.locationHandler.Create, s.authHandler.Authenticate)
	s.Echo.PUT("/api/v1/locations/:id", s.locationHandler.Update, s.authHandler.Authenticate)
	s.Echo.DELETE("/api/v1/locations/:id", s.locationHandler.Delete, s.authHandler.Authenticate)

	s.Echo.GET("/api/v1/customers/:id", s.customerHandler.Retrieve, s.authHandler.Authenticate)
	s.Echo.GET("/api/v1/customers", s.customerHandler.ListAll, s.authHandler.Authenticate)
	s.Echo.POST("/api/v1/customers", s.customerHandler.Create, s.authHandler.Authenticate)
	s.Echo.PUT("/api/v1/customers/:id", s.customerHandler.Update, s.authHandler.Authenticate)
	s.Echo.DELETE("/api/v1/customers/:id", s.customerHandler.Delete, s.authHandler.Authenticate)
}
