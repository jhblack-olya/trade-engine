/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package mysql

import (
	"fmt"
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
	err := s.db.Raw("SELECT * FROM OrderBooks WHERE user_id=? AND client_oid=?", userId, clientOid).Scan(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

func (s *Store) GetOrderByIdForUpdate(orderId int64) (*models.Order, error) {
	var order models.Order
	err := s.db.Raw("SELECT * FROM OrderBooks WHERE id=? FOR UPDATE", orderId).Scan(&order).Error
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
	err := s.db.Raw("SELECT * FROM OrderBooks WHERE id=?", orderId).Scan(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

func (s *Store) UpdateOrderStatus(orderId int64, oldStatus, newStatus models.OrderStatus, timer int64) (bool, error) {
	ret := s.db.Exec("UPDATE OrderBooks SET `status`=?,updated_at=?,expires_in=? WHERE id=? AND `status`=? ", newStatus, time.Now(), timer, orderId, oldStatus)
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}

func (s *Store) GetOpenLimitOrderByArt(side, art string) ([]*models.EstimateValue, error) {
	var orders []*models.EstimateValue
	fmt.Println(side, art)
	err := s.db.Raw("SELECT price,sum(size-filledArtBits)as quantity FROM OrderBooks WHERE side=? and art=? and  type=? and status=? group by price,side order by price", side, art, "limit", "open").Scan(&orders).Error
	if err == gorm.ErrRecordNotFound {
		return nil, err
	}
	return orders, err
}
