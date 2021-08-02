package mysql

import (
	"time"

	"gitlab.com/gae4/trade-engine/models"
)

func (s *Store) AddOrder(order *models.Order) error {
	order.UpdatedAt = time.Now()
	return s.db.Save(order).Error
}
