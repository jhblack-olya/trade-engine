package mysql

import "gitlab.com/gae4/trade-engine/models"

func (s *Store) GetTicksByProductId(productId string, granularity int64, limit int) ([]*models.Tick, error) {
	db := s.db.Where("product_id =?", productId).Where("granularity=?", granularity).
		Order("time DESC").Limit(limit)
	var ticks []*models.Tick
	err := db.Find(&ticks).Error
	return ticks, err
}
