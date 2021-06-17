package app

import (
	"github.com/backium/backend/controller"
	"github.com/backium/backend/handler"
	"github.com/backium/backend/repository/mongo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Server struct {
	Echo              *echo.Echo
	DB                mongo.DB
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
	s.Echo.HTTPErrorHandler = errorHandler
	s.setupHandlers()
	s.setupRoutes()
	return nil
}

func (s *Server) setupHandlers() {
	// setup dependencies
	userRepository := mongo.NewUserRepository(s.DB)
	merchantRepository := mongo.NewMerchantRepository(s.DB)
	locationRepository := mongo.NewLocationRepository(s.DB)
	customerRepository := mongo.NewCustomerRepository(s.DB)

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
