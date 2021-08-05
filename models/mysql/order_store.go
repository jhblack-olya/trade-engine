package mysql

import (
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.com/gae4/trade-engine/models"
)

func (s *Store) AddOrder(order *models.Order) error {
	order.UpdatedAt = time.Now()
	return s.db.Save(order).Error
}

func (s *Store) GetOrderByClientOid(userId int64, clientOid string) (*models.Order, error) {
	var order models.Order
	err := s.db.Raw("SELECT * FROM g_order WHERE user_id=? AND client_oid=?", userId, clientOid).Scan(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

func (s *Store) GetOrderByIdForUpdate(orderId int64) (*models.Order, error) {
	var order models.Order
	err := s.db.Raw("SELECT * FROM g_order WHERE id=? FOR UPDATE", orderId).Scan(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

func (s *Store) UpdateOrder(order *models.Order) error {
	order.UpdatedAt = time.Now()
	return s.db.Save(order).Error
}

func (s *Store) GetOrderById(orderId int64) (*models.Order, error) {
	var order models.Order
	err := s.db.Raw("SELECT * FROM g_order WHERE id=?", orderId).Scan(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

func (s *Store) UpdateOrderStatus(orderId int64, oldStatus, newStatus models.OrderStatus) (bool, error) {
	ret := s.db.Exec("UPDATE g_order SET `status`=?,updated_at=? WHERE id=? AND `status`=? ", newStatus, time.Now(), orderId, oldStatus)
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}
