package app

import (
	"net/http"

	"github.com/backium/backend/handler"
	"github.com/labstack/echo/v4"
)

func (s *Server) setupRoutes() {
	s.Echo.GET("/api/v1/merchants/:id", s.merchantHandler.Retrieve, s.authenticate)
	s.Echo.GET("/api/v1/merchants", s.merchantHandler.ListAll, s.authenticate)
	s.Echo.POST("/api/v1/merchants", s.merchantHandler.Create, s.authenticate)
	s.Echo.PUT("/api/v1/merchants/:id", s.merchantHandler.Update, s.authenticate)

	s.Echo.POST("/api/v1/signup", s.userHandler.Signup)
	s.Echo.POST("/api/v1/login", s.userHandler.Login)
	s.Echo.POST("/api/v1/signout", s.userHandler.Signout, s.authenticate)
}

// TODO: add session information to echo.Context
func (s *Server) authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("web_session")
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: session cookie missing")
		}
		ds, err := handler.DecodeSession(cookie.Value)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: error parsing session")
		}
		rs, err := s.SessionStorage.Get(c.Request().Context(), ds.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: session not found")
		}
		c.Logger().Infof("session found: %+v", rs)

		return next(&handler.AuthContext{
			Context: c,
			Session: rs,
		})
	}
}
