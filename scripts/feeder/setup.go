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

	var db mongo.DB
	for {
		d, err := mongo.New(mongoCfg.URI, mongoCfg.Name)
		if err == nil {
			db = d
			break
		}
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
	inventoryStorage := mongo.NewInventoryStorage(db)
	cashDrawerStorage := mongo.NewCashDrawerStorage(db)
	customerStorage := mongo.NewCustomerStorage(db)

	userService := core.UserService{
		UserStorage:       userRepository,
		MerchantStorage:   merchantStorage,
		LocationStorage:   locationStorage,
		EmployeeStorage:   employeeStorage,
		CashDrawerStorage: cashDrawerStorage,
	}

	catalogService := core.CatalogService{
		ItemVariationStorage: itemVariationStorage,
		InventoryStorage:     inventoryStorage,
		LocationStorage:      locationStorage,
	}

	orderingService := core.OrderingService{
		OrderStorage:         orderStorage,
		CategoryStorage:      categoryStorage,
		ItemStorage:          itemStorage,
		ItemVariationStorage: itemVariationStorage,
		TaxStorage:           taxStorage,
		DiscountStorage:      discountStorage,
		CustomerStorage:      customerStorage,
		CashDrawerStorage:    cashDrawerStorage,
		InventoryStorage:     inventoryStorage,
		PaymentStorage:       paymentStorage,
	}

	paymentService := core.PaymentService{
		PaymentStorage: paymentStorage,
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

	user, err := userService.Create(ctx, user, password)
	if err != nil {
		log.Fatalf("Could not create new user: %v", err)
	}

	merchant, _ := merchantStorage.Get(ctx, user.MerchantID)

	log.Printf("New user: email=%v, password=%v, merchant_id=%v", user.Email, password, user.MerchantID)

	ctx = core.ContextWithUser(ctx, &user)
	ctx = core.ContextWithMerchant(ctx, &merchant)

	// Fetch user location
	lq := core.LocationQuery{Filter: core.LocationFilter{MerchantID: user.MerchantID}}
	locations, _, err := locationStorage.List(ctx, lq)
	if err != nil {
		log.Fatalf("Could not fetch locations: %v", err)
	}
	locationIDs := []core.ID{locations[0].ID}

	// Creating merchant categories, items, variations, taxes and discounts
	log.Println("Creating merchant catalog ...")

	// Creating customers
	type customerdata struct {
		name  string
		email string
	}

	custdata := []customerdata{
		{"Alex Harper", "alex.harper@mail.com"},
		{"Saul Quispe", "saul.quispe@mail.com"},
		{"Ravi Dahr", "ravi.dahr@mail.com"},
		{"Timoteo Zurita", "timoteo.zurita@mail.com"},
		{"Diana Frias", "diana.frias@mail.com"},
		{"Jenny Rueda", "jenny.ruedas@mail.com"},
	}

	log.Printf("Creating %v customers ...", len(custdata))

	var customers []core.Customer
	for _, cd := range custdata {
		c := core.NewCustomer(cd.name, cd.email, user.MerchantID)
		customers = append(customers, c)
		if err := customerStorage.Put(ctx, c); err != nil {
			log.Fatalf("Could not create categories: %v", err)
		}
	}

	// Creating categories
	type catdata struct {
		name string
	}

	cdata := []catdata{
		{"Panes"},
	}

	log.Printf("Creating %v categories ...", len(cdata))

	var categories []core.Category
	for _, cd := range cdata {
		c := core.NewCategory(cd.name, user.MerchantID)
		categories = append(categories, c)
	}

	if err := categoryStorage.PutBatch(ctx, categories); err != nil {
		log.Fatalf("Could not create categories: %v", err)
	}

	// Creating items
	type itemdata struct {
		name          string
		cost          int64
		price         int64
		minimum_stock int64
		category      core.ID
	}

	idata := []itemdata{
		// Panes
		{"Tartaleta de Fresa", 175, 350, 4, categories[0].ID},
		{"Porcion de Budin", 150, 300, 4, categories[0].ID},
		{"Mil Hojas de manjar", 150, 300, 4, categories[0].ID},
		{"Tartaleta de Manzana", 175, 350, 4, categories[0].ID},
		{"Orejas", 125, 250, 4, categories[0].ID},
		{"Empanada de Pollo", 200, 400, 4, categories[0].ID},
		{"Empanada de Carne", 200, 400, 4, categories[0].ID},
		{"Crema Volteada", 150, 300, 4, categories[0].ID},
		{"Leche Asada", 150, 300, 4, categories[0].ID},
		{"Empanadas Mixtas", 150, 350, 4, categories[0].ID},
		{"Tres leches vainilla", 200, 450, 4, categories[0].ID},
		{"Tres leches chocolate", 200, 450, 4, categories[0].ID},
		{"Torta Chocolate", 200, 450, 4, categories[0].ID},
		{"Mil Hojas de Fresa", 200, 450, 4, categories[0].ID},
		{"Pionono de Fresa", 200, 450, 4, categories[0].ID},
		{"Torta Helada", 200, 450, 4, categories[0].ID},
		{"Cheesecake Maracuya", 200, 450, 4, categories[0].ID},
		{"Cheesecake Sauco", 200, 450, 4, categories[0].ID},
		{"Cheesecake Fresa", 200, 450, 4, categories[0].ID},
		{"Pye de Limon", 200, 400, 4, categories[0].ID},
		{"Empanadas de Salchicha", 200, 400, 4, categories[0].ID},
		{"Alfajor Pequeño", 7, 20, 4, categories[0].ID},
		{"Alfajor", 30, 75, 4, categories[0].ID},
		{"Relampagos", 100, 300, 4, categories[0].ID},
		{"Poseidon", 100, 250, 4, categories[0].ID},
		{"Pizza", 175, 350, 4, categories[0].ID},
		{"Churros", 50, 200, 4, categories[0].ID},
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
	for i, item := range items {
		iv := core.NewItemVariation(item.Name, item.ID, user.MerchantID)
		iv.Price = core.NewMoney(idata[i].price, core.PEN)
		iv.Cost = &core.Money{Value: idata[i].cost, Currency: core.PEN}
		iv.MinimumRequiredStock = idata[i].minimum_stock
		//iv.Measurement = core.PerItem
		iv.LocationIDs = locationIDs
		variations = append(variations, iv)
	}

	for _, v := range variations {
		if _, err := catalogService.PutItemVariation(ctx, v); err != nil {
			log.Fatalf("Could not create variations: %v", err)
		}
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
		{"Friends & Family", core.DiscountFixed, 200, 0},
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
		numOrders         = 1000
	)

	log.Printf("Creating %v orders ...", numOrders)

	var orders []core.Order
	for i := 0; i < numOrders; i++ {
		idx := rand.Int63() % int64(len(customers))
		schema := core.OrderSchema{
			CustomerID: customers[idx].ID,
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
	log.Printf("Paying orders ...")

	for i := 0; i < len(orders)/2; i++ {
		p := core.NewPayment(core.PaymentCash, orders[i].ID, user.MerchantID, locationIDs[0])
		p.Amount = core.NewMoney(orders[i].TotalAmount.Value, core.PEN)
		p.TipAmount = core.NewMoney(0, core.PEN)
		if _, err := paymentService.CreatePayment(ctx, p); err != nil {
			log.Fatalf("Could not create payment %v: %v", i, err)
		}
		if _, err := orderingService.PayOrder(ctx, orders[i].ID, []core.ID{p.ID}); err != nil {
			log.Fatalf("Could not pay order %v: %v", i, err)
		}
	}

}
