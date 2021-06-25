package main

import (
	"log"

	"github.com/backium/backend/http"
	"github.com/backium/backend/repository/mongo"
	"github.com/backium/backend/repository/redis"
	"github.com/labstack/echo/v4"
)

func main() {
	config, err := http.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	log.Printf("%+v", config)
	db, err := mongo.New(config.DBURI, config.DBName)
	if err != nil {
		panic(err)
	}
	userRepository := mongo.NewUserRepository(db)
	merchantStorage := mongo.NewMerchantStorage(db)
	locationStorage := mongo.NewLocationStorage(db)
	customerStorage := mongo.NewCustomerStorage(db)
	categoryStorage := mongo.NewCategoryStorage(db)
	itemVariationStorage := mongo.NewItemVariationStorage(db)
	itemStorage := mongo.NewItemRepository(db)
	taxStorage := mongo.NewTaxStorage(db)
	discountStorage := mongo.NewDiscountStorage(db)
	orderStorage := mongo.NewOrderStorage(db)

	redis := redis.NewSessionRepository(config.RedisURI, config.RedisPassword)
	s := http.Server{
		Echo:                 echo.New(),
		DB:                   db,
		UserRepository:       userRepository,
		MerchantStorage:      merchantStorage,
		LocationStorage:      locationStorage,
		CustomerStorage:      customerStorage,
		CategoryStorage:      categoryStorage,
		ItemStorage:          itemStorage,
		ItemVariationStorage: itemVariationStorage,
		TaxStorage:           taxStorage,
		DiscountStorage:      discountStorage,
		OrderStorage:         orderStorage,
		SessionRepository:    redis,
	}
	s.Setup()
	s.ListenAndServe(config.Port)
}
