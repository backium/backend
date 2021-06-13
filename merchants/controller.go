package merchants

type MerchantController struct {
	Repository MerchantRepository
}

func (c *MerchantController) Create(m Merchant) (Merchant, error) {
	return c.Repository.Create(m)
}

func (c *MerchantController) Update(m Merchant) (Merchant, error) {
	return c.Repository.Update(m)
}

func (c *MerchantController) Retrieve(id string) (Merchant, error) {
	return c.Repository.Retrieve(id)
}

func (c *MerchantController) ListAll() ([]Merchant, error) {
	return c.Repository.ListAll()
}
