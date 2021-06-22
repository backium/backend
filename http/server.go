package http

import (
	"github.com/backium/backend/core"
	"github.com/backium/backend/repository/mongo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Server struct {
	Echo                    *echo.Echo
	DB                      mongo.DB
	Handler                 Handler
	UserRepository          core.UserRepository
	MerchantRepository      core.MerchantRepository
	LocationRepository      core.LocationRepository
	CustomerRepository      core.CustomerRepository
	CategoryRepository      core.CategoryRepository
	ItemRepository          core.ItemRepository
	ItemVariationRepository core.ItemVariationStorage
	TaxRepository           core.TaxRepository
	SessionRepository       SessionRepository
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

	// setup services
	locationService := core.LocationService{LocationRepository: s.LocationRepository}
	customerService := core.CustomerService{CustomerRepository: s.CustomerRepository}
	merchantService := core.MerchantService{MerchantRepository: s.MerchantRepository}
	userService := core.UserService{
		UserRepository:     s.UserRepository,
		MerchantRepository: s.MerchantRepository,
		LocationRepository: s.LocationRepository,
	}
	catalogService := core.CatalogService{
		CategoryRepository:      s.CategoryRepository,
		ItemRepository:          s.ItemRepository,
		ItemVariationRepository: s.ItemVariationRepository,
		TaxRepository:           s.TaxRepository,
	}

	// setup handlers
	s.Handler = Handler{
		LocationService:   locationService,
		CustomerService:   customerService,
		MerchantService:   merchantService,
		UserService:       userService,
		CatalogService:    catalogService,
		SessionRepository: s.SessionRepository,
	}
}

func (s *Server) ListenAndServe(port string) {
	s.Echo.Logger.Fatal(s.Echo.Start(":" + port))
}
