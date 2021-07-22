package core

type Permission string

const (
	// UI permissions
	HomeAccess      Permission = "home_access"
	LocationAccess  Permission = "location_access"
	SalesAccess     Permission = "sales_access"
	InventoryAccess Permission = "inventory_access"
	StockAccess     Permission = "stock_access"
	GuestbookAccess Permission = "guestbook_access"
	// Write, read permissions
	CatalogWrite   Permission = "catalog_write"
	CatalogRead    Permission = "catalog_read"
	InventoryWrite Permission = "inventory_write"
	InventoryRead  Permission = "inventory_read"
	CustomerWrite  Permission = "customer_write"
	CustomerRead   Permission = "customer_read"
	LocationWrite  Permission = "location_write"
	LocationRead   Permission = "location_read"
	OrderWrite     Permission = "order_write"
	OrderRead      Permission = "order_read"
)

func Can(given []Permission, target Permission) bool {
	for _, p := range given {
		if p == target {
			return true
		}
	}
	return false
}
