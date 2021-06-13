package merchants

type Merchant struct {
	ID           string
	FirstName    string
	LastName     string
	BusinessName string
}

type MerchantRepository interface {
	Create(Merchant) (Merchant, error)
	Update(Merchant) (Merchant, error)
	Retrieve(string) (Merchant, error)
	ListAll() ([]Merchant, error)
	Delete(string) (Merchant, error)
}
