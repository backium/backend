package core

type Merchant struct {
	ID           string `bson:"_id"`
	FirstName    string `bson:"first_name,omitempty"`
	LastName     string `bson:"last_name,omitempty"`
	BusinessName string `bson:"business_name,omitempty"`
}

type MerchantRepository interface {
	Create(Merchant) (Merchant, error)
	Update(Merchant) (Merchant, error)
	Retrieve(string) (Merchant, error)
	ListAll() ([]Merchant, error)
	Delete(string) (Merchant, error)
}

type MerchantService struct {
	MerchantRepository MerchantRepository
}

func (c *MerchantService) Create(m Merchant) (Merchant, error) {
	return c.MerchantRepository.Create(m)
}

func (c *MerchantService) Update(m Merchant) (Merchant, error) {
	return c.MerchantRepository.Update(m)
}

func (c *MerchantService) Retrieve(id string) (Merchant, error) {
	return c.MerchantRepository.Retrieve(id)
}

func (c *MerchantService) ListAll() ([]Merchant, error) {
	return c.MerchantRepository.ListAll()
}
