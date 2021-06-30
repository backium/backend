package limits

var (
	MaxMap = map[Limit]int64{
		Item:          50,
		ItemVariation: 50,
		Category:      50,
		Order:         50,
	}
)

type Limit int

const (
	Item Limit = iota
	ItemVariation
	Category
	Order
)
