package app

import (
	"github.com/backium/backend/controller"
	"github.com/backium/backend/handler"
	"github.com/backium/backend/repository"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Server struct {
	Echo            *echo.Echo
	DB              repository.MongoDB
	merchantHandler handler.Merchant
	userHandler     handler.User
	SessionStorage  handler.SessionRepository
}

func (s *Server) Setup() error {
	v, err := NewValidator()
	if err != nil {
		return err
	}
	s.Echo.Validator = v
	s.Echo.Logger.SetLevel(log.INFO)
	s.setupHandlers()
	s.setupRoutes()
	return nil
}

func (s *Server) setupHandlers() {
	// setup dependencies
	merchantRepository := repository.NewMerchantMongoRepository(s.DB)
	userRepository := repository.NewUserMongoRepository(s.DB)

	// setup controllers
	merchantController := controller.Merchant{
		Repository: merchantRepository,
	}
	userController := controller.User{
		Repository:         userRepository,
		MerchantRepository: merchantRepository,
	}

	// setup handlers
	s.merchantHandler = handler.Merchant{
		Controller: merchantController,
	}
	s.userHandler = handler.User{
		Controller:     userController,
		SessionStorage: s.SessionStorage,
	}
}

func (s *Server) ListenAndServe(port string) {
	s.Echo.Logger.Fatal(s.Echo.Start(":" + port))
}
