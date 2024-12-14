/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package mysql

import (
	"fmt"
	"strings"
	"time"

	"github.com/jhblack-olya/trade-engine/models"
)

func (s *Store) AddBills(bills []*models.Bill) error {
	if len(bills) == 0 {
		return nil
	}
	var valueStrings []string
	for _, bill := range bills {
		valueString := fmt.Sprintf("(NOW(),%v, '%v', %v, %v, '%v', %v, '%v')",
			bill.UserId, bill.Currency, bill.Available, bill.Hold, bill.Type, bill.Settled, bill.Notes)
		valueStrings = append(valueStrings, valueString)
	}
	sql := fmt.Sprintf("INSERT INTO g_bill (created_at, user_id,currency,available,hold, type,settled,notes) VALUES %s", strings.Join(valueStrings, ","))
	return s.db.Exec(sql).Error
}

func (s *Store) GetUnsettledBillsByUserId(userId int64, currency string) ([]*models.Bill, error) {
	db := s.db.Where("settled=?", 0).Where("user_id=?", userId).
		Where("currency =?", currency).Order("id ASC").Limit(100)
	var bills []*models.Bill
	err := db.Find(&bills).Error
	return bills, err
}

func (s *Store) UpdateBill(bill *models.Bill) error {
	bill.UpdatedAt = time.Now()
	return s.db.Save(bill).Error
}

func (s *Store) GetUnsettledBills() ([]*models.Bill, error) {
	db := s.db.Where("settled =?", 0).Order("id ASC").Limit(100)

	var bills []*models.Bill
	err := db.Find(&bills).Error
	return bills, err
}
