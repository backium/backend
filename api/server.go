package api

import (
	"fmt"

	"github.com/backium/backend/merchants"
	"github.com/backium/backend/repository"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Echo            *echo.Echo
	DB              repository.MongoDB
	merchantHandler merchantHandler
}

func (s *Server) Setup() {
	s.setupHandlers()
	s.setupRoutes()
}

func (s *Server) setupHandlers() {
	// setup dependencies
	merchantRepository := repository.NewMerchantMongoRepository(s.DB)

	// setup controllers
	merchantController := merchants.MerchantController{
		Repository: merchantRepository,
	}

	// setup handlers
	s.merchantHandler = merchantHandler{
		controller: merchantController,
	}
}

func (s *Server) ListenAndServe(port int) {
	s.Echo.Logger.Fatal(s.Echo.Start(fmt.Sprintf(":%v", port)))
}
