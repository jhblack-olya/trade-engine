package mysql

import "gitlab.com/gae4/trade-engine/models"

func (s *Store) GetUnsettledFillsByOrderId(orderId int64) ([]*models.Fill, error) {
	db := s.db.Where("settled =?", 0).Where("order_id=?", orderId).
		Order("id ASC").Limit(100)

	var fills []*models.Fill
	err := db.Find(&fills).Error
	return fills, err
}

func (s *Store) UpdateFill(fill *models.Fill) error {
	return s.db.Save(fill).Error
}

func (s *Store) GetUnsettledFills(count int32) ([]*models.Fill, error) {
	db := s.db.Where("settled =?", 0).Order("id ASC").Limit(count)

	var fills []*models.Fill
	err := db.Find(&fills).Error
	return fills, err
}
