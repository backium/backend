package controller

import "github.com/backium/backend/entity"

type Merchant struct {
	Repository MerchantRepository
}

func (c *Merchant) Create(m entity.Merchant) (entity.Merchant, error) {
	return c.Repository.Create(m)
}

func (c *Merchant) Update(m entity.Merchant) (entity.Merchant, error) {
	return c.Repository.Update(m)
}

func (c *Merchant) Retrieve(id string) (entity.Merchant, error) {
	return c.Repository.Retrieve(id)
}

func (c *Merchant) ListAll() ([]entity.Merchant, error) {
	return c.Repository.ListAll()
}

type MerchantRepository interface {
	Create(entity.Merchant) (entity.Merchant, error)
	Update(entity.Merchant) (entity.Merchant, error)
	Retrieve(string) (entity.Merchant, error)
	ListAll() ([]entity.Merchant, error)
	Delete(string) (entity.Merchant, error)
}
