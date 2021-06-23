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
	merchantRepository := mongo.NewMerchantRepository(db)
	locationRepository := mongo.NewLocationRepository(db)
	customerRepository := mongo.NewCustomerRepository(db)
	categoryRepository := mongo.NewCategoryRepository(db)
	itemRepository := mongo.NewItemRepository(db)
	itemVariationRepository := mongo.NewItemVariationRepository(db)
	taxStorage := mongo.NewTaxStorage(db)
	discountStorage := mongo.NewDiscountStorage(db)
	orderStorage := mongo.NewOrderStorage(db)

	redis := redis.NewSessionRepository(config.RedisURI, config.RedisPassword)
	s := http.Server{
		Echo:                    echo.New(),
		DB:                      db,
		UserRepository:          userRepository,
		MerchantRepository:      merchantRepository,
		LocationRepository:      locationRepository,
		CustomerRepository:      customerRepository,
		CategoryRepository:      categoryRepository,
		ItemRepository:          itemRepository,
		ItemVariationRepository: itemVariationRepository,
		TaxStorage:              taxStorage,
		DiscountStorage:         discountStorage,
		OrderStorage:            orderStorage,
		SessionRepository:       redis,
	}
	s.Setup()
	s.ListenAndServe(config.Port)
}
