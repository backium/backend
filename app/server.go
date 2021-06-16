package app

import (
	"github.com/backium/backend/controller"
	"github.com/backium/backend/handler"
	"github.com/backium/backend/repository"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Server struct {
	Echo              *echo.Echo
	DB                repository.MongoDB
	merchantHandler   handler.Merchant
	authHandler       handler.Auth
	locationHandler   handler.Location
	customerHandler   handler.Customer
	SessionRepository handler.SessionRepository
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
	userRepository := repository.NewUserMongoRepository(s.DB)
	merchantRepository := repository.NewMerchantMongoRepository(s.DB)
	locationRepository := repository.NewLocationMongoRepository(s.DB)
	customerRepository := repository.NewCustomerMongoRepository(s.DB)

	// setup controllers
	merchantController := controller.Merchant{
		Repository: merchantRepository,
	}
	userController := controller.User{
		Repository:         userRepository,
		MerchantRepository: merchantRepository,
		LocationRepository: locationRepository,
	}
	locationController := controller.Location{
		Repository: locationRepository,
	}
	customerController := controller.Customer{
		Repository: customerRepository,
	}

	// setup handlers
	s.authHandler = handler.Auth{
		Controller:        userController,
		SessionRepository: s.SessionRepository,
	}
	s.merchantHandler = handler.Merchant{
		Controller: merchantController,
	}
	s.locationHandler = handler.Location{
		Controller: locationController,
	}
	s.customerHandler = handler.Customer{
		Controller: customerController,
	}
}

func (s *Server) ListenAndServe(port string) {
	s.Echo.Logger.Fatal(s.Echo.Start(":" + port))
}
