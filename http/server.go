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
	UserStorage          core.UserStorage
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
	CashDrawerStorage    core.CashDrawerStorage
	SessionRepository    core.SessionStorage
	Uploader             core.Uploader
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
	authorizer := core.Authorizer{
		ItemStorage:          s.ItemStorage,
		ItemVariationStorage: s.ItemVariationStorage,
		CategoryStorage:      s.CategoryStorage,
		EmployeeStorage:      s.EmployeeStorage,
	}
	locationService := core.LocationService{
		LocationStorage:      s.LocationStorage,
		CashDrawerStorage:    s.CashDrawerStorage,
		ItemVariationStorage: s.ItemVariationStorage,
		InventoryStorage:     s.InventoryStorage,
	}
	customerService := core.CustomerService{CustomerStorage: s.CustomerStorage}
	merchantService := core.MerchantService{MerchantStorage: s.MerchantStorage}
	userService := core.UserService{
		UserStorage:       s.UserStorage,
		EmployeeStorage:   s.EmployeeStorage,
		MerchantStorage:   s.MerchantStorage,
		LocationStorage:   s.LocationStorage,
		CashDrawerStorage: s.CashDrawerStorage,
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
		ItemStorage:          s.ItemStorage,
		CategoryStorage:      s.CategoryStorage,
		ItemVariationStorage: s.ItemVariationStorage,
		TaxStorage:           s.TaxStorage,
		DiscountStorage:      s.DiscountStorage,
		LocationStorage:      s.LocationStorage,
		CustomerStorage:      s.CustomerStorage,
		CashDrawerStorage:    s.CashDrawerStorage,
		InventoryStorage:     s.InventoryStorage,
		Uploader:             s.Uploader,
	}
	paymentService := core.PaymentService{
		PaymentStorage: s.PaymentStorage,
	}
	reportService := core.ReportService{
		OrderStorage:         s.OrderStorage,
		ItemStorage:          s.ItemStorage,
		ItemVariationStorage: s.ItemVariationStorage,
		CategoryStorage:      s.CategoryStorage,
	}

	// setup handlers
	s.Handler = Handler{
		Authorizer:        authorizer,
		LocationService:   locationService,
		CustomerService:   customerService,
		MerchantService:   merchantService,
		UserService:       userService,
		EmployeeService:   employeeService,
		CatalogService:    catalogService,
		OrderingService:   orderingService,
		PaymentService:    paymentService,
		ReportService:     reportService,
		SessionRepository: s.SessionRepository,
	}
}

func (s *Server) ListenAndServe(port string) {
	s.Echo.Logger.Fatal(s.Echo.Start(":" + port))
}
