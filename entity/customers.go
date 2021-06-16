package entity

type Customer struct {
	ID         string
	Name       string
	Email      string
	Phone      string
	Address    *Address
	MerchantID string
}
