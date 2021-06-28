package http

import (
	"github.com/backium/backend/core"
	"github.com/backium/backend/storage/mongo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Server struct {
	Echo                 *echo.Echo
	DB                   mongo.DB
	Handler              Handler
	UserRepository       core.UserRepository
	MerchantStorage      core.MerchantStorage
	LocationStorage      core.LocationStorage
	CustomerStorage      core.CustomerStorage
	CategoryStorage      core.CategoryStorage
	ItemStorage          core.ItemStorage
	ItemVariationStorage core.ItemVariationStorage
	TaxStorage           core.TaxStorage
	DiscountStorage      core.DiscountStorage
	OrderStorage         core.OrderStorage
	InventoryStorage     core.InventoryStorage
	SessionRepository    SessionRepository
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
	locationService := core.LocationService{LocationStorage: s.LocationStorage}
	customerService := core.CustomerService{CustomerStorage: s.CustomerStorage}
	merchantService := core.MerchantService{MerchantStorage: s.MerchantStorage}
	userService := core.UserService{
		UserRepository:  s.UserRepository,
		MerchantStorage: s.MerchantStorage,
		LocationStorage: s.LocationStorage,
	}
	catalogService := core.CatalogService{
		CategoryStorage:      s.CategoryStorage,
		ItemStorage:          s.ItemStorage,
		ItemVariationStorage: s.ItemVariationStorage,
		TaxStorage:           s.TaxStorage,
		DiscountStorage:      s.DiscountStorage,
		InventoryStorage:     s.InventoryStorage,
		LocationStorage:      s.LocationStorage,
	}
	orderingService := core.OrderingService{
		OrderStorage:         s.OrderStorage,
		ItemVariationStorage: s.ItemVariationStorage,
		TaxStorage:           s.TaxStorage,
		DiscountStorage:      s.DiscountStorage,
	}

	// setup handlers
	s.Handler = Handler{
		LocationService:   locationService,
		CustomerService:   customerService,
		MerchantService:   merchantService,
		UserService:       userService,
		CatalogService:    catalogService,
		OrderingService:   orderingService,
		SessionRepository: s.SessionRepository,
	}
}

func (s *Server) ListenAndServe(port string) {
	s.Echo.Logger.Fatal(s.Echo.Start(":" + port))
}
