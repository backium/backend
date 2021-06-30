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
	UserRepository       core.UserStorage
	MerchantStorage      core.MerchantStorage
	LocationStorage      core.LocationStorage
	CustomerStorage      core.CustomerStorage
	CategoryStorage      core.CategoryStorage
	ItemStorage          core.ItemStorage
	ItemVariationStorage core.ItemVariationStorage
	TaxStorage           core.TaxStorage
	DiscountStorage      core.DiscountStorage
	OrderStorage         core.OrderStorage
	PaymentStorage       core.PaymentStorage
	InventoryStorage     core.InventoryStorage
	EmployeeStorage      core.EmployeeStorage
	SessionRepository    core.SessionRepository
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
		UserStorage:     s.UserRepository,
		EmployeeStorage: s.EmployeeStorage,
		MerchantStorage: s.MerchantStorage,
		LocationStorage: s.LocationStorage,
	}
	employeeService := core.EmployeeService{
		EmployeeStorage: s.EmployeeStorage,
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
		PaymentStorage:       s.PaymentStorage,
		ItemVariationStorage: s.ItemVariationStorage,
		TaxStorage:           s.TaxStorage,
		DiscountStorage:      s.DiscountStorage,
	}
	paymentService := core.PaymentService{
		PaymentStorage: s.PaymentStorage,
	}

	// setup handlers
	s.Handler = Handler{
		LocationService:   locationService,
		CustomerService:   customerService,
		MerchantService:   merchantService,
		UserService:       userService,
		EmployeeService:   employeeService,
		CatalogService:    catalogService,
		OrderingService:   orderingService,
		PaymentService:    paymentService,
		SessionRepository: s.SessionRepository,
	}
}

func (s *Server) ListenAndServe(port string) {
	s.Echo.Logger.Fatal(s.Echo.Start(":" + port))
}
