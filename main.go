package main

import (
	"log"

	"github.com/backium/backend/app"
	"github.com/backium/backend/repository/mongo"
	"github.com/backium/backend/repository/redis"
	"github.com/labstack/echo/v4"
)

func main() {
	config, err := app.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	log.Printf("%+v", config)
	db, err := mongo.New(config.DBURI, config.DBName)
	if err != nil {
		panic(err)
	}
	redis := redis.NewSessionRepository(config.RedisURI, config.RedisPassword)
	s := app.Server{
		Echo:              echo.New(),
		DB:                db,
		SessionRepository: redis,
	}
	s.Setup()
	s.ListenAndServe(config.Port)
}
