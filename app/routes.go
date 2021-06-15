package app

func (s *Server) setupRoutes() {
	s.Echo.GET("/api/v1/merchants/:id", s.merchantHandler.Retrieve, s.authHandler.Authenticate)
	s.Echo.GET("/api/v1/merchants", s.merchantHandler.ListAll, s.authHandler.Authenticate)
	s.Echo.POST("/api/v1/merchants", s.merchantHandler.Create, s.authHandler.Authenticate)
	s.Echo.PUT("/api/v1/merchants/:id", s.merchantHandler.Update, s.authHandler.Authenticate)

	s.Echo.POST("/api/v1/signup", s.authHandler.Signup)
	s.Echo.POST("/api/v1/login", s.authHandler.Login)
	s.Echo.POST("/api/v1/signout", s.authHandler.Signout, s.authHandler.Authenticate)
}
