package main

import (
	"os"

	"github.com/backium/backend/api"
	"github.com/backium/backend/repository"
	"github.com/labstack/echo/v4"
)

func main() {
	uri := os.Getenv("MONGO_URI")
	db, err := repository.NewMongoDB(uri, "testing")
	if err != nil {
		panic(err)
	}
	s := api.Server{
		Echo: echo.New(),
		DB:   db,
	}
	s.Setup()
	s.ListenAndServe(3000)
}
