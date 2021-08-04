package mysql

import (
	"time"

	"gitlab.com/gae4/trade-engine/models"
)

func (s *Store) AddOrder(order *models.Order) error {
	order.UpdatedAt = time.Now()
	return s.db.Save(order).Error
}

func (s *Store) UpdateOrderStatus(orderId int64, oldStatus, newStatus models.OrderStatus) (bool, error) {
	ret := s.db.Exec("UPDATE g_order SET `status`=?,updated_at=? WHERE id=? AND `status`=? ", newStatus, time.Now(), orderId, oldStatus)
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}
