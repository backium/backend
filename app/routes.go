package app

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func (s *Server) setupRoutes() {
	s.Echo.Use(middleware.CORS())
	s.Echo.Use(s.loggerMiddleware)
	userGroup := s.Echo.Group("/api/v1", s.authHandler.Authenticate)
	pubGroup := s.Echo.Group("/api/v1")

	userGroup.GET("/merchants/:id", s.merchantHandler.Retrieve)
	userGroup.GET("/merchants", s.merchantHandler.ListAll)
	userGroup.POST("/merchants", s.merchantHandler.Create)
	userGroup.PUT("/merchants/:id", s.merchantHandler.Update)

	pubGroup.POST("/signup", s.authHandler.Signup)
	pubGroup.POST("/login", s.authHandler.Login)
	userGroup.POST("/signout", s.authHandler.Signout)

	userGroup.GET("/locations/:id", s.locationHandler.Retrieve)
	userGroup.GET("/locations", s.locationHandler.ListAll)
	userGroup.POST("/locations", s.locationHandler.Create)
	userGroup.PUT("/locations/:id", s.locationHandler.Update)
	userGroup.DELETE("/locations/:id", s.locationHandler.Delete)

	userGroup.GET("/customers/:id", s.customerHandler.Retrieve)
	userGroup.GET("/customers", s.customerHandler.ListAll)
	userGroup.POST("/customers", s.customerHandler.Create)
	userGroup.PUT("/customers/:id", s.customerHandler.Update)
	userGroup.DELETE("/customers/:id", s.customerHandler.Delete)

	userGroup.GET("/categories/:id", s.categoryHandler.Retrieve)
	userGroup.GET("/categories", s.categoryHandler.ListAll)
	userGroup.POST("/categories", s.categoryHandler.Create)
	userGroup.PUT("/categories/:id", s.categoryHandler.Update)
	userGroup.DELETE("/categories/:id", s.categoryHandler.Delete)
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
