package api

func (s *Server) setupRoutes() {
	s.Echo.GET("/api/v1/merchants/:id", s.merchantHandler.Retrieve)
	s.Echo.GET("/api/v1/merchants", s.merchantHandler.ListAll)
	s.Echo.POST("/api/v1/merchants", s.merchantHandler.Create)
	s.Echo.PUT("/api/v1/merchants/:id", s.merchantHandler.Update)
}
