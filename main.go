package main

import (
	"log"

	"github.com/backium/backend/cloudinary"
	"github.com/backium/backend/http"
	"github.com/backium/backend/storage/mongo"
	"github.com/backium/backend/storage/redis"
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
		log.Fatalf("mongodb: %v", err)
	}
	uploader, err := cloudinary.New(config.CloudinaryURI)
	if err != nil {
		log.Fatalf("cloudinary: %v", err)
	}
	userRepository := mongo.NewUserRepository(db)
	employeeStorage := mongo.NewEmployeeStorage(db)
	merchantStorage := mongo.NewMerchantStorage(db)
	locationStorage := mongo.NewLocationStorage(db)
	customerStorage := mongo.NewCustomerStorage(db)
	categoryStorage := mongo.NewCategoryStorage(db)
	itemVariationStorage := mongo.NewItemVariationStorage(db)
	itemStorage := mongo.NewItemRepository(db)
	taxStorage := mongo.NewTaxStorage(db)
	discountStorage := mongo.NewDiscountStorage(db)
	orderStorage := mongo.NewOrderStorage(db)
	paymentStorage := mongo.NewPaymentStorage(db)
	inventoryStorage := mongo.NewInventoryStorage(db)
	cashDrawerStorage := mongo.NewCashDrawerStorage(db)

	redis := redis.NewSessionRepository(config.RedisURI, config.RedisPassword)
	s := http.Server{
		Echo:                 echo.New(),
		DB:                   db,
		UserStorage:          userRepository,
		EmployeeStorage:      employeeStorage,
		MerchantStorage:      merchantStorage,
		LocationStorage:      locationStorage,
		CustomerStorage:      customerStorage,
		CategoryStorage:      categoryStorage,
		ItemStorage:          itemStorage,
		ItemVariationStorage: itemVariationStorage,
		TaxStorage:           taxStorage,
		DiscountStorage:      discountStorage,
		OrderStorage:         orderStorage,
		PaymentStorage:       paymentStorage,
		InventoryStorage:     inventoryStorage,
		CashDrawerStorage:    cashDrawerStorage,
		SessionRepository:    redis,
		Uploader:             uploader,
	}
	s.Setup()
	s.ListenAndServe(config.Port)
}
