package mysql

import (
	"github.com/jinzhu/gorm"
	"gitlab.com/gae4/trade-engine/models"
)

func (s *Store) GetProductById(id string) (*models.Product, error) {
	var product models.Product
	err := s.db.Raw("SELECT * FROM g_product WHERE id=?", id).Scan(&product).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &product, err
}

func (s *Store) GetProducts() ([]*models.Product, error) {
	var products []*models.Product
	err := s.db.Find(&products).Error
	return products, err
}
