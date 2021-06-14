package app

import (
	"github.com/backium/backend/controller"
	"github.com/backium/backend/handler"
	"github.com/backium/backend/repository"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Echo            *echo.Echo
	DB              repository.MongoDB
	merchantHandler handler.Merchant
}

func (s *Server) Setup() {
	s.setupHandlers()
	s.setupRoutes()
}

func (s *Server) setupHandlers() {
	// setup dependencies
	merchantRepository := repository.NewMerchantMongoRepository(s.DB)

	// setup controllers
	merchantController := controller.Merchant{
		Repository: merchantRepository,
	}

	// setup handlers
	s.merchantHandler = handler.Merchant{
		Controller: merchantController,
	}
}

func (s *Server) ListenAndServe(port string) {
	s.Echo.Logger.Fatal(s.Echo.Start(":" + port))
}
