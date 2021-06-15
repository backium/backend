package app

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func (s *Server) setupRoutes() {
	s.Echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Println("calling middleware")
			cookie, err := c.Cookie("web_session")
			if err != nil {
				return next(c)
			}
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("backium"), nil
			})

			fmt.Println("token", token)
			fmt.Println("claims", claims)

			session, err := s.SessionStorage.Get(c.Request().Context(), claims["session_id"].(string))
			if err != nil {
				fmt.Println("error", err)
			}

			fmt.Println("session", session)

			return next(c)
		}
	})
	s.Echo.GET("/api/v1/merchants/:id", s.merchantHandler.Retrieve)
	s.Echo.GET("/api/v1/merchants", s.merchantHandler.ListAll)
	s.Echo.POST("/api/v1/merchants", s.merchantHandler.Create)
	s.Echo.PUT("/api/v1/merchants/:id", s.merchantHandler.Update)

	s.Echo.POST("/api/v1/signup", s.userHandler.Signup)
	s.Echo.POST("/api/v1/login", s.userHandler.Login)
}
