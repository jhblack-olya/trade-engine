package mysql

import (
	"time"

	"gitlab.com/gae4/trade-engine/models"
)

func (s *Store) AddOrder(order *models.Order) error {
	order.CreatedAt = time.Now()
	return s.db.Create(order).Error
}
