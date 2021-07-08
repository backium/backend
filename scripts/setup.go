package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/backium/backend/core"
	"github.com/backium/backend/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	mongoCfg MongoConfig
)

type MongoConfig struct {
	URI  string
	Name string
}

func init() {
	mongoURI := flag.String("mongo-uri", "", "mongodb uri")
	mongoName := flag.String("mongo-name", "", "mongodb database name")

	flag.Parse()

	mongoCfg = MongoConfig{URI: *mongoURI, Name: *mongoName}
}

func main() {
	log.Println("Setting up local development environment ...")

	log.Println("Connecting to mongodb ...")
	db, err := mongo.New(mongoCfg.URI, mongoCfg.Name)
	if err != nil {
		log.Fatalf("Could not connect to mongoDB: %v", err)
	}

	// Setting up services
	userRepository := mongo.NewUserRepository(db)
	employeeStorage := mongo.NewEmployeeStorage(db)
	merchantStorage := mongo.NewMerchantStorage(db)
	locationStorage := mongo.NewLocationStorage(db)
	//customerStorage := mongo.NewCustomerStorage(db)
	categoryStorage := mongo.NewCategoryStorage(db)
	itemVariationStorage := mongo.NewItemVariationStorage(db)
	itemStorage := mongo.NewItemRepository(db)
	taxStorage := mongo.NewTaxStorage(db)
	discountStorage := mongo.NewDiscountStorage(db)
	orderStorage := mongo.NewOrderStorage(db)
	paymentStorage := mongo.NewPaymentStorage(db)
	//inventoryStorage := mongo.NewInventoryStorage(db)

	userService := core.UserService{
		UserStorage:     userRepository,
		MerchantStorage: merchantStorage,
		LocationStorage: locationStorage,
		EmployeeStorage: employeeStorage,
	}

	orderingService := core.OrderingService{
		OrderStorage:         orderStorage,
		CategoryStorage:      categoryStorage,
		ItemStorage:          itemStorage,
		ItemVariationStorage: itemVariationStorage,
		TaxStorage:           taxStorage,
		DiscountStorage:      discountStorage,
		PaymentStorage:       paymentStorage,
	}

	log.Println("Droping old documents ...")
	mdb := db.Client().Database(mongoCfg.Name)
	if err := mdb.Drop(context.TODO()); err != nil {
		log.Fatalf("Could not drop database: %v", err)
	}

	ctx := context.Background()

	// Creating merchant account with one default location
	log.Println("Creating dev account ...")
	email := "test@mail.com"
	password := "Test@123!"

	user := core.NewUserOwner()
	user.Email = email

	user, err = userService.Create(ctx, user, password)
	if err != nil {
		log.Fatalf("Could not create new user: %v", err)
	}

	log.Printf("New user: email=%v, password=%v, merchant_id=%v", user.Email, password, user.MerchantID)

	// Fetch user location
	lq := core.LocationQuery{Filter: core.LocationFilter{MerchantID: user.MerchantID}}
	locations, _, err := locationStorage.List(ctx, lq)
	if err != nil {
		log.Fatalf("Could not fetch locations: %v", err)
	}
	locationIDs := []core.ID{locations[0].ID}

	// Creating merchant categories, items, variations, taxes and discounts
	log.Println("Creating merchant catalog ...")

	// Creating categories
	type catdata struct {
		name string
	}

	cdata := []catdata{
		{"Food"},
		{"Drinks"},
		{"Desserts"},
	}

	log.Printf("Creating %v categories ...", len(cdata))

	var categories []core.Category
	for _, cd := range cdata {
		c := core.NewCategory(cd.name, user.MerchantID)
		c.LocationIDs = locationIDs
		categories = append(categories, c)
	}

	if err := categoryStorage.PutBatch(ctx, categories); err != nil {
		log.Fatalf("Could not create categories: %v", err)
	}

	// Creating items
	type itemdata struct {
		name     string
		category core.ID
	}

	idata := []itemdata{
		// Food
		{"Aloha Crepe", categories[0].ID},
		{"Banana Crepe", categories[0].ID},
		{"Beef Quesadilla", categories[0].ID},
		{"American Cheese", categories[0].ID},
		{"Avocado Toast", categories[0].ID},
		{"Fetuccini Past", categories[0].ID},
		// Drinks
		{"Americano", categories[1].ID},
		{"Black Eye", categories[1].ID},
		{"Almond Milk", categories[1].ID},
		{"Chocolate Milk", categories[1].ID},
		{"Water", categories[1].ID},
		{"Boba Drinks", categories[1].ID},
		//Desserts
		{"Bars", categories[2].ID},
		{"Acai Bowls", categories[2].ID},
		{"Arroz con leche", categories[2].ID},
		{"Cookies", categories[2].ID},
		{"Coffe Cake", categories[2].ID},
		{"Yogurt", categories[2].ID},
	}

	log.Printf("Creating %v items ...", len(idata))

	var items []core.Item
	for _, id := range idata {
		i := core.NewItem(id.name, id.category, user.MerchantID)
		i.LocationIDs = locationIDs
		items = append(items, i)
	}

	if err := itemStorage.PutBatch(ctx, items); err != nil {
		log.Fatalf("Could not create items: %v", err)
	}

	// Creating item variations
	log.Printf("Creating %v item variations ...", len(idata))

	var variations []core.ItemVariation
	for _, item := range items {
		iv := core.NewItemVariation(item.Name, item.ID, user.MerchantID)
		price := 500 + rand.Int63n(5000)
		iv.Price = core.NewMoney(price, core.PEN)
		iv.LocationIDs = locationIDs
		variations = append(variations, iv)
	}

	if err := itemVariationStorage.PutBatch(ctx, variations); err != nil {
		log.Fatalf("Could not create variations: %v", err)
	}

	// Creating taxes
	type taxdata struct {
		name       string
		percentage float64
	}

	tdata := []taxdata{
		{"IGV", 9.5},
	}

	log.Printf("Creating %v taxes ...", len(tdata))

	var taxes []core.Tax
	for _, td := range tdata {
		t := core.NewTax(td.name, user.MerchantID)
		t.Percentage = td.percentage
		t.LocationIDs = locationIDs
		taxes = append(taxes, t)
	}

	if err := taxStorage.PutBatch(ctx, taxes); err != nil {
		log.Fatalf("Could not create taxes: %v", err)
	}

	// Creating discounts
	type discdata struct {
		name       string
		typ        core.DiscountType
		amount     int64
		percentage float64
	}

	ddata := []discdata{
		{"Employee On Shift", core.DiscountPercentage, 0, 25},
		{"Friends & Family", core.DiscountFixed, 500, 0},
		{"Polic & Fire", core.DiscountPercentage, 0, 20},
	}

	log.Printf("Creating %v discounts ...", len(ddata))

	var discounts []core.Discount
	for _, dd := range ddata {
		d := core.NewDiscount(dd.name, dd.typ, user.MerchantID)
		d.Amount = core.NewMoney(dd.amount, core.PEN)
		d.Percentage = dd.percentage
		d.LocationIDs = locationIDs
		discounts = append(discounts, d)
	}

	if err := discountStorage.PutBatch(ctx, discounts); err != nil {
		log.Fatalf("Could not create discounts: %v", err)
	}

	// Creating orders

	const (
		maxItems          = 5
		maxItemQuantity   = 3
		maxTaxes          = 1
		maxDiscounts      = 2
		maxOrderTimeRange = 60 * 24 * 60 * 60
		numOrders         = 500
	)

	log.Printf("Creating %v orders ...", numOrders)

	var orders []core.Order
	for i := 0; i < numOrders; i++ {
		schema := core.OrderSchema{
			Currency:   core.PEN,
			LocationID: locationIDs[0],
			MerchantID: user.MerchantID,
		}

		// Choose randomly how many line items to add
		in := rand.Int63()%maxItems + 1
		for vi := int64(0); vi < in; vi++ {
			idx := rand.Int63() % int64(len(variations))
			variation := variations[idx]
			schema.ItemVariations = append(schema.ItemVariations, core.OrderSchemaItemVariation{
				UID:      string(core.NewID("item_uid")),
				ID:       variation.ID,
				Quantity: rand.Int63()%maxItemQuantity + 1,
			})
		}

		// Choose randomly how many discounts to add
		dn := rand.Int63()%maxDiscounts + 1
		for di := int64(0); di < dn; di++ {
			idx := rand.Int63() % int64(len(discounts))
			discount := discounts[idx]
			schema.Discounts = append(schema.Discounts, core.OrderSchemaDiscount{
				UID: string(core.NewID("disc_uid")),
				ID:  discount.ID,
			})
		}

		// Choose randomly how many discounts to add
		tn := rand.Int63()%maxTaxes + 1
		for ti := int64(0); ti < tn; ti++ {
			idx := rand.Int63() % int64(len(taxes))
			tax := taxes[idx]
			schema.Taxes = append(schema.Taxes, core.OrderSchemaTax{
				UID: string(core.NewID("tax_uid")),
				ID:  tax.ID,
			})
		}

		order, err := orderingService.CreateOrder(ctx, schema)
		if err != nil {
			log.Fatalf("Could not create order %v: %v", i, err)
		}

		createdAt := time.Now().Unix() - (rand.Int63() % maxOrderTimeRange)
		order.CreatedAt = createdAt
		order.UpdatedAt = createdAt

		q := bson.M{"$set": order}
		coll := db.Client().Database(mongoCfg.Name).Collection("orders")
		if _, err := coll.UpdateByID(ctx, order.ID, q); err != nil {
			log.Fatalf("Could not update order %v: %v", i, err)
		}

		orders = append(orders, order)
	}
}
