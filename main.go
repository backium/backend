package main

import (
	"github.com/backium/backend/app"
	"github.com/backium/backend/repository"
	"github.com/labstack/echo/v4"
)

func main() {
	config, err := app.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	db, err := repository.NewMongoDB(config.DBURI, config.DBName)
	if err != nil {
		panic(err)
	}
	s := app.Server{
		Echo: echo.New(),
		DB:   db,
	}
	s.Setup()
	s.ListenAndServe(config.Port)
}
